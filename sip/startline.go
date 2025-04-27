/*
# Software Name : Newkah-SIP-Layer
# SPDX-FileCopyrightText: Copyright (c) 2025 - Orange Business - OINIS/Services/NSF

# Authors:
# - Moatassem Talaat <moatassem.talaat@orange.com>

---
*/

package sip

import (
	"fmt"
	. "siploadbalancer/global"
)

type SipStartLine struct {
	Method Method
	Host   string
	Port   string

	RUri string

	StatusCode   int
	ReasonPhrase string
}

func (ssl *SipStartLine) BuildStartLine(mt MessageType) string {
	if mt == REQUEST {
		return fmt.Sprintf("%s %s %s\r\n", ssl.Method, ssl.RUri, SipVersion)
	}
	return fmt.Sprintf("%s %d %s\r\n", SipVersion, ssl.StatusCode, ssl.ReasonPhrase)
}
