/*
# Software Name : Session Router (SR)
# SPDX-FileCopyrightText: Copyright (c) Orange Business - OINIS/Services/NSF
# SPDX-License-Identifier: Apache-2.0
#
# This software is distributed under the Apache-2.0
# See the "LICENSES" directory for more details.
#
# Authors:
# - Moatassem Talaat <moatassem.talaat@orange.com>

---
*/

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

func NewRequestMessage(md Method, up string) *SipMessage {
	sipmsg := &SipMessage{
		MsgType: REQUEST,
		StartLine: SipStartLine{
			Method: md,
		},
	}
	return sipmsg
}

func NewResponseMessage(sc int, rp string) *SipMessage {
	if sc < 100 || sc > 699 {
		sc = 400
	}
	sipmsg := &SipMessage{
		MsgType: RESPONSE,
		StartLine: SipStartLine{
			StatusCode:   sc,
			ReasonPhrase: rp,
		},
	}
	return sipmsg
}

// ==========================================================================

func (sipmsg *SipMessage) String() string {
	if sipmsg.MsgType == REQUEST {
		return sipmsg.StartLine.Method.String()
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
	for _, h := range sipmsg.getHeaderNames() {
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

func (sipmsg *SipMessage) getHeaderNames() []string {
	return sipmsg.Headers.hnames
}
