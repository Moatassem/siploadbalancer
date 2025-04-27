/*
# Software Name : SIPLoadBalancer

# Author:
# - Moatassem Talaat <eng.moatassem@gmail.com>

---
*/

package webserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"
	. "siploadbalancer/global"
	"siploadbalancer/sip"
)

func StartWS(ip net.IP) {
	r := http.NewServeMux()

	r.HandleFunc("GET /api/v1/stats", serveStats)
	r.Handle("GET /metrics", Prometrics.Handler())
	r.HandleFunc("GET /", serveHome)

	ws := fmt.Sprintf("%s:%d", ip, HttpTcpPort)

	WtGrp.Add(1)
	go func() {
		defer WtGrp.Done()
		log.Fatal(http.ListenAndServe(ws, r))
	}()

	log.Println("Loading API Webserver...")
	log.Printf("Success: HTTP %s\n", ws)

	log.Printf("Prometheus metrics available at http://%s/metrics\n", ws)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Write(fmt.Appendf(nil, "<h1>%s API Webserver</h1>\n", BUE))
}

func serveStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	BToMB := func(b uint64) uint64 {
		return b / 1000 / 1000
	}

	data := struct {
		CPUCount        int
		GoRoutinesCount int
		Alloc           uint64
		System          uint64
		GCCycles        uint32
		CallsCacheCount int
	}{
		CPUCount:        runtime.NumCPU(),
		GoRoutinesCount: runtime.NumGoroutine(),
		Alloc:           BToMB(m.Alloc),
		System:          BToMB(m.Sys),
		GCCycles:        m.NumGC,
		CallsCacheCount: sip.LoadBalancer.CallsCacheCount(),
	}

	response, _ := json.Marshal(data)
	w.Write(response)
}
