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
var CSCliConn *net.UDPConn
var ServerConn *net.UDPConn
var ch chan string
var myId string
var logicalClock ClockStruct
var holding bool
var allowedRequest []bool
var timeOut int
var heldQueue []ClockStruct
var wanting bool
var messageToSend string

func CheckError(err error) {
	/*
		Simple function to verify error
	*/
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

func executeReceiveMessage(logicalClockReceived ClockStruct, receivedId int) {
	if !holding {
		if logicalClock.Request {
			fmt.Println("Received request message: ", logicalClock.Message)
			answerBack(logicalClock.Message, receivedId)
		} else {
			if !wanting {
				fmt.Println("Received answer message", logicalClockReceived.Message, "from", receivedId)

				receivedIdx := receivedId - 1
				allowedRequest[receivedIdx] = true

				if checkAllowed(allowedRequest) {
					go holdCS(logicalClockReceived.Message)
				}
			} else {
				if logicalClockReceived.myTurn {
					go holdCS(logicalClockReceived.Message)
				} else {
					time.Sleep(time.Second * 1)  // wait the other process start
					broadcastMessage(messageToSend, true)
				}
			}
		}
	} else {
		heldQueue = append(heldQueue, logicalClockReceived)
		fmt.Println(logicalClockReceived.Message, "ignorado")
		fmt.Println(logicalClockReceived.Id, "Added to queue")
		fmt.Println(logicalClockReceived.Clock, "Received clock")
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
	logicalClock.Request = logicalClockReceived.Request
	receivedId := logicalClockReceived.Id

	logicalClock.Message = logicalClockReceived.Message

	executeClockCycle()
	executeReceiveMessage(logicalClockReceived, receivedId)
}

func holdCS(message string) {
	holding = true
	fmt.Println("Entrei na CS")

	answerBack(message, 0)
	time.Sleep(time.Second * 10)  // time on critical section

	holding = false
	fmt.Println("Sai da CS")

	for i := 0; i < len(heldQueue); i++ {
		answerBack(heldQueue[i].Message, heldQueue[i].Id)
	}
	heldQueue = make([]ClockStruct, 0)
}

func checkAllowed(allowedRequest []bool) bool {
	for i := 0; i < nServers; i++ {
		if allowedRequest[i] == false {
			return false
		}
	}
	return true
}

//func verifyWhoHolds() {
//	time.Sleep(time.Second * 3)  // timeout
//	if !checkAllowed(allowedRequest) {
//		for i := 0; i < nServers; i++ {
//			if allowedRequest[i] == false {
//				fmt.Printf("Servers that holds CS: %d \n", i+1)
//				wanting = true
//			}
//		}
//	}
//}

func doClientJob(otherProcess int) {

	jsonRequestByte, err := json.Marshal(logicalClock)
	CheckError(err)

	buf := jsonRequestByte

	if otherProcess == -1 {
		_, err = CSCliConn.Write(buf)
	} else {
		_, err = CliConn[otherProcess].Write(buf)
	}
	if err != nil {
		fmt.Println("error")
		fmt.Println(jsonRequestByte, err)
	}
	time.Sleep(time.Second * 1)
}

func getMyPortNumber(portArg string, myId string) string {
	portNum, err := strconv.Atoi(portArg[1:])
	CheckError(err)

	idNum, err := strconv.Atoi(myId)
	CheckError(err)

	num := portNum + idNum - 1

	newPortStr := strconv.Itoa(num)

	return ":" + newPortStr
}

func executeClockCycle() {
	logicalClock.Clock++
	fmt.Println("Executing clock cycle")
	fmt.Println("Current logical Clock: ", logicalClock.Clock)
}

func initConnections() error {
	ch  = make(chan string)

	timeOut = 3
	nonOtherServers := 2
	nServers = len(os.Args) - nonOtherServers
	wanting = false
	holding = false

	/* the 2 remove the name (Process) and remove the fist port, in the case it is my port */
	if nServers <= 0 {
		return errors.New("insufficient number of servers")
	}

	myId = os.Args[1]
	myPort = getMyPortNumber(os.Args[2], myId)

	idNum, err := strconv.Atoi(myId)
	CheckError(err)

	CliConn = make([]*net.UDPConn, nServers)
	logicalClock = ClockStruct{Id: idNum, Clock: 0, Message: "", Request: false, myTurn: false}
	allowedRequest = make([]bool, nServers)

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

	ServerAddr, err := net.ResolveUDPAddr("udp", ":10001")
	CheckError(err)

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)

	CSCliConn, err = net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)

	// init server
	ServerAddr, err = net.ResolveUDPAddr("udp", myPort)
	CheckError(err)

	ServerConn, err = net.ListenUDP("udp", ServerAddr)
	CheckError(err)

	return nil
}

func readInput(ch chan string) {
	/*
		Non-blocking async routine to listen for terminal input
	*/

	reader := bufio.NewReader(os.Stdin)
	for {
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}

}

func broadcastMessage(message string, request bool){

	idNum, err := strconv.Atoi(myId)
	CheckError(err)

	logicalClock.Message = message
	logicalClock.Request = request

	if request {
		allowedRequest = make([]bool, 5)
		allowedRequest[idNum - 1] = true
	}

	for i := 0; i < nServers; i++ {
		if i+1 != idNum {
			go doClientJob(i)
		}
	}

	executeClockCycle()
}

func answerBack(message string, receivedId int){
	logicalClock.Message = message
	logicalClock.Request = false
	go doClientJob(receivedId - 1)
}

func main() {
	e := initConnections()
	CheckError(e)

	/*  close all connections at the end of this function  */
	defer ServerConn.Close()
	for i := 0; i < nServers; i++ {
		defer CliConn[i].Close()
	}

	go readInput(ch)

	var valid bool
	for {
		go doServerJob()
		select {
			case messageToSend, valid = <-ch:
				if valid {
					fmt.Printf("Message: %s \n", messageToSend)
					broadcastMessage(messageToSend, true)
				} else {
					fmt.Println("Channel Closed!")
				}
			default:
				time.Sleep(time.Second * 1)
		}
	}
}
