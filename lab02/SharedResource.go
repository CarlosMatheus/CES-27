package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

type ClockStruct struct {
	Id     int
	Clock  int
	Message string
	Request bool
}

var err string
var ServerConn *net.UDPConn
var ch chan string

func CheckError(err error) {
	/*
		Simple function to verify error
	*/
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func doServerJob() {
	buf := make([]byte, 1024)

	n, _, err := ServerConn.ReadFromUDP(buf[0:])
	CheckError(err)

	var logicalClockReceived ClockStruct
	err = json.Unmarshal(buf[:n], &logicalClockReceived)
	CheckError(err)

	fmt.Println("Received logical Clock:", logicalClockReceived.Clock)
	fmt.Println("Received", logicalClockReceived.Message)

}

func initConnections() error {
	/*
		init shared resources server
	 */
	ch  = make(chan string)

	ServerAddr, err := net.ResolveUDPAddr("udp", ":10001")
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

func main() {
	e := initConnections()

	if e != nil {
		fmt.Println(e)
		return
	}

	defer ServerConn.Close()

	go readInput(ch)

	for {
		go doServerJob()
		select {
			case x, valid := <-ch:
				if valid {
					fmt.Printf("Destiny Id: %s \n", x)
				} else {
					fmt.Println("Channel Closed!")
				}
			default:
				time.Sleep(time.Second * 1)
		}
	}
}
