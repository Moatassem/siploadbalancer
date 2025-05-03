package cl

import (
	"sync"
	"time"

	"siploadbalancer/prometheus"
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

	return cl
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

func (clmtr *CallLimiter) CanAcceptNewSession() bool {
	clmtr.mu.Lock()
	defer clmtr.mu.Unlock()
	if clmtr.rate == -1 || clmtr.callCount < clmtr.rate { // it is not <= because i didn't add yet 1 to callCount
		clmtr.callCount++
		return true // Call can be attempted
	}
	return false // Rate limit exceeded
}
