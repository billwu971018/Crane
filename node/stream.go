package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	//for application 2 and 3
	//"sort"

	"../basic"
	set "github.com/deckarep/golang-set"
)

func setTopo() {
	//new master will be the one with minimum IP address
	minIP := ""
	memberList := basic.KeysView2(memberMap)
	for _, currIP := range memberList {
		if (currIP < minIP && currIP != "") || minIP == "" {
			minIP = currIP
		}
	}
	clientIP = minIP
	fmt.Println("client" + clientIP)

	//broadcasting
	for desIP := range memberMap {
		if desIP != localIP {
			send(desIP, "clientIP", localIP, clientIP)
		}
	}
}

func zeroLevel(progress int) {
	standbycon, _ := net.Dial("udp", STANDBYIP+":1350")

	s := localFileReader()
	// numOfMachines:=len(memberMap)-2
	var taskMachines []string
	for ip := range memberMap {
		if ip != clientIP {
			taskMachines = append(taskMachines, ip)
		}
	}
	numOfMachines := len(taskMachines)
	numOftasks := len(s)
	ceil := int(math.Ceil(float64(numOftasks) / float64(numOfMachines)))
	if progress == 0 {
		send(clientIP, "setSlices", localIP, strconv.Itoa(numOftasks))
		fmt.Println("new progress!")
	} else {
		fmt.Println("continune!" + strconv.Itoa(progress))
	}

	writeConns := []net.Conn{}
	for j := 0; j < numOfMachines; j++ {
		fmt.Println(taskMachines[j])
		con, _ := net.Dial("tcp", taskMachines[j]+":1234")
		writeConns = append(writeConns, con)
	}
	var wg sync.WaitGroup

	for i := 0; i < ceil; i++ {
		for j := 0; j < numOfMachines; j++ {
			if i*numOfMachines+j < numOftasks {
				if i*numOfMachines+j < progress {
					continue
				}
				for n := 0; n < 3; n++ {
					wg.Add(1)
					go func(content string, con net.Conn, idx int, j int) {
						wg.Done()
						str := basic.Lpad(strconv.Itoa(idx), "0", 20)
						con.Write([]byte(str + content))
						// fmt.Println(str + content)

					}(s[i*numOfMachines+j], writeConns[(j+n)%numOfMachines], i*numOfMachines+j, j)
				}
				wg.Wait()
				// send current progress
				// fmt.Println("progress:" + strconv.Itoa(i*numOfMachines+j))
				standbycon.Write([]byte(strconv.Itoa(i*numOfMachines + j)))
				// time.Sleep(500 * time.Millisecond)

			}
		}
	}
	fmt.Println("finished!")

}

func firstLevel() {
	//udp_addr, _ := net.ResolveUDPAddr("udp", ":1234")
	ln, _ := net.Listen("tcp", ":1234")
	saveIP := clientIP
	writeConn, _ := net.Dial("tcp", clientIP+":1235")
	// var wg sync.WaitGroup
	for {
		ReadCon, _ := ln.Accept()
		defer ReadCon.Close()
		byteArray := make([]byte, 65535)
		ReadCon.Read(byteArray[0:])
		// fmt.Println("received")
		// wg.Add(1)
		if clientIP != saveIP {
			saveIP = clientIP
			writeConn, _ = net.Dial("tcp", clientIP+":1235")
		}

		// go func() {
		n := bytes.Index(byteArray, []byte{0})
		received := string(byteArray[:n])
		idx := received[:20]
		line := received[20:]
		res := wordcount(line)
		fmt.Println(idx + ":" + res + ":" + line)

		writeConn.Write([]byte(idx + res))
		// 	wg.Done()
		// }()
		// wg.Wait()
	}
}

func secondLevel(totalNum int) {
	//udp_addr, _ := net.ResolveUDPAddr("udp", ":1235")
	ln, _ := net.Listen("tcp", ":1235")
	count := 0
	finished_idx := set.NewSet()
	// var wg sync.WaitGroup
	fmt.Println("start!")
	for {
		byteArray := make([]byte, 65535)
		ReadCon, _ := ln.Accept()
		defer ReadCon.Close()
		ReadCon.Read(byteArray[0:])

		// wg.Add(1)

		// go func() {
		n := bytes.Index(byteArray, []byte{0})
		received := string(byteArray[:n])
		idx := received[:20]
		line := received[20:]
		res_int, _ := strconv.Atoi(line)
		fmt.Println(idx + ":" + line)

		if !finished_idx.Contains(idx) {
			finished_idx.Add(idx)
			count = count + res_int
			count_str := strconv.Itoa(count)
			result = count_str
			fmt.Println(count_str)
		}
		// wg.Done()
		// }()
		// wg.Wait()
	}

}

//application 1
func wordcount(s string) string {
	return strconv.Itoa(len(strings.Split(s, " ")))
}

/**
//application 2
func reddit(s[][] string) string{
	m := make(map[[]string]int)
	for _, curr := range s{
		m[curr] = strconv.Atoi(curr[11])
	}
	type kv struct {
        Key   []string
        Value int
    }
	var ss []kv
  for k, v := range m {
        ss = append(ss, kv{k, v})
    }

    sort.Slice(ss, func(i, j int) bool {
        return ss[i].Value > ss[j].Value
    })
	var ret [][]string
	for i, val := range kv{
		if i == 100{
			break
		}
		append(ret, val)
	}
	return ret
}
**/

/**
//application 3
func twitter(s[][] string) string{
	m := make(map[string]int)
	for _, curr : range s{
		m[curr[4]] = m[curr[4]]+1
	}
	return m
}
**/

func localFileReader() []string {
	inputFile := "../input.txt"
	//inputFile := "../input.csv"
	if localIP == debugIP {
		inputFile = "./input.txt"
		//inputFile = "./input.csv"
	}
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	var s []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		s = append(s, line)
	}
	return s

}
