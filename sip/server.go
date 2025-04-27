package sip

import (
	"fmt"
	"net"
	"os"
	"siploadbalancer/global"
	"strings"
	"time"
)

func startListening(ip net.IP, prt int) (*net.UDPConn, error) {
	socket := net.UDPAddr{}
	socket.IP = ip
	socket.Port = prt
	return net.ListenUDP("udp", &socket)
}

func StartServer(redisskt, ipv4, lbmode string, sipskts []string) net.IP {
	var err error

	serverIP := net.ParseIP(ipv4)
	fmt.Print("Attempting to listen on SIP...")
	ServerConnection, err = startListening(serverIP, global.SipUdpPort)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	startWorkers()
	udpLoopWorkers()
	fmt.Println("Success: UDP", ServerConnection.LocalAddr().String())

	fmt.Print("Checking Caching Server...")
	// ripv4skt, err := redis.SetupCheckRedis(redisskt, "", 0, 15) //TODO: add redis password, db and expiryMin
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(3)
	// }
	// fmt.Printf("Ready! [%s]\n", ripv4skt)
	fmt.Println("Skipped!")

	sipnodes := make([]*SipNode, 0, len(sipskts))
	for _, sipskt := range sipskts {
		skt := strings.Split(sipskt, ":")
		if len(skt) != 2 {
			fmt.Printf("SIP Server: %s - badly formatted - format: IPv4:Port;IPv4:Port;...", sipskt)
			continue
		}
		sipIpv4 := net.ParseIP(skt[0])
		if sipIpv4 == nil {
			fmt.Printf("SIP Server IPv4: %s - invalid", skt[0])
			continue
		}
		sipprt := global.Str2Int[int](skt[1])
		if sipprt == 0 {
			fmt.Printf("SIP Server Port: %s - invalid", skt[1])
			continue
		}
		wt := 100
		sipnodes = append(sipnodes, &SipNode{
			UdpAddr:     &net.UDPAddr{IP: sipIpv4, Port: sipprt, Zone: ""},
			Cost:        100,
			Weight:      wt,
			Description: "",
			IsAlive:     false,
			accWeight:   wt,
			Key:         global.GetTagOrKey(),
		})
	}

	LoadBalancer = NewLoadBalancer(lbmode, sipnodes)
	periodicProbing()

	return serverIP
}

func periodicProbing() {
	global.WtGrp.Add(1)
	ticker := time.NewTicker(global.ProbingInterval * time.Second)
	go func() {
		defer global.WtGrp.Done()
		for range ticker.C {
			LoadBalancer.ProbeSipNodes()
		}
	}()
}
