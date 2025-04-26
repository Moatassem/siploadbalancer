package sip

import (
	"fmt"
	"math/rand"
	"net"

	. "siploadbalancer/global"
)

type SipHeaders struct {
	hmap   map[string][]string
	hnames []string
}

func NewSipHeaders() SipHeaders {
	return SipHeaders{hmap: make(map[string][]string)}
}

func NewSHsPointer() *SipHeaders {
	headers := NewSipHeaders()
	return &headers
}

func (hdrs *SipHeaders) DropTopVia() {
	currentVia := hdrs.hmap[ViaSmall]
	if len(currentVia) >= 2 {
		hdrs.hmap[ViaSmall] = currentVia[1:]
	}
}

func (hdrs *SipHeaders) AddTopVia() {
	getBranch := func() string {
		charset := "abcdefghijklmnopqrstuvwxyz0123456789"
		result := make([]byte, 10)
		for i := range result {
			result[i] = charset[rand.Intn(len(charset))]
		}
		return MagicCookie + string(result)
	}

	getHeader := func() string {
		udpsocket := ServerConnection.LocalAddr().(*net.UDPAddr)
		return fmt.Sprintf("SIP/2.0/UDP %s;branch=%s", udpsocket, getBranch())
	}

	currentVia := hdrs.hmap[ViaSmall]
	hvalues := make([]string, 0, 1+len(currentVia))
	hvalues = append(hvalues, getHeader())
	hvalues = append(hvalues, currentVia...)
	hdrs.hmap[ViaSmall] = hvalues
}

func (hdrs *SipHeaders) Add(headerName string, headerValue string) {
	hdrs.hnames = append(hdrs.hnames, headerName)
	hdrs.hmap[headerName] = append(hdrs.hmap[headerName], headerValue)
}

func (headers *SipHeaders) Values(headerName string) (bool, []string) {
	v, ok := headers.hmap[headerName]
	if ok {
		return true, v
	}

	return false, nil
}
