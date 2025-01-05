package sip

import (
	"errors"
	"fmt"
	"net"
	"os"
	"siploadbalancer/cl"
	. "siploadbalancer/global"
	"siploadbalancer/redis"
	"strings"
)

func udpLoop(conn *net.UDPConn) {
	defer WtGrp.Done()
	defer func() {
		if r := recover(); r != nil {
			LogCallStack(r)
			udpLoop(conn)
		}
	}()
	for {
		buf := BufferPool.Get().(*[]byte)
		n, addr, err := conn.ReadFromUDP(*buf)
		if err != nil {
			fmt.Println(err)
			continue
		}
		go func() {
			pdu := (*buf)[:n]
			for {
				if len(pdu) == 0 {
					break
				}
				msg, pdutmp, err := processPDU(pdu)
				if err != nil {
					fmt.Println("Bad PDU -", err)
					fmt.Println(string(pdu))
					break
				} else if msg == nil {
					break
				}
				ss, newSesType := sessionGetter(msg)
				if ss != nil {
					ss.RemoteUDP = addr
					ss.UDPListenser = conn
				}
				sipStack(msg, ss, newSesType)
				pdu = pdutmp
			}
			BufferPool.Put(buf)
		}()
	}
}

func processPDU(payload []byte) (*SipMessage, []byte, error) {
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
	var startLine = SipStartLine{}

	sipmsg := NewSipMessage()
	msgmap := NewSIPHeaders(false)

	idx := 0
	var _dblCrLfIdx, _bodyStartIdx, lnIdx, cntntLength int

	_dblCrLfIdx = GetNextIndex(payload, "\r\n\r\n")

	if _dblCrLfIdx == -1 {
		//empty sip message
		return nil, nil, nil
	}

	msglines := strings.Split(string(payload[:_dblCrLfIdx]), "\r\n")

	lnIdx = 0
	var matches []string
	//start line parsing
	if RMatch(msglines[lnIdx], RequestStartLinePattern, &matches) {
		msgType = REQUEST
		startLine.StatusCode = 0
		startLine.Method = MethodFromName(ASCIIToUpper(matches[1]))
		if startLine.Method == UNKNOWN {
			return sipmsg, nil, errors.New("invalid method for Request message")
		}
		startLine.Ruri = matches[2]
		startLine.StartLine = msglines[0]
		if startLine.Method == INVITE && RMatch(startLine.Ruri, INVITERURI, &matches) {
			startLine.UriScheme = ASCIIToLower(matches[1])
			startLine.UserPart = matches[2]
			startLine.HostPart = matches[4]
			startLine.UserParameters = ParseParameters(matches[3])
			startLine.UriParameters = ParseParameters(matches[5])
		}
	}
	sipmsg.MsgType = msgType
	sipmsg.StartLine = &startLine

	lnIdx += 1

	//headers parsing
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
				sipmsg.FromHeader = value
			case To.LowerCaseString():
				tag := DicFieldRegEx[Tag].FindStringSubmatch(value)
				if tag != nil {
					sipmsg.ToTag = tag[1]
					if tag[1] != "" && startLine.Method == INVITE {
						startLine.Method = ReINVITE
					}
				}
				sipmsg.ToHeader = value
			case Call_ID.LowerCaseString():
				sipmsg.CallID = value
			case Via.LowerCaseString():
				via := DicFieldRegEx[ViaBranchPattern].FindStringSubmatch(value)
				if via != nil {
					sipmsg.ViaBranch = via[1]
					if !strings.HasPrefix(via[1], MagicCookie) {
						LogWarning(LTSIPStack, fmt.Sprintf("Received message [%v] having non-RFC3261 Via branch [%v]", startLine.Method.String(), via[1]))
					}
					if len(via[1]) <= len(MagicCookie) {
						LogWarning(LTSIPStack, fmt.Sprintf("Received message [%v] having too short Via branch [%v]", startLine.Method.String(), via[1]))
					}
				}
			}
			msgmap.Add(headerLC, value)
		}
	}

	return sipmsg, payload, nil
}

func startListening(ip net.IP, prt int) (*net.UDPConn, error) {
	socket := net.UDPAddr{}
	socket.IP = ip
	socket.Port = prt
	return net.ListenUDP("udp", &socket)
}

func StartServer(redisskt string, ipv4 string) (*net.UDPConn, net.IP) {
	serverIP := net.ParseIP(ipv4)

	fmt.Print("Attempting to listen on SIP...")
	serverUDPListener, err := startListening(serverIP, SipUdpPort)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	fmt.Println("Success: UDP", serverUDPListener.LocalAddr().String())

	fmt.Print("Checking Caching Server...")
	ripv4skt, err := redis.SetupCheckRedis(redisskt, "", 0, 15) //TODO: add redis password, db and expiryMin
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	fmt.Printf("Ready! [%s]\n", ripv4skt)
	// fmt.Println("Skipped!")

	fmt.Print("Setting Rate Limiter...")
	CallLimiter = cl.NewCallLimiter(RateLimit, Prometrics)
	fmt.Printf("OK (%d)\n", RateLimit)

	return serverUDPListener, serverIP
}
