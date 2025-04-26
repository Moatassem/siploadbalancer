package main

import (
	"fmt"
	"os"
	. "siploadbalancer/global"
	"siploadbalancer/prometheus"
	"siploadbalancer/sip"
	"siploadbalancer/webserver"
)

// environment variables
const (
	Redis_IpPort    string = "redis_socket"
	Own_Ip_IPv4     string = "server_ipv4"
	Own_Sip_UdpPort string = "sip_udp_port"
	Own_Http_port   string = "http_port"
	Own_Call_Limit  string = "call_limit"
)

// check global environment variables
func checkEVs() (redisskt string, ipv4 string) {
	redisskt, ok := os.LookupEnv(Redis_IpPort)
	if !ok {
		fmt.Println("No Redis socket provided!")
		os.Exit(1)
	}
	ipv4, ok = os.LookupEnv(Own_Ip_IPv4)
	if !ok {
		fmt.Println("No self IPv4 address provided!")
		os.Exit(1)
	}
	sup, ok := os.LookupEnv(Own_Sip_UdpPort)
	if ok {
		SipUdpPort = Str2Int[int](sup)
	} else {
		SipUdpPort = 5066
	}
	hp, ok := os.LookupEnv(Own_Http_port)
	if ok {
		HttpTcpPort = Str2Int[int](hp)
	} else {
		HttpTcpPort = 8081
	}
	cl, ok := os.LookupEnv(Own_Call_Limit)
	if ok {
		RateLimit = Str2Int[int](cl)
	} else {
		RateLimit = 1500
	}
	return
}

func greeting() {
	fmt.Printf("Welcome to %s - Product of OINIS NSF\n", BUE)
}

func main() {
	greeting()
	Prometrics = prometheus.NewMetrics()
	conn, ip := sip.StartServer(checkEVs())
	defer conn.Close()
	webserver.StartWS(ip)
	WtGrp.Wait()
}
