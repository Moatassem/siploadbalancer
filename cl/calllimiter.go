/*
# Software Name : Newkah-SIP-Layer
# SPDX-FileCopyrightText: Copyright (c) 2025 - Orange Business - OINIS/Services/NSF

# Authors:
# - Moatassem Talaat <moatassem.talaat@orange.com>

---
*/

package cl

import (
	"siploadbalancer/prometheus"
	"sync"
	"time"
)

type CallLimiter struct {
	rate      int          // rate limiter
	ticker    *time.Ticker // ticker for timing
	callCount int          // current call count
	mu        sync.Mutex   // mutex for thread safety
}

func NewCallLimiter(rate int, pm *prometheus.Metrics) *CallLimiter {
	cl := &CallLimiter{
		rate:   rate,
		ticker: time.NewTicker(time.Second),
	}
	go cl.resetCount(pm)
	return cl
}

func (clmtr *CallLimiter) resetCount(pm *prometheus.Metrics) {
	for range clmtr.ticker.C {
		clmtr.mu.Lock()
		pm.Caps.Set(float64(clmtr.callCount))
		clmtr.callCount = 0
		clmtr.mu.Unlock()
	}
}

func (clmtr *CallLimiter) AcceptNewCall() bool {
	clmtr.mu.Lock()
	defer clmtr.mu.Unlock()
	if clmtr.rate == -1 || clmtr.callCount < clmtr.rate {
		clmtr.callCount++
		return true // Call can be attempted
	}
	return false // Rate limit exceeded
}
