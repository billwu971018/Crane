package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"time"

	"../basic"
)

func setStandBy() {
	//new master will be the one with minimum IP address
	maxIP := ""
	memberList := basic.KeysView2(memberMap)
	for _, currIP := range memberList {
		if currIP < maxIP && currIP != "" && currIP != clientIP && currIP > localIP {
			maxIP = currIP
		}
	}
	STANDBYIP = maxIP
	if STANDBYIP == "" {
		STANDBYIP = memberList[0]
	}

	//STANDBYIP = debugIP
	fmt.Println("STANDBYIP " + STANDBYIP)

	//broadcasting
	for desIP := range memberMap {
		if desIP != localIP {
			send(desIP, "STANDBYIP", localIP, STANDBYIP)
		}
	}
}

func listenMaster() {
	isStandBy = true
	fmt.Println("listenMaster")
	udp_addr, _ := net.ResolveUDPAddr("udp", ":1350")
	conn, _ := net.ListenUDP("udp", udp_addr)
	defer conn.Close()
	progressLocal := 0
	ch := make(chan int)
	chr := make(chan bool)
	// Start a goroutine to read from our net connection
	go func(ch chan int, chr chan bool) {
		for {
			// try to read the data
			data := make([]byte, 512)
			conn.ReadFromUDP(data)

			n := bytes.Index(data, []byte{0})
			data_str := string(data[:n])
			if data_str == "ok" {
				chr <- true
				continue
			}
			fmt.Println(progressLocal)
			data_i, _ := strconv.Atoi(data_str)
			// send data if we read some.
			ch <- data_i
		}
	}(ch, chr)

	// continuously read from the connection
	for {
		ticker := time.Tick(time.Second)
		select {
		// This case means we recieved data on the connection
		case progressLocal = <-ch:
		case <-chr:
			// Do something with the data
		// This case means we got an error and the goroutine has finished

		// This will timeout on the read.
		case <-ticker:
			fmt.Println("Master leave!")
			// set master
			election()
			setStandBy()
			zeroLevel(progressLocal)
			return
		}
	}
}

func keepMaster() {
	saveStandBy := STANDBYIP
	conn, _ := net.Dial("udp", STANDBYIP+":1350")

	for {
		if STANDBYIP != saveStandBy {
			saveStandBy = STANDBYIP
			conn, _ = net.Dial("udp", STANDBYIP+":1350")
		}
		if saveStandBy != "" {
			conn.Write([]byte("ok"))
		}
		time.Sleep(100 * time.Millisecond)
	}
}
