package sip

import (
	"fmt"
	"net"
	"os"
	"siploadbalancer/global"
)

func startListening(ip net.IP, prt int) (*net.UDPConn, error) {
	socket := net.UDPAddr{}
	socket.IP = ip
	socket.Port = prt
	return net.ListenUDP("udp", &socket)
}

func StartServer(redisskt string, ipv4 string) (*net.UDPConn, net.IP) {
	serverIP := net.ParseIP(ipv4)

	fmt.Print("Attempting to listen on SIP...")
	serverUDPListener, err := startListening(serverIP, global.SipUdpPort)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	ServerConnection = serverUDPListener
	startWorkers()
	udpLoopWorkers()
	fmt.Println("Success: UDP", serverUDPListener.LocalAddr().String())

	fmt.Print("Checking Caching Server...")
	// ripv4skt, err := redis.SetupCheckRedis(redisskt, "", 0, 15) //TODO: add redis password, db and expiryMin
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(3)
	// }
	// fmt.Printf("Ready! [%s]\n", ripv4skt)
	fmt.Println("Skipped!")

	return serverUDPListener, serverIP
}

// func periodicUAProbing(conn *net.UDPConn) {
// 	global.WtGrp.Add(1)
// 	defer global.WtGrp.Done()
// 	ticker := time.NewTicker(time.Duration(ProbingInterval) * time.Second)
// 	for range ticker.C {
// 		ProbeUA(conn, ASUserAgent)
// 		for _, phne := range phone.Phones.All() {
// 			if phne.IsReachable && phne.IsRegistered {
// 				ProbeUA(conn, phne.GetUA())
// 			}
// 		}
// 	}
// }
