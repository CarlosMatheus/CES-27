package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type ClockStruct struct {
	Id     int
	Clock  int
	Message string
	Request bool
}

var err string
var myPort string
var nServers int
var CliConn []*net.UDPConn
var ServerConn *net.UDPConn
var ch chan string
var myId string
var logicalClock ClockStruct

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

	n, _, err := ServerConn.ReadFromUDP(buf[0:])
	CheckError(err)

	var logicalClockReceived ClockStruct
	err = json.Unmarshal(buf[:n], &logicalClockReceived)
	CheckError(err)

	if logicalClockReceived.Clock > logicalClock.Clock {
		logicalClock.Clock = logicalClockReceived.Clock
	}

	logicalClock.Clock++

	fmt.Println("Current logical Clock:", logicalClock.Clock)
	fmt.Println("Received", logicalClockReceived.Message)

	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func doClientJob(otherProcess int) {
	jsonRequestByte, err := json.Marshal(logicalClock)
	CheckError(err)

	buf := jsonRequestByte
	_, err = CliConn[otherProcess].Write(buf)
	if err != nil {
		fmt.Println(jsonRequestByte, err)
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
	myPort = ":10001"
	CliConn = make([]*net.UDPConn, nServers)
	logicalClock = ClockStruct{Id: 0, Clock: 0, Message: "", Request: false}

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
						err = errors.New("Id out of range")
						PrintError(err)
					} else {
						go doClientJob(destiny)
					}
				} else {
					fmt.Println("Channel Closed!")
				}
			default:
				time.Sleep(time.Second * 1)
		}
	}
}
