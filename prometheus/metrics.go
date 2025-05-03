package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	Registry    *prometheus.Registry
	ConSessions prometheus.Gauge
	Caps        prometheus.Gauge
}

func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	caps := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "LoadBalancer",
		Name:      "CallAttemptPerSecond",
		Help:      "Shows concurrent sessions active",
	})
	reg.MustRegister(caps)

	concurrentSessions := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "LoadBalancer",
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

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}
