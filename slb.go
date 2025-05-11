/*
# Software Name : SIPLoadBalancer

# Author:
# - Moatassem Talaat <eng.moatassem@gmail.com>

---
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"siploadbalancer/cl"
	"siploadbalancer/global"
	"siploadbalancer/prometheus"
	"siploadbalancer/sip"
	"siploadbalancer/webserver"
)

func greeting() {
	fmt.Printf("Welcome to MT %s\n", global.BUE)
}

func main() {
	greeting()
	global.Prometrics = prometheus.NewMetrics()
	ip, hp, rate := sip.InitializeServer(readJsonFile())
	global.CallLimiter = cl.NewCallLimiter(rate, global.Prometrics, &global.WtGrp)
	// defer sip.ServerConnection.Close()
	webserver.StartWS(ip, hp)
	sip.StartSS()
	global.WtGrp.Wait()
}

func readJsonFile() []byte {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		os.Exit(1)
	}
	exeDir := filepath.Dir(exePath)

	jsonPath := filepath.Join(exeDir, "data.json")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		os.Exit(1)
	}

	return data
}
