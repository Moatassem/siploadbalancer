package sip

import (
	"cmp"
	"log"
	"net"
	. "siploadbalancer/global"
	"slices"
	"sync"
	"time"
)

var LoadBalancer = newLoadBalancer()

type (
	LoadBalancingNode struct {
		nodeIdx      int
		SIPNodes     []*SIPNode
		Distribution Distribution
		CallCache    map[string]*CallCache
		mu           sync.RWMutex
	}

	SIPNode struct {
		IPv4        string
		Port        int
		Name        string
		Description string

		UdpAddr *net.UDPAddr
		Hits    int
		LastHit time.Time
		Weight  int
		IsAlive bool

		mu sync.RWMutex
	}

	Status       string
	Distribution string

	CallCache struct {
		SIPNode    *SIPNode
		OtherAddr  *net.UDPAddr
		IsOutbound bool
		CallID     string
		FromTag    string
		ViaBranch  string
		CallStatus Status
		Messages   []string

		timeoutTmr *time.Timer
		clearTmr   *time.Timer
		mu         sync.RWMutex
	}
)

// add Via on top from my own IPv4:Port UDP .. keep contact header
// parse the sip message until you get Call-ID, Via and From headers .. including start-line
// remove my Via in responses (initial ones)

const (
	StatusProgressing Status = "Progressing" // received OPTIONS, INVITE, MESSAGE, REGISTER
	StatusRejected    Status = "Rejected"    // received 3xx-6xx
	StatusAnswered    Status = "Answered"    // received 2xx
	StatusCancelled   Status = "Cancelled"   // received CANCEL
	StatusTimedout    Status = "Timedout"    // received no responses in time

	DistribRoundRobin Distribution = "RoundRobin"
	DistribRandom     Distribution = "Random"
	DistribLeastHit   Distribution = "LeastHit"
	DistribMostIdle   Distribution = "MostIdle"

	timeoutTimerDuration = 32 * time.Second
	clearTimerDuration   = 10 * time.Second
)

func newLoadBalancer() *LoadBalancingNode {
	return &LoadBalancingNode{
		Distribution: DistribRoundRobin,
		CallCache:    make(map[string]*CallCache),
	}
}

func (lb *LoadBalancingNode) GetNode() *SIPNode {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	switch lb.Distribution {
	case DistribRoundRobin:
		lb.nodeIdx++
		if lb.nodeIdx >= len(lb.SIPNodes) {
			lb.nodeIdx = 0
		}
		return lb.SIPNodes[lb.nodeIdx]
	case DistribLeastHit:
		slices.SortFunc(lb.SIPNodes, func(a, b *SIPNode) int { return cmp.Compare(a.Hits, b.Hits) })
		return lb.SIPNodes[0]
	case DistribMostIdle:
		slices.SortFunc(lb.SIPNodes, func(a, b *SIPNode) int {
			if a.LastHit.Before(b.LastHit) {
				return -1
			}
			if a.LastHit.After(b.LastHit) {
				return 1
			}
			return 0
		})
		return lb.SIPNodes[0]
	default: // DistribRandom
		idx := RandomNum(len(lb.SIPNodes))
		return lb.SIPNodes[idx]
	}

}

func (lb *LoadBalancingNode) DeleteCallCache(callID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	delete(lb.CallCache, callID)
}

func (lb *LoadBalancingNode) AddOrGetCallCache(sipmsg *SipMessage, msgAddr *net.UDPAddr) (*CallCache, *net.UDPAddr) {
	lb.mu.RLock()
	cc, ok := lb.CallCache[sipmsg.CallID]
	lb.mu.RUnlock()

	if ok {
		cc.mu.Lock()
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
				cc.clearTmr = getClearTimer(cc.CallID)
			case IsNegative(stsCode):
				cc.CallStatus = StatusRejected
				cc.clearTmr = getClearTimer(cc.CallID)
			}
		} else {
			sipmsg.Headers.AddTopVia()
		}
		cc.mu.Unlock()

		if AreUAddrsEqual(cc.OtherAddr, msgAddr) {
			return cc, cc.SIPNode.UdpAddr
		}
		return cc, cc.OtherAddr
	}

	if sipmsg.IsResponse() || !sipmsg.GetMethod().IsDialogueCreating() {
		log.Printf("Message [%s] cannot initiate a dialogue - Dropping", sipmsg.String())
		return nil, nil
	}

	var rmtAddr *net.UDPAddr
	var err error

	rmtAddr, err = BuildUDPAddr(sipmsg.StartLine.Host, sipmsg.StartLine.Port)
	if err != nil {
		log.Printf("Message [%s] contains not reachable host - Error [%s] - Dropping", sipmsg.String(), err)
	}

	var azrAddr *net.UDPAddr
	var isout bool

	msgAddrstrg := msgAddr.String()
	sn := Find(lb.SIPNodes, func(x *SIPNode) bool { return x.UdpAddr.String() == msgAddrstrg })
	if sn == nil { // inbound from Access to Core
		sn = lb.GetNode()
		sn.AddHit()
		azrAddr = msgAddr
	} else { // outbound from Core to Access
		azrAddr = rmtAddr
		isout = true
	}

	cc = &CallCache{
		SIPNode:    sn,
		OtherAddr:  azrAddr,
		IsOutbound: isout,
		CallID:     sipmsg.CallID,
		FromTag:    sipmsg.FromTag,
		ViaBranch:  sipmsg.ViaBranch,
		CallStatus: StatusProgressing,
		Messages:   []string{sipmsg.String()},
	}

	sipmsg.Headers.AddTopVia()

	cc.timeoutTmr = time.NewTimer(timeoutTimerDuration)
	go cc.timeoutHandler()

	lb.mu.Lock()
	lb.CallCache[sipmsg.CallID] = cc
	lb.mu.Unlock()

	return cc, rmtAddr
}

func (cc *CallCache) timeoutHandler() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.CallStatus = StatusTimedout
	cc.clearTmr = getClearTimer(cc.CallID)
}

func getClearTimer(callID string) *time.Timer {
	return time.AfterFunc(clearTimerDuration, func() { LoadBalancer.DeleteCallCache(callID) })
}

func (sn *SIPNode) AddHit() {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	sn.Hits++
	sn.LastHit = time.Now().UTC()
}
