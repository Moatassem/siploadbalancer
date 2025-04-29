package global

import (
	"bytes"
	"log"
	"net"
	"runtime"
	"strings"

	"golang.org/x/exp/rand"
)

func AreUAddrsEqual(addr1, addr2 *net.UDPAddr) bool {
	if addr1 == nil || addr2 == nil {
		return addr1 == addr2
	}
	return addr1.IP.Equal(addr2.IP) && addr1.Port == addr2.Port && addr1.Zone == addr2.Zone
}

func Str2Int[T int | int8 | int16 | int32 | int64](s string) T {
	var out T
	if len(s) == 0 {
		return out
	}
	idx := 0
	isN := s[idx] == '-'
	if isN {
		idx++
	}
	for i := idx; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return out
		}
		out = out*10 + T(s[i]-'0')
	}
	if isN {
		return -out
	}
	return out
}

func Str2uint[T uint | uint8 | uint16 | uint32 | uint64](s string) T {
	var out T
	if len(s) == 0 {
		return out
	}
	for i := range len(s) {
		out = out*10 + T(s[i]-'0')
	}
	return out
}

func Str2Uint[T uint | uint8 | uint16 | uint32 | uint64](s string) T {
	var out T
	if len(s) == 0 {
		return out
	}
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return out
		}
		out = out*10 + T(s[i]-'0')
	}
	return out
}

func Int2Str(val int) string {
	if val == 0 {
		return "0"
	}
	buf := make([]byte, 10)
	return int2str(buf, val)
}

func Uint16ToStr(val uint16) string {
	if val == 0 {
		return "0"
	}
	buf := make([]byte, 5)
	return uint2str(buf, val)
}

// Uint32ToStr converts a uint32 to its string representation.
func Uint32ToStr(val uint32) string {
	if val == 0 {
		return "0"
	}
	buf := make([]byte, 10)
	return uint2str(buf, val)
}

// Uint64ToStr converts a uint64 to its string representation.
func Uint64ToStr(val uint64) string {
	if val == 0 {
		return "0"
	}
	buf := make([]byte, 20)
	return uint2str(buf, val)
}

func uint2str[T uint16 | uint32 | uint64](buf []byte, val T) string {
	i := len(buf)
	for val >= 10 {
		i--
		buf[i] = '0' + byte(val%10)
		val /= 10
	}
	i--
	buf[i] = '0' + byte(val)

	return string(buf[i:])
}

func int2str[T int | int8 | int16 | int32 | int64](buf []byte, val T) string {
	isNeg := val < 0
	if isNeg {
		val *= -1
	}
	i := len(buf)
	for val >= 10 {
		i--
		buf[i] = '0' + byte(val%10)
		val /= 10
	}
	i--
	buf[i] = '0' + byte(val)

	if isNeg {
		return "-" + string(buf[i:])
	}
	return string(buf[i:])
}

func GetNextIndex(pdu []byte, markstrng string) int {
	return bytes.Index(pdu, []byte(markstrng))
}

func BuildSipUdpSocket(host, port string) (*net.UDPAddr, error) {
	if port == "" {
		return net.ResolveUDPAddr("udp", host+":5060")
	}
	return net.ResolveUDPAddr("udp", host+":"+port)
}

func BuildUdpSocket(udpskt string) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", udpskt)
}

func LogCallStack(r any) {
	log.Printf("Panic Recovered! Error:\n%v", r)
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	log.Printf("Stack trace:\n%s\n", buf[:n])
}

func RandomNumMinMax(min int, max int) int {
	return rand.Intn(max-min+1) + min
}

func RandomNum(max int) int {
	return rand.Intn(max)
}

func ASCIIToLower(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := range len(s) {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += byte(DeltaRune)
		}
		b.WriteByte(c)
	}
	return b.String()
}

func ASCIIToUpper(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := range len(s) {
		c := s[i]
		if 'a' <= c && c <= 'z' {
			c -= byte(DeltaRune)
		}
		b.WriteByte(c)
	}
	return b.String()
}

func RMatch(s string, rgxfp FieldPattern, mtch *[]string) bool {
	if s == "" {
		return false
	}
	*mtch = DicFieldRegEx[rgxfp].FindStringSubmatch(s)
	return *mtch != nil
}

func (m Method) IsDialogueCreating() bool {
	switch m {
	case OPTIONS, INVITE, MESSAGE, REGISTER, SUBSCRIBE:
		return true
	}
	return false
}

// =====================================================

func Find[T any](items []*T, predicate func(*T) bool) *T {
	for _, item := range items {
		if predicate(item) {
			return item
		}
	}
	return nil
}

func All[T any](items []*T, predicate func(*T) bool) bool {
	for _, item := range items {
		if !predicate(item) {
			return false
		}
	}
	return true
}

func IsProvisional(sc int) bool {
	return 100 <= sc && sc <= 199
}

func IsProvisional18x(sc int) bool {
	return 180 <= sc && sc <= 189
}

func Is18xOrPositive(sc int) bool {
	return (180 <= sc && sc <= 189) || (200 <= sc && sc <= 299)
}

func IsFinal(sc int) bool {
	return 200 <= sc && sc <= 699
}

func IsPositive(sc int) bool {
	return 200 <= sc && sc <= 299
}

func IsNegative(sc int) bool {
	return 300 <= sc && sc <= 699
}

func IsRedirection(sc int) bool {
	return 300 <= sc && sc <= 399
}

func IsNegativeClient(sc int) bool {
	return 400 <= sc && sc <= 499
}

func IsNegativeServer(sc int) bool {
	return 500 <= sc && sc <= 599
}

func IsNegativeGlobal(sc int) bool {
	return 600 <= sc && sc <= 699
}
