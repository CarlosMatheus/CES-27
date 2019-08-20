package main

import (
	"bufio"
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
var CliConn []*net.UDPConn
var ServerConn *net.UDPConn // connection with my server (where I receive messages from others processes)
var ch chan string
var myId string

/* Simple function to verify error */
func CheckError(err error) {
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
	buf := make([]byte, 1024)

		n, addr, err := ServerConn.ReadFromUDP(buf)
		fmt.Println("Received", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
}

func doClientJob(otherProcess int, i int) {
	msg := strconv.Itoa(i)
	i++
	buf := []byte(msg)
	_, err := CliConn[otherProcess].Write(buf)
	if err != nil {
		fmt.Println(msg, err)
	}
	time.Sleep(time.Second * 1)
}

func getMyPortNumber(portArg string, myId string) string {
	portNum, err := strconv.Atoi(portArg[1:])
	CheckError(err)

	idNum, err := strconv.Atoi(myId)
	CheckError(err)

	num := portNum - idNum + 1

	newPortStr := strconv.Itoa(num)

	return ":" + newPortStr
}

func initConnections() error {
	ch  = make(chan string)

	nonOtherServers := 2
	nServers = len(os.Args) - nonOtherServers

	/* the 2 remove the name (Process) and remove the fist port, in the case it is my port */
	if nServers <= 0 {
		return errors.New("insufficient number of servers")
	}

	myId = os.Args[1]
	//myPort = os.Args[2]
	myPort = getMyPortNumber(os.Args[2], myId)
	//fmt.Println(myPort)

	CliConn = make([]*net.UDPConn, nServers)

	// Init client
	for otherProcess := 0; otherProcess < nServers; otherProcess++ {
		port := os.Args[otherProcess + nonOtherServers]

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

func readInput(ch chan string) {
	// Non-blocking async routine to listen for terminal input
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}
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

	go readInput(ch)

	for {
		go doServerJob()

		select {
		case x, valid := <-ch:
			if valid {
				fmt.Printf("Destiny Id: %s \n", x)
				destiny, err := strconv.Atoi(x)
				CheckError(err)

				destiny--
				if destiny >= nServers {
					err = errors.New("id out of range")
					PrintError(err)
				} else {
					go doClientJob(destiny, 100)
				}
				//CheckError(err)
				//for j := 0; j < nServers; j++ {
				//	go doClientJob(j, 100)
				//}
			} else {
				fmt.Println("Channel Closed!")
			}
		default:
			time.Sleep(time.Second * 1)
		}
	}
}

