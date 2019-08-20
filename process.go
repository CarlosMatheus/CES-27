package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

var err string
var myPort string
var nServers int
var CliConn []*net.UDPConn //
var ServerConn *net.UDPConn // connection with my server (where I receive messages from others processes)

/* Simple function to verify error */
func CheckError(err error) {
	// same function used on the wordpress tutorial
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func PrintError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func doServerJob() {
	// read (one time ) the message from the UDP connection
	// Write on the screen the received message ( indicate whether )

	// Lets prepare an address at any address at port 10001
	//ServerAddr, err := net.ResolveUDPAddr("udp", ":10001")
	//CheckError(err)
	//
	//ServerConn, err = net.ListenUDP("udp", ServerAddr)
	//CheckError(err)

	//defer ServerConn.Close()

	buf := make([]byte, 1024)

		n, addr, err := ServerConn.ReadFromUDP(buf)
		fmt.Println("Received", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
}

func doClientJob(otherProcess int, i int) {
	// Send message (with value i) to the server of the process other server
	//addr := "127.0.0.1:" + strconv.Itoa(otherProcess + 10002)

	//fmt.Println(addr)

	//ServerAddr, err := net.ResolveUDPAddr("udp", addr)
	//ServerAddr, err := net.ResolveUDPAddr("udp", ":10001")
	//CheckError(err)
	//
	//LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	//CheckError(err)
	//
	//CliConn[otherProcess], err = net.DialUDP("udp", LocalAddr, ServerAddr)
	//CheckError(err)

	//defer CliConn[otherProcess].Close()
	//idx := 0

	msg := strconv.Itoa(i)
	i++
	buf := []byte(msg)
	_, err := CliConn[otherProcess].Write(buf)
	if err != nil {
		fmt.Println(msg, err)
	}
	time.Sleep(time.Second * 1)
}

func initConnections() (error) {
	nServers = len(os.Args) - 2
	/* the 2 remove the name (Process) and remove the fist port, in the case it is my port */
	if nServers <= 0 {
		return errors.New("Insuficient number of servers")
	}

	myPort = os.Args[1]

	CliConn = make([]*net.UDPConn, nServers)

	//ServerConn = make(*net.UDPConn)
	//fmt.Println(nServers)

	// Init client
	for otherProcess := 0; otherProcess < nServers; otherProcess++ {

		port := os.Args[otherProcess + 2]

		ServerAddr, err := net.ResolveUDPAddr("udp", port)
		CheckError(err)

		LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
		CheckError(err)

		CliConn[otherProcess], err = net.DialUDP("udp", LocalAddr, ServerAddr)
		CheckError(err)
	}
	// init server
	ServerAddr, err := net.ResolveUDPAddr("udp", myPort)
	CheckError(err)

	ServerConn, err = net.ListenUDP("udp", ServerAddr)
	CheckError(err)


	return nil
}

func main() {

	e := initConnections()

	if e != nil {
		fmt.Println(e)
		return
	}

	defer ServerConn.Close()

	for i := 0; i < nServers; i++ {
		defer CliConn[i].Close()
	}

	i:=0

	for {
		// server
		go doServerJob()

		// Client
		for j := 0; j < nServers; j++ {
			//fmt.Println("j", j)
			go doClientJob(j, i)

			// wait a while
			time.Sleep(time.Second * 1)
			i++
		}
	}
}

