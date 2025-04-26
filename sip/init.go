package sip

import (
	"errors"
	"fmt"
	"log"
	"net"
	"runtime"
	. "siploadbalancer/global"
	"strings"
	"sync"
)

var (
	ServerConnection *net.UDPConn
	WorkerCount      = runtime.NumCPU()
	QueueSize        = 3500
	packetQueue      = make(chan Packet, QueueSize)
	BufferPool       = newSyncPool()
)

const (
	BufferSize int = 4096
)

func newSyncPool() *sync.Pool {
	return &sync.Pool{
		New: func() any {
			lst := make([]byte, BufferSize)
			return &lst
		},
	}
}

type Packet struct {
	sourceAddr *net.UDPAddr
	buffer     *[]byte
	bytesCount int
}

func startWorkers() {
	WtGrp.Add(WorkerCount)
	for range WorkerCount {
		go worker(packetQueue)
	}
}

func udpLoopWorkers() {
	WtGrp.Add(1)
	defer func() {
		WtGrp.Done()
		if r := recover(); r != nil {
			LogCallStack(r)
			udpLoopWorkers()
		}
	}()
	go func() {
		for {
			buf := BufferPool.Get().(*[]byte)
			n, addr, err := ServerConnection.ReadFromUDP(*buf)
			if err != nil {
				fmt.Println(err)
				continue
			}
			packetQueue <- Packet{sourceAddr: addr, buffer: buf, bytesCount: n}
		}
	}()
}

func worker(queue <-chan Packet) {
	defer WtGrp.Done()
	for packet := range queue {
		processPacket(packet)
	}
}

func processPacket(packet Packet) {
	pdu := (*packet.buffer)[:packet.bytesCount]
	for {
		if len(pdu) == 0 {
			break
		}
		msg, pdutmp, err := parsePDU(pdu)
		if err != nil {
			fmt.Println("Bad PDU -", err)
			fmt.Println(string(pdu))
			break
		} else if msg == nil {
			break
		}
		callHandler(msg, packet.sourceAddr)
		pdu = pdutmp
	}
	BufferPool.Put(packet.buffer)
}

func parsePDU(payload []byte) (*SipMessage, []byte, error) {
	defer func() {
		if r := recover(); r != nil {
			// check if pdu is rqst >> send 400 with Warning header indicating what was wrong or unable to parse
			// or discard rqst if totally wrong
			// if pdu is rsps >> discard
			// in any case, log this pdu by saving its hex stream and why it was wrong
			LogCallStack(r)
		}
	}()

	var msgType MessageType
	var startLine SipStartLine

	sipmsg := new(SipMessage)
	msgmap := NewSHsPointer()

	var _dblCrLfIdx, _bodyStartIdx, lnIdx, cntntLength, cntntLengthComputed int

	_dblCrLfIdxInt := GetNextIndex(payload, "\r\n\r\n")

	if _dblCrLfIdxInt == -1 {
		// empty sip message
		return nil, nil, nil
	}

	_dblCrLfIdx = _dblCrLfIdxInt

	msglines := strings.Split(string(payload[:_dblCrLfIdx]), "\r\n")

	lnIdx = 0
	var matches []string
	// start line parsing
	if RMatch(msglines[lnIdx], RequestStartLinePattern, &matches) {
		msgType = REQUEST
		startLine.StatusCode = 0
		startLine.Method = MethodFromName(ASCIIToUpper(matches[1]))
		if startLine.Method == UNKNOWN {
			return sipmsg, nil, errors.New("invalid method for Request message")
		}
		startLine.RUri = matches[2]
		if startLine.Method == INVITE && RMatch(startLine.RUri, INVITERURI, &matches) {
			startLine.Host = matches[5]
			startLine.Port = matches[6]
		}
	} else {
		if RMatch(msglines[lnIdx], ResponseStartLinePattern, &matches) {
			msgType = RESPONSE
			code := Str2Int[int](matches[2])
			if code < 100 || code > 699 {
				return nil, nil, errors.New("invalid code for Response message")
			}
			startLine.StatusCode = code
			startLine.ReasonPhrase = matches[3]
		} else {
			sipmsg.MsgType = INVALID
			return sipmsg, nil, errors.New("invalid message")
		}
	}
	sipmsg.MsgType = msgType
	sipmsg.StartLine = startLine

	lnIdx += 1

	// headers parsing

	for i := lnIdx; i < len(msglines) && msglines[i] != ""; i++ {
		matches := DicFieldRegEx[FullHeader].FindStringSubmatch(msglines[i])
		if matches != nil {
			headerLC := ASCIIToLower(matches[1])
			value := matches[2]
			switch headerLC {
			case From.LowerCaseString():
				tag := DicFieldRegEx[Tag].FindStringSubmatch(value)
				if tag != nil {
					sipmsg.FromTag = tag[1]
				}
			case To.LowerCaseString():
				tag := DicFieldRegEx[Tag].FindStringSubmatch(value)
				if tag != nil && tag[1] != "" {
					sipmsg.ToTag = tag[1]
					if startLine.Method == INVITE {
						startLine.Method = ReINVITE
					}
				}
			case Content_Length.LowerCaseString():
				cntntLength = Str2Int[int](value)
			case Call_ID.LowerCaseString():
				sipmsg.CallID = value
			case Via.LowerCaseString():
				via := DicFieldRegEx[ViaBranchPattern].FindStringSubmatch(value)
				if via != nil {
					if sipmsg.ViaBranch == "" {
						sipmsg.ViaBranch = via[1]
					}
				}
			}
			msgmap.Add(headerLC, value)
		}
	}

	_bodyStartIdx = _dblCrLfIdx + 4 // CrLf x 2

	// automatic deducing of content-length
	cntntLengthComputed = len(payload) - _bodyStartIdx

	if cntntLengthComputed != cntntLength {
		log.Printf("Discrepancy encountered in Content-Length - computed [%d] vs received [%d]", cntntLengthComputed, cntntLength)
	}

	sipmsg.Headers = msgmap

	// body parsing
	if cntntLength == 0 {
		payload = payload[_bodyStartIdx:]
		return sipmsg, payload, nil
	}

	sipmsg.Body = payload[_bodyStartIdx : _bodyStartIdx+cntntLength]

	payload = payload[_bodyStartIdx+cntntLength:]

	return sipmsg, payload, nil
}

func callHandler(sipmsg *SipMessage, msgAddr *net.UDPAddr) {
	defer func() {
		if r := recover(); r != nil {
			LogCallStack(r)
		}
	}()

	cc, rmtAddr := LoadBalancer.AddOrGetCallCache(sipmsg, msgAddr)
	if cc == nil || rmtAddr == nil {
		return
	}

	ServerConnection.WriteTo(sipmsg.Bytes(), rmtAddr)
}
