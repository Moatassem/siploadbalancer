package sip

import (
	"bytes"
	"fmt"
	. "siploadbalancer/global"
)

type SipMessage struct {
	MsgType   MessageType
	StartLine SipStartLine
	Headers   *SipHeaders
	Body      []byte

	CallID    string
	FromTag   string
	ToTag     string
	ViaBranch string
}

func BuildResponseMessage(rqstmsg *SipMessage, sc int, rp string) *SipMessage {
	hdrs := NewSipHeaders()
	hdrs.Add(Via, rqstmsg.Headers.GetHeaderValues(ViaHeader)...)
	hdrs.Add(From, rqstmsg.Headers.GetHeaderValues(From)...)
	hdrs.Add(To, rqstmsg.Headers.GetHeaderValues(To)...)
	hdrs.Add(Call_ID, rqstmsg.CallID)
	hdrs.Add(CSeq, rqstmsg.Headers.GetHeaderValues(CSeq)...)
	hdrs.Add(Server, BUE)
	hdrs.Add(Content_Length, "0")

	rspnsmsg := &SipMessage{
		MsgType: RESPONSE,
		StartLine: SipStartLine{
			StatusCode:   sc,
			ReasonPhrase: rp,
		},
		Headers: hdrs,
	}

	return rspnsmsg
}

// ==========================================================================

func (sipmsg *SipMessage) String() string {
	if sipmsg.MsgType == REQUEST {
		return string(sipmsg.StartLine.Method)
	}

	return Int2Str(sipmsg.StartLine.StatusCode)
}

func (sipmsg *SipMessage) Bytes() []byte {
	var bb bytes.Buffer

	// startline
	if sipmsg.IsRequest() {
		sl := sipmsg.StartLine
		bb.WriteString(sl.BuildStartLine(REQUEST))
	} else {
		sl := sipmsg.StartLine
		bb.WriteString(sl.BuildStartLine(RESPONSE))
	}

	// headers - build and write
	for _, h := range sipmsg.Headers.hnames {
		_, values := sipmsg.Headers.Values(h)
		for _, hv := range values {
			if hv != "" {
				bb.WriteString(fmt.Sprintf("%s: %s\r\n", h, hv))
			}
		}
	}

	// write separator
	bb.WriteString("\r\n")

	// write body bytes
	bb.Write(sipmsg.Body)

	return bb.Bytes()
}

// ===========================================================================

func (sipmsg *SipMessage) IsOutOfDialgoue() bool {
	return sipmsg.ToTag == ""
}

func (sipmsg *SipMessage) IsResponse() bool {
	return sipmsg.MsgType == RESPONSE
}

func (sipmsg *SipMessage) IsRequest() bool {
	return sipmsg.MsgType == REQUEST
}

func (sipmsg *SipMessage) GetMethod() Method {
	return sipmsg.StartLine.Method
}

func (sipmsg *SipMessage) GetStatusCode() int {
	return sipmsg.StartLine.StatusCode
}

// ======================================
