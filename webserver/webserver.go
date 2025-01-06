/*
# Software Name : Newkah-SIP-Layer
# SPDX-FileCopyrightText: Copyright (c) 2025 - Orange Business - OINIS/Services/NSF

# Authors:
# - Moatassem Talaat <moatassem.talaat@orange.com>

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
)

func StartWS(ip net.IP) {
	r := http.NewServeMux()

	r.HandleFunc("GET /api/v1/stats", serveStats)
	r.Handle("GET /metrics", Prometrics.Handler())
	r.HandleFunc("GET /", serveHome)

	ws := fmt.Sprintf("%s:%d", ip, HttpTcpPort)

	go func() {
		defer WtGrp.Done()
		log.Fatal(http.ListenAndServe(ws, r))
	}()

	log.Println("Loading API Webserver...")
	log.Println(fmt.Sprintf("Success: HTTP %s", ws))

	log.Println(fmt.Sprintf("Prometheus metrics available at http://%s/metrics\n", ws))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("<h1> %s API Webserver</h1>\n", BUE)))
}

func serveStats(w http.ResponseWriter, r *http.Request) {
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
	}{CPUCount: runtime.NumCPU(),
		GoRoutinesCount: runtime.NumGoroutine(),
		Alloc:           BToMB(m.Alloc),
		System:          BToMB(m.Sys),
		GCCycles:        m.NumGC,
	}

	response, _ := json.Marshal(data)
	w.Write(response)
}
