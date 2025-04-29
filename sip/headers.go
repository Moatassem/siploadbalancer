package sip

import (
	"fmt"
	"net"
	"slices"
	"strings"

	. "siploadbalancer/global"
)

type SipHeaders struct {
	hmap   map[string][]string
	hnames []string
}

func NewSipHeaders() *SipHeaders {
	headers := SipHeaders{hmap: make(map[string][]string)}
	return &headers
}

func (hdrs *SipHeaders) DropTopVia() {
	hdrs.DropTopHeaderValue(ViaHeader)
}

func buildViaHeader(viaBranch string) string {
	udpsocket := ServerConnection.LocalAddr().(*net.UDPAddr)
	return fmt.Sprintf("SIP/2.0/UDP %s;branch=%s", udpsocket, viaBranch)
}

func (hdrs *SipHeaders) AddTopVia(viaBranch string) {
	hdrs.AddTopHeaderValue(ViaHeader, buildViaHeader(viaBranch))
}

func (hdrs *SipHeaders) Add(headerName string, headerValues ...string) {
	idx := hdrs.GetHeaderIndex(headerName)
	var hnm string
	if idx == -1 {
		hnm = headerName
		hdrs.hnames = append(hdrs.hnames, headerName)
	} else {
		hnm = hdrs.hnames[idx]
	}
	hdrs.hmap[hnm] = append(hdrs.hmap[hnm], headerValues...)
}

func (hdrs *SipHeaders) AddTopHeaderValue(headerName string, topValue string) {
	idx := hdrs.GetHeaderIndex(headerName)
	if idx == -1 {
		fmt.Printf("Could not find header [%s] values to amend", headerName)
		return
	}

	currentVia := hdrs.GetHeaderValues(ViaHeader)

	hvalues := make([]string, 0, 1+len(currentVia))
	hvalues = append(hvalues, topValue)
	hvalues = append(hvalues, currentVia...)

	hdrs.hmap[hdrs.hnames[idx]] = hvalues
}

func (hdrs *SipHeaders) DropTopHeaderValue(headerName string) {
	idx := hdrs.GetHeaderIndex(headerName)
	if idx == -1 {
		fmt.Printf("Could not find header [%s] values to amend", headerName)
		return
	}

	currentVia := hdrs.hmap[ViaHeader]

	if len(currentVia) < 2 {
		fmt.Printf("Could not find enough header [%s] values to adjust", headerName)
		return
	}

	hdrs.hmap[hdrs.hnames[idx]] = currentVia[1:]
}

func (headers *SipHeaders) Values(headerName string) (bool, []string) {
	v, ok := headers.hmap[headerName]
	if ok {
		return true, v
	}

	return false, nil
}

func (hdrs *SipHeaders) GetHeaderIndex(hn string) int {
	return slices.IndexFunc(hdrs.hnames, func(x string) bool { return strings.EqualFold(x, hn) })
}

func (hdrs *SipHeaders) GetHeaderValues(hn string) []string {
	idx := hdrs.GetHeaderIndex(hn)
	if idx == -1 {
		return nil
	}
	return hdrs.hmap[hdrs.hnames[idx]]
}
