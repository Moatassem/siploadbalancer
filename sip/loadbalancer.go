package sip

import (
	"cmp"
	"fmt"
	"log"
	"net"

	. "siploadbalancer/global"
	"slices"
	"strings"
	"sync"
	"time"
)

var LoadBalancer *LoadBalancingNode

type (
	LoadBalancingNode struct {
		SipNodes             []*SipNode   `json:"sipNodes"`
		Distribution         Distribution `json:"distribution"`
		ProbingInterval      int          `json:"probingInterval"`
		TimeoutTimerDuration int          `json:"timeoutTimerDuration"`
		ClearTimerDuration   int          `json:"clearTimerDuration"`

		sipNodesMap map[string]*SipNode `json:"-"`
		SipNodesLB  []string            `json:"sipNodesLB"`
		nodeIdx     int                 `json:"-"`

		callsCache map[string]*CallCache //`json:"-"`
		mu         sync.RWMutex          `json:"-"`
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
		isAlive bool

		mu sync.RWMutex
	}

	Status       string
	Distribution string

	CallCache struct {
		SIPNode      *SipNode
		OtherAddr    *net.UDPAddr
		IsInbound    bool
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

	LongTimeFormat string = "Mon, 02 Jan 2006 15:04:05 GMT"
	JsonTimeFormat string = "2006-01-02T15:04:05Z"

	TimeoutTimerDD = 32 * time.Second // DD = Default Duration
	ClearTimerDD   = 10 * time.Second
)

func NewLoadBalancer(inputData inputData) *LoadBalancingNode {
	sipnodes := make([]*SipNode, 0, len(inputData.Servers))
	sipNodesMap := make(map[string]*SipNode, len(inputData.Servers))
	for _, srvr := range inputData.Servers {
		sipIpv4 := net.ParseIP(srvr.Ipv4)
		if sipIpv4 == nil {
			fmt.Printf("SIP Server IPv4: %s - invalid", srvr.Ipv4)
			continue
		}

		sipprt := srvr.Port
		if sipprt == 0 {
			fmt.Printf("SIP Server Port: %d - invalid", srvr.Port)
			continue
		}

		udpAddr := &net.UDPAddr{IP: sipIpv4, Port: sipprt, Zone: ""}

		if slices.ContainsFunc(sipnodes, func(x *SipNode) bool {
			return AreUAddrsEqual(x.UdpAddr, udpAddr) || strings.EqualFold(x.Description, srvr.Description)
		}) {
			fmt.Println("Duplicate Server record - Skipped")
			continue
		}

		sn := &SipNode{
			Key:         GetTagOrKey(),
			UdpAddr:     udpAddr,
			Description: srvr.Description,
			Cost:        srvr.Cost,
			Weight:      srvr.Weight,
			accWeight:   srvr.Weight,
			isAlive:     false,
			Hits:        0,
		}

		sipnodes = append(sipnodes, sn)
		sipNodesMap[sn.Key] = sn
	}

	lbn := &LoadBalancingNode{
		SipNodes:             sipnodes,
		Distribution:         Distribution(inputData.LoadbalanceMode),
		ProbingInterval:      inputData.ProbingInterval,
		TimeoutTimerDuration: inputData.TimeoutTimerDuration,
		ClearTimerDuration:   inputData.ClearTimerDuration,

		sipNodesMap: sipNodesMap,
		SipNodesLB:  computeSipNodesLB(sipnodes),
		callsCache:  make(map[string]*CallCache),
	}

	return lbn
}

func createClearTimer(callID string) *time.Timer {
	duration := time.Duration(LoadBalancer.ClearTimerDuration) * time.Second
	return time.AfterFunc(duration, func() { LoadBalancer.DeleteCallCache(callID) })
}

func computeSipNodesLB(snlst []*SipNode) []string {
	grandweight := 0
	for _, wh := range snlst {
		grandweight += wh.Weight
	}

	lblst := make([]string, grandweight)
	for gw := range grandweight {
		sn := snlst[0]

		for i := len(snlst) - 1; i > 0; i-- {
			if snlst[i].accWeight >= sn.accWeight {
				sn = snlst[i]
			}
		}

		sn.accWeight -= grandweight

		for _, wh := range snlst {
			weight := wh.Weight
			accWeight := wh.accWeight
			wh.accWeight = weight + accWeight
		}

		lblst[gw] = sn.Key
	}

	return lblst
}

func (lb *LoadBalancingNode) CallsCacheCount() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	return len(lb.callsCache)
}

func (lb *LoadBalancingNode) CallsCache() map[string]*CallCache {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	return lb.callsCache
}

func (lb *LoadBalancingNode) GetNode() *SipNode {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	var outNode *SipNode
	for len(lb.SipNodes) > 0 && outNode == nil {
		switch lb.Distribution {
		case DistribRoundRobin:
			nd := lb.SipNodes[lb.nodeIdx]
			lb.nodeIdx++
			if lb.nodeIdx >= len(lb.SipNodes) {
				lb.nodeIdx = 0
			}
			outNode = nd
		case DistribLeastHit:
			slices.SortFunc(lb.SipNodes, func(a, b *SipNode) int { return cmp.Compare(a.Hits, b.Hits) })
			outNode = lb.SipNodes[0]
		case DistribLeastCost:
			slices.SortFunc(lb.SipNodes, func(a, b *SipNode) int { return cmp.Compare(a.Cost, b.Cost) })
			outNode = lb.SipNodes[0]
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
			outNode = lb.SipNodes[0]
		case DistribWeighted:
			ndKey := lb.SipNodesLB[lb.nodeIdx]
			lb.nodeIdx++
			if lb.nodeIdx >= len(lb.SipNodesLB) {
				lb.nodeIdx = 0
			}
			outNode = lb.sipNodesMap[ndKey]
		default: // DistribRandom
			outNode = lb.SipNodes[RandomNum(len(lb.SipNodes))]
		}

		if outNode.IsDead() {
			if All(lb.SipNodes, func(x *SipNode) bool { return x.IsDead() }) {
				return nil
			}
			outNode = nil
		}
	}

	return outNode
}

func (lb *LoadBalancingNode) DeleteCallCache(callID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	delete(lb.callsCache, callID)
	Prometrics.ConSessions.Dec()
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

		probemsg := BuildOptionsMessage(viaBranch, localstr, remotestr, callid, frmTag)

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
		Prometrics.ConSessions.Inc()

		sendMessage(probemsg, sn.UdpAddr)
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

		var duplicateMsg bool
		cc.Messages, duplicateMsg = AddIfNew(cc.Messages, sipmsg.String())

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
			if cc.IsInbound && duplicateMsg && cc.SIPNode.IsDead() { // if sipnode dies in the middle
				defer cc.mu.Unlock()
				sendMessage(BuildResponseMessage(sipmsg, 503, "Server Unreachable"), srcAddr)
				return nil, nil
			}
			sipmsg.Headers.AddTopVia(cc.OwnViaBranch)
		}

		cc.mu.Unlock()

		if AreUAddrsEqual(cc.OtherAddr, srcAddr) {
			return cc, cc.SIPNode.UdpAddr
		}

		return cc, cc.OtherAddr
	}

	if sipmsg.IsResponse() || !sipmsg.GetMethod().IsDialogueCreating() {
		// log.Printf("Message [%s] cannot initiate a dialogue - Dropping", sipmsg.String())
		return nil, nil
	}

	if !sipmsg.Headers.DecrementMaxForwards() {
		sendMessage(BuildResponseMessage(sipmsg, 483, "Too Many Hops"), srcAddr)
		return nil, nil
	}

	var rmtAddr, azrAddr *net.UDPAddr
	var isingress bool

	sn := Find(lb.SipNodes, func(x *SipNode) bool { return AreUAddrsEqual(x.UdpAddr, srcAddr) })
	if sn == nil { // inbound from Access to Core
		if !CallLimiter.CanAcceptNewSession() {
			sendMessage(BuildResponseMessage(sipmsg, 480, "Call Limiter Exceeded"), srcAddr)
			return nil, nil
		}
		sn = lb.GetNode()
		if sn == nil {
			log.Printf("No more alive servers!")
			sendMessage(BuildResponseMessage(sipmsg, 503, "No Available Servers"), srcAddr)
			return nil, nil
		}
		sn.AddHit()
		azrAddr = srcAddr
		rmtAddr = sn.UdpAddr
		isingress = true
	} else { // outbound from Core to Access
		msgTargetAddr, err := BuildSipUdpSocket(sipmsg.StartLine.Host, sipmsg.StartLine.Port)
		if err != nil {
			log.Printf("Message [%s] contains unreachable host - Error [%s] - Dropping", sipmsg.String(), err)
			return nil, nil
		}
		azrAddr = msgTargetAddr
		rmtAddr = msgTargetAddr
	}

	cc = &CallCache{
		SIPNode:      sn,
		OtherAddr:    azrAddr,
		IsInbound:    isingress,
		CallID:       sipmsg.CallID,
		FromTag:      sipmsg.FromTag,
		OwnViaBranch: GetViaBranch(),
		CallStatus:   StatusProgressing,
		Messages:     []string{sipmsg.String()},
	}
	cc.StartTimeoutTimer()

	lb.mu.Lock()
	lb.callsCache[sipmsg.CallID] = cc
	Prometrics.ConSessions.Inc()
	lb.mu.Unlock()

	sipmsg.Headers.AddTopVia(cc.OwnViaBranch)

	return cc, rmtAddr
}

func (cc *CallCache) timeoutHandler() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.IsProbing {
		cc.SIPNode.SetAlive(false)
		LoadBalancer.DeleteCallCache(cc.CallID)
		return
	}

	cc.clearTmr = createClearTimer(cc.CallID)
	cc.CallStatus = StatusTimedout
}

func (cc *CallCache) StartTimeoutTimer() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	var interval int
	if cc.IsProbing {
		interval = ProbingTimeout
	} else {
		interval = LoadBalancer.TimeoutTimerDuration
	}

	duration := time.Duration(interval) * time.Second
	cc.timeoutTmr = time.AfterFunc(duration, func() { cc.timeoutHandler() })
}

func (sn *SipNode) String() string {
	return fmt.Sprintf("%s (%s)", sn.Description, sn.UdpAddr)
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

	if sn.isAlive != flag {
		stamp := time.Now().UTC().Format(JsonTimeFormat)
		var newsts string
		if flag {
			newsts = "ALIVE"
		} else {
			newsts = "DEAD"
		}
		fmt.Printf("%s became %s on %s\n", sn, newsts, stamp)
	}

	sn.isAlive = flag
}

func (sn *SipNode) IsDead() bool {
	sn.mu.RLock()
	defer sn.mu.RUnlock()

	return !sn.isAlive
}

func sendMessage(sipmsg *SipMessage, rmtUDPAddr *net.UDPAddr) {
	_, err := ServerConnection.WriteTo(sipmsg.Bytes(), rmtUDPAddr)
	if err != nil {
		log.Println("Failed to send response message - error:", err)
	}
}
