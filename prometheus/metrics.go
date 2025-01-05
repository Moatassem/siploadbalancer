/*
# Software Name : Newkah-SIP-Layer
# SPDX-FileCopyrightText: Copyright (c) 2025 - Orange Business - OINIS/Services/NSF

# Authors:
# - Moatassem Talaat <moatassem.talaat@orange.com>

---
*/

package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all the custom Prometheus metrics for the application.
type Metrics struct {
	Registry    *prometheus.Registry
	ConSessions prometheus.Gauge
	Caps        prometheus.Gauge
}

// NewMetrics initializes a new custom Prometheus registry and returns an instance of Metrics.
func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()

	// Register default Go runtime metrics
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Initialize custom metrics here, e.g.:
	caps := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "SIP_Layer",
		Name:      "CallAttemptPerSecond",
		Help:      "Shows concurrent sessions active",
	})
	reg.MustRegister(caps)

	concurrentSessions := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "SIP_Layer",
		Name:      "ConcurrentSessions",
		Help:      "Shows concurrent sessions active",
	})
	reg.MustRegister(concurrentSessions)

	metrics := &Metrics{
		Registry:    reg,
		ConSessions: concurrentSessions,
		Caps:        caps,
	}

	return metrics
}

// Handler returns an HTTP handler that serves the metrics on a specified endpoint.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}
