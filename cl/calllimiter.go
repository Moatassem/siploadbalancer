package cl

import (
	"fmt"
	"sync"
	"time"

	"siploadbalancer/prometheus"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	cyan   = "\033[36m"
	white  = "\033[37m"
	gray   = "\033[90m"
)

type CallLimiter struct {
	rate      int          // rate limiter
	ticker    *time.Ticker // ticker for timing
	callCount int          // current call count
	mu        sync.Mutex   // mutex for thread safety
}

func NewCallLimiter(rate int, pm *prometheus.Metrics, wg *sync.WaitGroup) *CallLimiter {
	cl := &CallLimiter{
		rate:   rate,
		ticker: time.NewTicker(time.Second),
	}
	wg.Add(1)
	go cl.resetCount(pm, wg)

	fmt.Printf("Call Limiter set: %s\n", getCLstate(rate))
	return cl
}

func getCLstate(rate int) string {
	switch rate {
	case -1:
		return fmt.Sprintf("%s%d %s%s", yellow, rate, "(Unlimited CAPS)", reset)
	case 0:
		return fmt.Sprintf("%s%d %s%s", red, rate, "(Server Disabled)", reset)
	default:
		return fmt.Sprintf("%s%d %s%s", green, rate, "CAPS", reset)
	}
}

func (clmtr *CallLimiter) resetCount(pm *prometheus.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()
	for range clmtr.ticker.C {
		clmtr.mu.Lock()
		pm.Caps.Set(float64(clmtr.callCount))
		clmtr.callCount = 0
		clmtr.mu.Unlock()
	}
}

func (clmtr *CallLimiter) IsExceeded() bool {
	clmtr.mu.Lock()
	defer clmtr.mu.Unlock()
	if clmtr.rate == -1 || clmtr.callCount < clmtr.rate { // it is not <= because i didn't add yet 1 to callCount
		clmtr.callCount++
		return false // Call can be attempted
	}
	return true // Rate limit exceeded
}
