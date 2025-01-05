package global

import (
	"siploadbalancer/cl"
	"siploadbalancer/prometheus"
	"sync"
)

const (
	BUE        string = "SipLoadBalancer/v1.0"
	BufferSize int    = 4096
)

var (
	SipUdpPort  int
	HttpTcpPort int
	RateLimit   int

	Prometrics *prometheus.Metrics

	WtGrp      sync.WaitGroup
	BufferPool *sync.Pool = NewSyncPool()

	CallLimiter *cl.CallLimiter
)

func NewSyncPool() *sync.Pool {
	return &sync.Pool{
		New: func() any {
			b := make([]byte, BufferSize)
			return &b
		},
	}
}
