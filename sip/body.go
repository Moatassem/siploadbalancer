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

type MessageBody struct {
	PartsBytes map[BodyType]ContentPart //used to store incoming/outgoing body parts
	// BodyType     BodyType
	MessageBytes []byte //used to store the generated body bytes for sending msgs
}

type ContentPart struct {
	Headers SipHeaders
	Bytes   []byte
}

func NewContentPart(bt BodyType, bytes []byte) ContentPart {
	ct := ContentPart{}
	ct.Bytes = bytes
	ct.Headers = NewSH()
	ct.Headers.AddHeader(Content_Type, DicBodyContentType[bt])
	return ct
}

// ===============================================================

func NewMSCXML(xml string) MessageBody {
	hdrs := *NewSIPHeaders()
	hdrs.AddHeader(Content_Length, DicBodyContentType[MSCXML])
	return MessageBody{PartsBytes: map[BodyType]ContentPart{MSCXML: {hdrs, []byte(xml)}}}
}

func NewTRJSON(jsonbytes []byte) MessageBody {
	hdrs := *NewSIPHeaders()
	hdrs.AddHeader(Content_Length, DicBodyContentType[AppJson])
	return MessageBody{PartsBytes: map[BodyType]ContentPart{AppJson: {hdrs, jsonbytes}}}
}

// ===============================================================

func (messagebody *MessageBody) WithNoBody() bool {
	return messagebody.PartsBytes == nil
}

func (messagebody *MessageBody) WithInvalidBody() bool {
	if messagebody.PartsBytes == nil {
		return false
	}
	for k := range messagebody.PartsBytes {
		if k == Invalid {
			return true
		}
	}
	return false
}

func (messagebody *MessageBody) IsMultiPartBody() bool {
	return len(messagebody.PartsBytes) > 1
}

func (messagebody *MessageBody) ContainsSDP() bool {
	if messagebody.PartsBytes == nil {
		return false
	}
	_, ok := messagebody.PartsBytes[SDP]
	return ok
}

func (messagebody *MessageBody) IsJSON() bool {
	if messagebody.PartsBytes == nil {
		return false
	}
	_, ok := messagebody.PartsBytes[AppJson]
	return ok
}

func (messagebody *MessageBody) ContentLength() int {
	return len(messagebody.MessageBytes)
}
