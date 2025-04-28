package main

import (
	"fmt"
	"os"
	. "siploadbalancer/global"
	"siploadbalancer/prometheus"
	"siploadbalancer/sip"
	"siploadbalancer/webserver"
	"strings"
)

// environment variables
const (
	Redis_IpPort     string = "redis_socket"
	Own_Ip_IPv4      string = "server_ipv4"
	Own_Sip_UdpPort  string = "sip_udp_port"
	Own_Http_port    string = "http_port"
	LoadBalance_Mode string = "lbmode"
	SIP_UdpServers   string = "sip_udp_servers"
)

// check global environment variables
func checkEVs() (redisskt, ipv4, lbmode string, sipskts []string) {
	var ok bool
	// redisskt, ok := os.LookupEnv(Redis_IpPort)
	// if !ok {
	// 	fmt.Println("No Redis socket provided!")
	// 	os.Exit(1)
	// }

	if lbmode, ok = os.LookupEnv(LoadBalance_Mode); !ok {
		fmt.Println("No load-balancing mode provided!")
		os.Exit(1)
	}

	if ipv4, ok = os.LookupEnv(Own_Ip_IPv4); !ok {
		fmt.Println("No self IPv4 address provided!")
		os.Exit(1)
	}

	if sipnodes, ok := os.LookupEnv(SIP_UdpServers); ok {
		sipskts = strings.Split(sipnodes, ";")
	} else {
		fmt.Println("No SIP servers provided!")
		os.Exit(1)
	}

	if sup, ok := os.LookupEnv(Own_Sip_UdpPort); ok {
		SipUdpPort = Str2Int[int](sup)
	} else {
		SipUdpPort = 5066
	}

	if hp, ok := os.LookupEnv(Own_Http_port); ok {
		HttpTcpPort = Str2Int[int](hp)
	} else {
		HttpTcpPort = 8081
	}

	return
}

func greeting() {
	fmt.Printf("Welcome to %s - Product of OINIS NSF\n", BUE)
}

func main() {
	greeting()
	Prometrics = prometheus.NewMetrics()
	ip := sip.StartServer(checkEVs())
	// defer sip.ServerConnection.Close()
	webserver.StartWS(ip)
	WtGrp.Wait()
}
