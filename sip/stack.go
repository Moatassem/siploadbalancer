package sip

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"siploadbalancer/cl"
	. "siploadbalancer/global"
	"siploadbalancer/redis"
	"strconv"
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
	msgmap := NewSIPHeaders()

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
	} else {
		if RMatch(msglines[lnIdx], ResponseStartLinePattern, &matches) {
			msgType = RESPONSE
			code := Str2int[int](matches[2])
			if code < 100 || code > 699 {
				return nil, nil, errors.New("invalid code for Response message")
			}
			startLine.StatusCode = code
			startLine.ReasonPhrase = matches[3]
			startLine.UriParameters = ParseParameters(matches[4])
			startLine.StartLine = msglines[0]
		} else {
			sipmsg.MsgType = INVALID
			return sipmsg, nil, errors.New("invalid message")
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
			case P_Asserted_Identity.LowerCaseString():
				sipmsg.PAIHeaders = append(sipmsg.PAIHeaders, value)
			case Diversion.LowerCaseString():
				sipmsg.DivHeaders = append(sipmsg.DivHeaders, value)
			case Call_ID.LowerCaseString():
				sipmsg.CallID = value
			case Max_Forwards.LowerCaseString():
				mx, err := strconv.Atoi(value)
				if err != nil {
					log.Println(fmt.Sprintf("Invalid Max-Forwards header - %v", err.Error()))
				} else if mx < 0 || mx > 255 {
					log.Println("Invalid Max-Forwards header - Too little/big")
				} else {
					sipmsg.MaxFwds = mx
				}
			case Contact.LowerCaseString():
				rc := DicFieldRegEx[URIFull].FindStringSubmatch(value)
				if rc != nil {
					sipmsg.RCURI = rc[1]
				}
			case Record_Route.LowerCaseString():
				rc := DicFieldRegEx[URIFull].FindStringSubmatch(value)
				if rc != nil {
					sipmsg.RRURI = rc[1]
				}
			case CSeq.LowerCaseString():
				cseq := DicFieldRegEx[CSeqHeader].FindStringSubmatch(value)
				if cseq == nil {
					log.Println("Invalid CSeq header")
					return nil, nil, errors.New("invalid CSeq header")
				}
				sipmsg.CSeqNum = Str2uint[uint32](cseq[1])
				sipmsg.CSeqMethod = MethodFromName(cseq[2])
				if startLine.StatusCode == 0 {
					r1 := startLine.Method.String()
					r2 := ASCIIToUpper(cseq[2])
					if r1 != r2 {
						log.Println(fmt.Sprintf("Invalid Request Method: %v vs CSeq Method: %v", r1, r2))
						return nil, nil, errors.New("invalid CSeq header")
					}
				}
			case Via.LowerCaseString():
				via := DicFieldRegEx[ViaBranchPattern].FindStringSubmatch(value)
				if via != nil {
					sipmsg.ViaBranch = via[1]
				}
			}
			msgmap.Add(headerLC, value)
		}
	}

	_bodyStartIdx = _dblCrLfIdx + 4 //CrLf x 2

	//automatic deducing of content-length
	cntntLength = len(payload) - _bodyStartIdx
	sipmsg.ContentLength = uint16(cntntLength)

	if ok, values := msgmap.ValuesHeader(Content_Length); ok {
		cntntLength, _ = strconv.Atoi(values[0])
	} else {
		if ok, _ := msgmap.ValuesHeader(Content_Type); ok {
			msgmap.AddHeader(Content_Length, strconv.Itoa(cntntLength))
		} else {
			msgmap.AddHeader(Content_Length, "0")
		}
	}
	sipmsg.Headers = msgmap

	//body parsing
	if cntntLength == 0 {
		payload = payload[_bodyStartIdx:]
		sipmsg.Body = &MessageBody{}
		return sipmsg, payload, nil
	}
	if len(payload) < _bodyStartIdx+cntntLength {
		log.Println("bad content-length or fragmented pdu")
		return nil, nil, errors.New("bad content-length or fragmented pdu")
	}
	// ---------------------------------
	var MB = MessageBody{PartsBytes: map[BodyType]ContentPart{}}

	var cntntTypeSections map[string]string
	ok, v := msgmap.ValuesHeader(Content_Type)
	if !ok {
		return nil, nil, errors.New("bad message - invalid body")
	}
	cntntTypeSections = CleanAndSplitHeader(v[0])
	if cntntTypeSections == nil {
		// LogWarning("Content-Type header is missing while Content-Length is non-zero - Message skipped", LogTitle.SIPStack)
		return nil, nil, errors.New("bad message - invalid body")
	}

	cntntType := ASCIIToLower(cntntTypeSections["!headerValue"])

	if !strings.Contains(cntntType, "multipart") {
		MB.PartsBytes[GetBodyType(cntntType)] = ContentPart{
			Bytes: payload[_bodyStartIdx : _bodyStartIdx+cntntLength], //do not trim last two bytes i.e. \r\n = -2
		}
		payload = payload[_bodyStartIdx+cntntLength:]
	} else {
		payload = payload[_bodyStartIdx:]
		boundary := cntntTypeSections["boundary"]
		markBoundary := "--" + boundary
		endBoundary := "--" + boundary + "--\r\n"
		idxEnd := 0
		for {
			idx = GetNextIndex(payload, markBoundary)
			if idx == -1 || string(payload) == endBoundary {
				break
			}
			payload = payload[idx+len(markBoundary)+2:]
			idx = GetNextIndex(payload, "\r\n\r\n")
			idxEnd = GetNextIndex(payload, markBoundary)
			msglines = strings.Split(string(payload[:idx]), "\r\n")
			bt := None
			partHeaders := NewSH()
			for _, ln := range msglines {
				matches = DicFieldRegEx[FullHeader].FindStringSubmatch(ln)
				if matches != nil {
					h := ASCIIToLower(matches[1])
					partHeaders.Add(h, matches[2])
					if h == "content-type" {
						cntntType = matches[2]
						bt = GetBodyType(cntntType)
					}
				}
			}
			if bt != None {
				MB.PartsBytes[bt] = ContentPart{
					Headers: partHeaders,
					Bytes:   payload[:idxEnd], //do not trim last two bytes i.e. \r\n = -2
				}
			}
			payload = payload[idxEnd:]
		}
		if len(MB.PartsBytes) == 1 {
			log.Println("Missing body part Content-Type - skipped")
		}
	}
	sipmsg.Body = &MB

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
