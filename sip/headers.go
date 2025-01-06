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
	"strings"
)

type SipHeaders struct {
	_map map[string][]string
}

func NewSH() SipHeaders {
	return SipHeaders{_map: map[string][]string{}}
}

// Used in outgoing messages - pointer i.e. mutable
func NewSIPHeaders() *SipHeaders {
	headers := NewSH()
	return &headers
}

func NewSHQ850OrSIP(Q850OrSIP int, Details string, retryAfter string) *SipHeaders {
	headers := NewSIPHeaders()
	if retryAfter != "" {
		headers.AddHeader(Retry_After, retryAfter)
	}
	if Q850OrSIP == 0 {
		if strings.TrimSpace(Details) != "" {
			headers.AddHeader(Warning, fmt.Sprintf("399 Newkah \"%s\"", Details))
		}
	} else {
		reason := ""
		if Q850OrSIP <= 127 {
			reason = fmt.Sprintf("Q.850;cause=%d", Q850OrSIP)
		} else {
			reason = fmt.Sprintf("SIP;cause=%d", Q850OrSIP)
		}
		if strings.TrimSpace(Details) != "" {
			reason += fmt.Sprintf(";text=\"%s\"", Details)
		}
		headers.AddHeader(Reason, reason)
	}
	return headers
}

// ==========================================

func (headers *SipHeaders) GetHeaderNames() []string {
	lst := []string{}
	for h := range headers._map {
		lst = append(lst, h)
	}
	return lst
}

// headerName is case insensitive
func (headers *SipHeaders) HeaderExists(headerName string) bool {
	headerName = ASCIIToLower(headerName)
	_, ok := headers._map[headerName]
	return ok
}

func (headers *SipHeaders) HeaderCount(headerName string) int {
	headerName = ASCIIToLower(headerName)
	v, ok := headers._map[headerName]
	if ok {
		return len(v)
	}
	return 0
}

func (headers *SipHeaders) DoesValueExistInHeader(headerName string, headerValue string) bool {
	headerValue = ASCIIToLower(headerValue)
	_, values := headers.Values(headerName)
	for _, hv := range values {
		if strings.Contains(ASCIIToLower(hv), headerValue) {
			return true
		}
	}
	return false
}

func (headers *SipHeaders) AddHeader(header HeaderEnum, headerValue string) {
	headers.Add(header.String(), headerValue)
}

func (headers *SipHeaders) Add(headerName string, headerValue string) {
	headerName = ASCIIToLower(headerName)
	v, ok := headers._map[headerName]
	if ok {
		headers._map[headerName] = append(v, headerValue)
	} else {
		headers._map[headerName] = []string{headerValue}
	}
}

func (headers *SipHeaders) SetHeader(header HeaderEnum, headerValue string) {
	headers.Set(header.String(), headerValue)
}

func (headers *SipHeaders) Set(headerName string, headerValue string) {
	headerName = ASCIIToLower(headerName)
	headers._map[headerName] = []string{headerValue}
}

func (headers *SipHeaders) ValuesHeader(header HeaderEnum) (bool, []string) {
	return headers.Values(header.String())
}

func (headers *SipHeaders) Values(headerName string) (bool, []string) {
	headerName = ASCIIToLower(headerName)
	v, ok := headers._map[headerName]
	if ok {
		return true, v
	} else {
		return false, nil
	}
}

func (headers *SipHeaders) ValuesWithHeaderPrefix(headersPrefix string) map[string][]string {
	headersPrefix = ASCIIToLower(headersPrefix)
	data := map[string][]string{}
	for k, v := range headers._map {
		if strings.HasPrefix(k, headersPrefix) {
			data[HeaderCase(k)] = v
		}
	}
	return data
}

func (headers *SipHeaders) ValueHeader(header HeaderEnum) string {
	return headers.Value(header.String())
}

func (headers *SipHeaders) Value(headerName string) string {
	if ok, v := headers.Values(headerName); ok {
		return v[0]
	}
	return ""
}

func (headers *SipHeaders) Delete(headerName string) bool {
	headerName = ASCIIToLower(headerName)
	_, ok := headers._map[headerName]
	if ok {
		delete(headers._map, headerName)
	}
	return ok
}

func (headers *SipHeaders) ContainsToTag() bool {
	toheader := headers._map["to"]
	return strings.Contains(ASCIIToLower(toheader[0]), "tag")
}
