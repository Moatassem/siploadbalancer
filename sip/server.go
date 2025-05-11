package sip

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"siploadbalancer/global"
	"time"
)

type inputData struct {
	IPv4          string `json:"ipv4"`
	SipUdpPort    int    `json:"sipUdpPort"`
	HttpPort      int    `json:"httpPort"`
	CachingServer string `json:"cachingServer"`

	LoadbalanceMode          string `json:"loadbalancemode"`
	MaxCallAttemptsPerSecond int    `json:"maxCallAttemptsPerSecond"`
	ProbingInterval          int    `json:"probingInterval"`
	TimeoutTimerDuration     int    `json:"timeoutTimerDuration"`
	ClearTimerDuration       int    `json:"clearTimerDuration"`

	Servers []struct {
		Ipv4        string `json:"ipv4"`
		Port        int    `json:"port"`
		Description string `json:"description"`
		Weight      int    `json:"weight"`
		Cost        int    `json:"cost"`
	} `json:"servers"`
}

func startListening(ip net.IP, prt int) (*net.UDPConn, error) {
	socket := net.UDPAddr{}
	socket.IP = ip
	socket.Port = prt
	return net.ListenUDP("udp", &socket)
}

func InitializeServer(data []byte) (net.IP, int, int) {
	var (
		inputData inputData
		err       error
	)

	if err = json.Unmarshal(data, &inputData); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	serverIP := net.ParseIP(inputData.IPv4)
	fmt.Print("Attempting to listen on SIP...")
	if ServerConnection, err = startListening(serverIP, inputData.SipUdpPort); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	fmt.Println("Success: UDP", ServerConnection.LocalAddr().String())

	fmt.Print("Checking Caching Server...")
	// ripv4skt, err := redis.SetupCheckRedis(redisskt, "", 0, 15) //TODO: add redis password, db and expiryMin
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(3)
	// }
	// fmt.Printf("Ready! [%s]\n", ripv4skt)
	fmt.Println("Skipped!")

	LoadBalancer = NewLoadBalancer(inputData)

	return serverIP, inputData.HttpPort, inputData.MaxCallAttemptsPerSecond
}

func StartSS() {
	startWorkers()
	udpLoopWorkers()
	periodicProbing()

	fmt.Println("SipLoadBalancer Server Ready!")
}

func periodicProbing() {
	global.WtGrp.Add(1)
	duration := time.Duration(LoadBalancer.ProbingInterval) * time.Second
	ticker := time.NewTicker(duration)
	LoadBalancer.ProbeSipNodes() // to run once when system starts
	go func() {
		defer global.WtGrp.Done()
		for range ticker.C {
			LoadBalancer.ProbeSipNodes()
		}
	}()
}
