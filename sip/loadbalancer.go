package sip

import (
	"cmp"
	"fmt"
	"log"
	"net"
	. "siploadbalancer/global"
	"slices"
	"sync"
	"time"
)

var LoadBalancer *LoadBalancingNode

type (
	LoadBalancingNode struct {
		SipNodes     []*SipNode
		Distribution Distribution

		nodeIdx    int
		callsCache map[string]*CallCache
		mu         sync.RWMutex
	}

	SipNode struct {
		UdpAddr     *net.UDPAddr
		Description string
		Cost        int
		Weight      int
		accWeight   int

		Key     string
		Hits    int
		LastHit time.Time
		IsAlive bool

		mu sync.RWMutex
	}

	Status       string
	Distribution string

	CallCache struct {
		SIPNode      *SipNode
		OtherAddr    *net.UDPAddr
		IsOutbound   bool
		CallID       string
		FromTag      string
		OwnViaBranch string
		CallStatus   Status
		Messages     []string
		IsProbing    bool

		timeoutTmr *time.Timer
		clearTmr   *time.Timer
		mu         sync.RWMutex
	}
)

const (
	StatusProgressing Status = "Progressing" // received Dialogue-creating methods
	StatusRejected    Status = "Rejected"    // received 3xx-6xx
	StatusAnswered    Status = "Answered"    // received 2xx
	StatusCancelled   Status = "Cancelled"   // received CANCEL
	StatusTimedout    Status = "Timedout"    // received no responses in time

	DistribRoundRobin Distribution = "RoundRobin"
	DistribLeastHit   Distribution = "LeastHit"
	DistribLeastCost  Distribution = "LeastCost"
	DistribMostIdle   Distribution = "MostIdle"
	DistribWeighted   Distribution = "Weighted"
	DistribRandom     Distribution = "Random"

	timeoutTimerDuration = 32 * time.Second
	clearTimerDuration   = 10 * time.Second
)

func NewLoadBalancer(lbm string, sipnodes []*SipNode) *LoadBalancingNode {
	return &LoadBalancingNode{
		SipNodes:     sipnodes,
		Distribution: DistribRoundRobin,
		callsCache:   make(map[string]*CallCache),
	}
}

func createClearTimer(callID string) *time.Timer {
	return time.AfterFunc(clearTimerDuration, func() { LoadBalancer.DeleteCallCache(callID) })
}

func (lb *LoadBalancingNode) CallsCacheCount() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	return len(lb.callsCache)
}

func (lb *LoadBalancingNode) GetNode() *SipNode {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	switch lb.Distribution {
	case DistribRoundRobin:
		nd := lb.SipNodes[lb.nodeIdx]
		lb.nodeIdx++
		if lb.nodeIdx >= len(lb.SipNodes) {
			lb.nodeIdx = 0
		}
		return nd
	case DistribLeastHit:
		slices.SortFunc(lb.SipNodes, func(a, b *SipNode) int { return cmp.Compare(a.Hits, b.Hits) })
		return lb.SipNodes[0]
	case DistribLeastCost:
		slices.SortFunc(lb.SipNodes, func(a, b *SipNode) int { return cmp.Compare(a.Cost, b.Cost) })
		return lb.SipNodes[0]
	case DistribMostIdle:
		slices.SortFunc(lb.SipNodes, func(a, b *SipNode) int {
			if a.LastHit.Before(b.LastHit) {
				return -1
			}
			if a.LastHit.After(b.LastHit) {
				return 1
			}
			return 0
		})
		return lb.SipNodes[0]
	case DistribWeighted:
		panic("not implemented yet")
	default: // DistribRandom
		return lb.SipNodes[RandomNum(len(lb.SipNodes))]
	}
}

func (lb *LoadBalancingNode) DeleteCallCache(callID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	delete(lb.callsCache, callID)
}

func (lb *LoadBalancingNode) ProbeSipNodes() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, sn := range lb.SipNodes {
		callid := GetCallID()
		viaBranch := GetViaBranch()
		frmTag := GetTagOrKey()
		localstr := ServerConnection.LocalAddr().String()
		remotestr := sn.UdpAddr.String()

		hdrs := NewSipHeaders()
		hdrs.Add(Via, buildViaHeader(viaBranch))
		hdrs.Add(From, fmt.Sprintf("<sip:ping@%s>;tag=%s", localstr, frmTag))
		hdrs.Add(To, fmt.Sprintf("<sip:ping@%s>", remotestr))
		hdrs.Add(Call_ID, callid)
		hdrs.Add(CSeq, fmt.Sprintf("911 %s", OPTIONS))
		hdrs.Add(Contact, fmt.Sprintf("<sip:%s>", localstr))
		hdrs.Add(Max_Forwards, "70")
		hdrs.Add(User_Agent, BUE)
		hdrs.Add(Content_Length, "0")

		probemsg := &SipMessage{
			MsgType: REQUEST,
			StartLine: SipStartLine{
				Method: OPTIONS,
				RUri:   fmt.Sprintf("sip:%s", remotestr),
			},
			Headers: hdrs,
		}

		cc := &CallCache{
			SIPNode:      sn,
			CallID:       callid,
			FromTag:      frmTag,
			OwnViaBranch: viaBranch,
			CallStatus:   StatusProgressing,
			IsProbing:    true,
		}
		cc.StartTimeoutTimer()

		lb.callsCache[callid] = cc

		_, err := ServerConnection.WriteTo(probemsg.Bytes(), sn.UdpAddr)
		if err != nil {
			log.Println("Failed to send probing message - error:", err)
		}
	}
}

func (lb *LoadBalancingNode) AddOrGetCallCache(sipmsg *SipMessage, srcAddr *net.UDPAddr) (*CallCache, *net.UDPAddr) {
	lb.mu.RLock()
	cc, ok := lb.callsCache[sipmsg.CallID]
	lb.mu.RUnlock()

	if ok {
		cc.mu.Lock()

		if cc.IsProbing {
			defer cc.mu.Unlock()

			if cc.timeoutTmr.Stop() {
				cc.SIPNode.SetAlive(true)
				LoadBalancer.DeleteCallCache(cc.CallID)
			}

			return nil, nil
		}

		cc.Messages = append(cc.Messages, sipmsg.String())
		if sipmsg.IsResponse() {
			cc.timeoutTmr.Stop()
			sipmsg.Headers.DropTopVia()

			stsCode := sipmsg.StartLine.StatusCode
			switch {
			case IsProvisional(stsCode):
				cc.CallStatus = StatusProgressing
			case IsPositive(stsCode):
				cc.CallStatus = StatusAnswered
				cc.clearTmr = createClearTimer(cc.CallID)
			case IsNegative(stsCode):
				cc.CallStatus = StatusRejected
				cc.clearTmr = createClearTimer(cc.CallID)
			}
		} else {
			sipmsg.Headers.AddTopVia(cc.OwnViaBranch)
		}
		cc.mu.Unlock()

		if AreUAddrsEqual(cc.OtherAddr, srcAddr) {
			return cc, cc.SIPNode.UdpAddr
		}

		return cc, cc.OtherAddr
	}

	if sipmsg.IsResponse() || !sipmsg.GetMethod().IsDialogueCreating() {
		log.Printf("Message [%s] cannot initiate a dialogue - Dropping", sipmsg.String())
		return nil, nil
	}

	var rmtAddr, azrAddr *net.UDPAddr
	var isout bool

	sn := Find(lb.SipNodes, func(x *SipNode) bool { return AreUAddrsEqual(x.UdpAddr, srcAddr) })
	if sn == nil { // inbound from Access to Core
		sn = lb.GetNode()
		sn.AddHit()
		azrAddr = srcAddr
		rmtAddr = sn.UdpAddr
	} else { // outbound from Core to Access
		msgTargetAddr, err := BuildSipUdpSocket(sipmsg.StartLine.Host, sipmsg.StartLine.Port)
		if err != nil {
			log.Printf("Message [%s] contains not reachable host - Error [%s] - Dropping", sipmsg.String(), err)
			return nil, nil
		}
		azrAddr = msgTargetAddr
		rmtAddr = msgTargetAddr
		isout = true
	}

	cc = &CallCache{
		SIPNode:      sn,
		OtherAddr:    azrAddr,
		IsOutbound:   isout,
		CallID:       sipmsg.CallID,
		FromTag:      sipmsg.FromTag,
		OwnViaBranch: GetViaBranch(),
		CallStatus:   StatusProgressing,
		Messages:     []string{sipmsg.String()},
	}
	cc.StartTimeoutTimer()

	lb.mu.Lock()
	lb.callsCache[sipmsg.CallID] = cc
	lb.mu.Unlock()

	sipmsg.Headers.AddTopVia(cc.OwnViaBranch)

	return cc, rmtAddr
}

func (cc *CallCache) timeoutHandler() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.clearTmr = createClearTimer(cc.CallID)

	if cc.IsProbing {
		cc.SIPNode.SetAlive(false)
		return
	}

	cc.CallStatus = StatusTimedout
}

func (cc *CallCache) StartTimeoutTimer() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.timeoutTmr = time.AfterFunc(timeoutTimerDuration, func() { cc.timeoutHandler() })
}

func (sn *SipNode) AddHit() {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	sn.Hits++ // TODO: find a way to rest this count!
	sn.LastHit = time.Now().UTC()
}

func (sn *SipNode) SetAlive(flag bool) {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	sn.IsAlive = flag

	// fmt.Printf("SipNode: %s - IsAlive: %v\n", sn.UdpAddr.String(), sn.IsAlive)
}
