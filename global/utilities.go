package global

import (
	"fmt"
	"log"
	"runtime"
)

func Str2int[T int | int8 | int16 | int32 | int64](s string) T {
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
	for i := 0; i < len(s); i++ {
		out = out*10 + T(s[i]-'0')
	}
	return out
}

func LogCallStack(r any) {
	log.Println(fmt.Sprintf("Panic Recovered! Error:\n%v\n", r))
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	log.Println(fmt.Sprintf("Stack trace:\n%s\n", buf[:n]))
}
