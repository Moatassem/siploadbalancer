package global

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"runtime"
	"strings"

	"slices"

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
	markBytes := []byte(markstrng)
	for i := 0; i <= len(pdu)-len(markBytes); i++ {
		k := 0
		for k < len(markBytes) {
			if pdu[i+k] != markBytes[k] {
				goto nextloop
			}
			k++
		}
		return i
	nextloop:
	}
	return -1
}

func BuildUDPAddr(host, port string) (*net.UDPAddr, error) {
	if port == "" {
		return net.ResolveUDPAddr("udp", host+":5060")
	}
	return net.ResolveUDPAddr("udp", host+":"+port)
}

func LogCallStack(r any) {
	log.Printf("Panic Recovered! Error:\n%v", r)
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	log.Printf("Stack trace:\n%s\n", buf[:n])
}

func DropVisualSeparators(strng string) string {
	strng = strings.ReplaceAll(strng, ".", "")
	strng = strings.ReplaceAll(strng, "-", "")
	strng = strings.ReplaceAll(strng, "(", "")
	strng = strings.ReplaceAll(strng, ")", "")
	return strng
}

func CleanAndSplitHeader(HeaderValue string, DropParameterValueDQ ...bool) map[string]string {
	if HeaderValue == "" {
		return nil
	}

	NVC := map[string]string{}
	splitChar := ';'

	splitCharFirstIndex := strings.IndexRune(HeaderValue, splitChar)
	if splitCharFirstIndex == -1 {
		NVC["!headerValue"] = HeaderValue
		return NVC
	} else {
		NVC["!headerValue"] = HeaderValue[:splitCharFirstIndex]
	}

	chrlst := []rune(HeaderValue[splitCharFirstIndex:])
	sb := strings.Builder{}

	var fn, fv string
	DQO := false
	dropDQ := len(DropParameterValueDQ) > 0 && DropParameterValueDQ[0]

	for i := 0; i < len(chrlst); {
		switch chrlst[i] {
		case ' ':
			if DQO {
				sb.WriteRune(chrlst[i])
			}
		case '=':
			if DQO {
				sb.WriteRune(chrlst[i])
			} else {
				fn = sb.String()
				sb.Reset()
			}
		case splitChar:
			if DQO {
				sb.WriteRune(chrlst[i])
			} else {
				if sb.Len() == 0 {
					break
				}
				fv = sb.String()
				NVC[fn] = DropConcatenationChars(fv, dropDQ)
				fn, fv = "", ""
				sb.Reset()
			}
		case '"':
			if DQO {
				fv = sb.String()
				NVC[fn] = DropConcatenationChars(fv, dropDQ)
				fn, fv = "", ""
				sb.Reset()
				DQO = false
			} else {
				DQO = true
			}
		default:
			sb.WriteRune(chrlst[i])
		}
		chrlst = append(chrlst[:i], chrlst[i+1:]...)
	}

	if fn != "" && sb.Len() > 0 {
		fv = sb.String()
		NVC[fn] = DropConcatenationChars(fv, dropDQ)
	}

	return NVC
}

func DropConcatenationChars(s string, dropDQ bool) string {
	if dropDQ {
		s = strings.ReplaceAll(s, "'", "")
		return strings.ReplaceAll(s, `"`, "")
	}
	return s
}

func RandomNumMinMax(min int, max int) int {
	return rand.Intn(max-min+1) + min
}

func RandomNum(max int) int {
	return rand.Intn(max)
}

func GetBodyType(contentType string) BodyType {
	contentType = ASCIIToLower(contentType)
	for k, v := range DicBodyContentType {
		if v == contentType {
			return k
		}
	}
	if strings.Contains(contentType, "xml") {
		return AnyXML
	}
	return Invalid
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

func LowerDash(s string) string {
	return strings.ReplaceAll(ASCIIToLower(s), " ", "-")
}

func ASCIIPascal(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := range len(s) {
		c := s[i]
		if 'a' <= c && c <= 'z' && (i == 0 || s[i-1] == '-') {
			c -= byte(DeltaRune)
		}
		b.WriteByte(c)
	}
	return b.String()
}

func HeaderCase(h string) string {
	h = ASCIIToLower(h)
	for k := range HeaderStringtoEnum {
		if ASCIIToLower(k) == h {
			return k
		}
	}
	return ASCIIPascal(h)
}

func ASCIIToLowerInPlace(s []byte) {
	for i := range s {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		s[i] = c
	}
}

func RMatch(s string, rgxfp FieldPattern, mtch *[]string) bool {
	if s == "" {
		return false
	}
	*mtch = DicFieldRegEx[rgxfp].FindStringSubmatch(s)
	return *mtch != nil
}

func RReplace1(input string, rgxfp FieldPattern, replacement string) string {
	return DicFieldRegEx[rgxfp].ReplaceAllString(input, replacement)
}

func RReplaceNumberOnly(input string, replacement string) string {
	return DicFieldRegEx[ReplaceNumberOnly].ReplaceAllString(input, replacement)
}

func TranslateInternal(input string, matches []string) (string, error) {
	if input == "" {
		return "", nil
	}
	if matches == nil {
		return "", fmt.Errorf("empty matches slice")
	}
	sbToInt := func(sb strings.Builder) int {
		return Str2Int[int](sb.String())
	}

	item := func(idx int, dblbrkt bool) string {
		if idx >= len(matches) {
			if dblbrkt {
				return fmt.Sprintf("${%v}", idx)
			}
			return fmt.Sprintf("$%v", idx)
		}
		return matches[idx]
	}

	var b strings.Builder
outerloop:
	for i := 0; i < len(input); i++ {
		c := input[i]
		if c == '$' {
			i++
			if i == len(input) {
				b.WriteByte(c)
				return b.String(), nil
			}
			c = input[i]
			if c == '$' {
				b.WriteByte(c)
				continue outerloop
			}
			var grpsb strings.Builder
			for {
				if '0' <= c && c <= '9' {
					grpsb.WriteByte(c)
					i++
					if i == len(input) {
						v := item(sbToInt(grpsb), false)
						b.WriteString(v)
						return b.String(), nil
					}
					c = input[i]
				} else if c == '{' {
					if grpsb.Len() == 0 {
						break
					} else {
						b.WriteByte(c)
						v := item(sbToInt(grpsb), false)
						b.WriteString(v)
						continue outerloop
					}
				} else {
					if grpsb.Len() == 0 {
						b.WriteByte('$')
						b.WriteByte(c)
					} else {
						v := item(sbToInt(grpsb), false)
						b.WriteString(v)
					}
					continue outerloop
				}
			}
			for {
				i++
				if i == len(input) {
					return "", fmt.Errorf("bracket unclosed")
				}
				c = input[i]
				if '0' <= c && c <= '9' {
					grpsb.WriteByte(c)
				} else if c == '}' {
					if grpsb.Len() == 0 {
						return "", fmt.Errorf("bracket closed without group index")
					}
					v := item(sbToInt(grpsb), true)
					b.WriteString(v)
					continue outerloop
				} else if c == '{' {
					b.WriteByte(c)
					continue outerloop
				} else {
					return "", fmt.Errorf("invalid character within bracket")
				}
			}
		}
		b.WriteByte(c)
	}
	return b.String(), nil
}

func TranslateExternal(input string, rgxstring string, trans string) string {
	rgx, err := regexp.Compile(rgxstring)
	if err != nil {
		return ""
	}
	result := []byte{}
	result = rgx.ExpandString(result, trans, input, rgx.FindStringSubmatchIndex(input))
	return string(result)
}

// Use rgx.FindStringSubmatchIndex(input) to get matches
func TranslateResult(rgx *regexp.Regexp, input string, trans string, matches []int) string {
	result := []byte{}
	result = rgx.ExpandString(result, trans, input, matches)
	return string(result)
}

func (m Method) IsDialogueCreating() bool {
	switch m {
	case OPTIONS, INVITE, MESSAGE, REGISTER:
		return true
	}
	return false
}

func (m Method) RequiresACK() bool {
	switch m {
	case INVITE, ReINVITE:
		return true
	}
	return false
}

// =====================================================

func Any[T any](items []*T, predict func(*T) bool) bool {
	return slices.ContainsFunc(items, predict)
}

func Find[T any](items []*T, predict func(*T) bool) *T {
	for _, item := range items {
		if predict(item) {
			return item
		}
	}
	return nil
}

func Filter[T any](items []*T, predict func(*T) bool) []*T {
	result := []*T{}
	for _, item := range items {
		if predict(item) {
			result = append(result, item)
		}
	}
	return result
}

func FindFirstValue[T1 comparable, T2 any](m map[T1]*T2, predict func(*T2) bool) *T2 {
	for _, item := range m {
		if predict(item) {
			return item
		}
	}
	return nil
}

func FirstKeyValue[T1 comparable, T2 any](m map[T1]T2) (T1, T2) {
	var key T1
	var value T2
	for k, v := range m {
		return k, v
	}
	return key, value
}

func FirstKey[T1 comparable, T2 any](m map[T1]T2) T1 {
	k, _ := FirstKeyValue(m)
	return k
}

func FirstValue[T1 comparable, T2 any](m map[T1]T2) T2 {
	_, v := FirstKeyValue(m)
	return v
}

func GetEnum[T1 comparable, T2 comparable](m map[T1]T2, i T2) T1 {
	var rslt T1
	for k, v := range m {
		if v == i {
			return k
		}
	}
	return rslt
}

// =====================================================

func (he HeaderEnum) LowerCaseString() string {
	h := HeaderEnumToString[he]
	return ASCIIToLower(h)
}

func (he HeaderEnum) String() string {
	return HeaderEnumToString[he]
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
