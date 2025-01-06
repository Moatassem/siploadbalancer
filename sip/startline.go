/*
# Software Name : Newkah-SIP-Layer
# SPDX-FileCopyrightText: Copyright (c) 2025 - Orange Business - OINIS/Services/NSF

# Authors:
# - Moatassem Talaat <moatassem.talaat@orange.com>

---
*/

package sip

import (
	. "siploadbalancer/global"
)

type SipStartLine struct {
	Method
	UriScheme      string
	UserPart       string
	HostPart       string
	UserParameters *map[string]string

	Ruri          string
	UriParameters *map[string]string

	StatusCode   int
	ReasonPhrase string

	StartLine string //only set for incoming messages - to be removed!!!
}

func (sl *SipStartLine) SetRURIOnly(ruri string) {
	sl.UriScheme = ""
	sl.UserPart = ""
	sl.HostPart = ""
	sl.UserParameters = nil
	sl.UriParameters = nil
	sl.Ruri = ruri
}

type RequestPack struct {
	Method
	Max70         bool
	CustomHeaders *SipHeaders
	IsProbing     bool
}

type ResponsePack struct {
	StatusCode    int
	ReasonPhrase  string
	ContactHeader string

	CustomHeaders *SipHeaders
}

func NewResponsePackWarning(stc int, warning string) ResponsePack {
	return ResponsePack{
		StatusCode:    stc,
		CustomHeaders: NewSHQ850OrSIP(0, warning, ""),
	}
}

// reason != "" ==> Warning & Reason headers are always created.
//
// reason == "" ==>
//
// stc == 0 ==> only Warning header
//
// stc != 0 ==> only Reason header
func NewResponsePackSRW(stc int, warning string, reason string) ResponsePack {
	var hdrs *SipHeaders
	if reason == "" {
		hdrs = NewSHQ850OrSIP(stc, warning, "")
	} else {
		hdrs = NewSHQ850OrSIP(0, warning, "")
		hdrs.SetHeader(Reason, reason)
	}
	return ResponsePack{
		StatusCode:    stc,
		CustomHeaders: hdrs,
	}
}
