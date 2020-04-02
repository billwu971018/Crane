package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"../basic"
)

const (
	// INTRODUCEIP = "172.22.154.230"
	// INTRODUCEIP    = "192.168.50.179"
	INTRODUCEIP    = "172.22.154.202"
	debugIP        = "192.168.50.21"
	UDPPORT        = ":50060"
	TIMEOUT        = 500
	PINGTIME       = 300
	MESSAGETIMEOUT = 1000 * 30
)

var (
	memberMap    = make(map[string]string)
	localIP      = basic.GetIP()
	statues      = false //leave or join
	count        = make(map[string]int)
	isIntroducer = false
	isStandBy    = false

	isMaster   = false
	verbose    = true
	isClient   = false
	isOrder    = false
	is         = false
	fileMap    = make(map[string][]string)
	machineMap = make(map[string][]string)
	versionMap = make(map[string][]string)

	lsDone    = make(chan string)
	storeDone = make(chan string)

	masterIP = INTRODUCEIP

	FILELOCAITON = "../files/"
	orderAddr    = []string{}

	clientIP    = ""
	stopreceive = false

	STANDBYIP = ""
	result    = ""
)

func detectFail(targetIP string) {
	fmt.Println(basic.GetTime() + "\tdetect fail: " + targetIP)

	for desIP := range memberMap {
		send(desIP, "leave", targetIP, "")
	}

	send("127.0.0.1", "leave", targetIP, "")

	// if masterIP == targetIP {
	// 	fmt.Println("master leave! elect now!")
	// 	send(orderAddr[0], "election", localIP, "")
	// }
	if clientIP == targetIP {
		fmt.Println("client leave! elect now!")
		setTopo()
	}
}

func ping() {
	for {
		if statues {
			keys := basic.KeysView2(memberMap)
			basic.Shuffle(keys)
			for _, desIP := range keys {
				if desIP == localIP {
					continue
				}
				res := send(desIP, "ping", localIP, "")
				if !res { //ping failed
					if _, ok := memberMap[desIP]; ok {
						if count[desIP] > 1 {
							count[desIP] = 0
							detectFail(desIP)
						} else {
							count[desIP] = count[desIP] + 1
						}
					}
				} else { //ping successed
					count[desIP] = 0
				}
				time.Sleep(PINGTIME * time.Millisecond)
			}
		}
	}
}

func send(desIP string, cmd string, localIP string, message string) bool {
	var pingDone = make(chan bool)
	conn, err := net.Dial("udp", desIP+UDPPORT)
	defer conn.Close()

	timeStr := basic.GetTime()
	if err != nil {
		return false
	}

	sendStr := cmd + "@" + localIP + "@" + timeStr + "@" + message
	conn.Write([]byte(sendStr))

	if cmd == "ping" {
		// wait for ack in these cases
		go func() {
			byteArray := make([]byte, 256)
			conn.Read(byteArray)

			n := bytes.Index(byteArray, []byte{0})
			res := string(byteArray[:n])

			if res == "ack" {
				pingDone <- true
			}
		}()

		select {
		case <-time.After(TIMEOUT * time.Millisecond):
			return false
		case <-pingDone:
			return true
		}
	}

	// if cmd != "ping" && cmd != "register" {
	// 	fmt.Println("send: " + sendStr)

	// }
	return true

}

func listenUDP() {
	udp_addr, _ := net.ResolveUDPAddr("udp", UDPPORT)

	conn, _ := net.ListenUDP("udp", udp_addr)
	//defer conn.Close()

	for {
		// read cmd message
		byteArray := make([]byte, 256)
		_, raddr, _ := conn.ReadFromUDP(byteArray[0:])
		n := bytes.Index(byteArray, []byte{0})
		res := string(byteArray[:n])
		cmd, ip, _, message := basic.SplitLink(res)

		if cmd != "ping" && cmd != "register" {
			fmt.Println("receive: " + res)
		}
		switch cmd {
		case "ping": // ping
			conn.WriteToUDP([]byte("ack"), raddr)
		case "intro": //intro
			if _, ok := memberMap[ip]; ok {
				continue
			}
			keys := make([]string, 0, len(memberMap))
			for desIP := range memberMap {
				send(desIP, "join", ip, "")
				keys = append(keys, desIP)
			}
			memberList := strings.Join(keys, "%")
			send(ip, "setlist", localIP, memberList)
			timeStr := basic.GetTime()
			memberMap[ip] = timeStr
			fmt.Println(timeStr + "\tjoin: " + ip)

			basic.ShowList(memberMap)

		case "join": //join
			t := basic.GetTime()
			fmt.Println(t + "\tjoin: " + ip)
			memberMap[ip] = t
			count[ip] = 0
			basic.ShowList(memberMap)
		case "leave": //leave
			fmt.Println(basic.GetTime() + "\tleave: " + ip)
			delete(memberMap, ip)
			basic.ShowList(memberMap)
			if isMaster {
				fmt.Println("reassign replica" + ip)
				ReassignMachine(ip)
				delete(machineMap, ip)
			}

		case "setlist": //initialize membership list
			timeStr := basic.GetTime()
			fmt.Println(timeStr + "\tinitialize list: " + ip)

			memberList := strings.Split(message, "%")
			for _, member := range memberList {
				memberMap[member] = timeStr
			}

			memberMap[INTRODUCEIP] = timeStr
			fmt.Println(timeStr + "\tjoin successed")
			basic.ShowList(memberMap)
			statues = true

		case "register": //register
			if _, ok := memberMap[ip]; !ok {
				memberMap[ip] = basic.GetTime()
				basic.ShowList(memberMap)
			}
		case "download": // only useful for master
			sdfsfilename := strings.Split(message, "%")[1]
			storedIPs := fileMap[sdfsfilename]
			if len(storedIPs) > 0 {
				send(storedIPs[0], "downloadForward", localIP, message+"%"+ip)
				send(ip, "downloadReturn", localIP, "true")
			} else {
				send(ip, "downloadReturn", localIP, "false")
			}
		case "downloadForward":
			res := strings.Split(message, "%")
			localfilename := res[0]
			sdfsfilename := res[1]
			version := res[2]
			dstIP := res[3]
			version_i, _ := strconv.Atoi(version)
			verisonfilename := versionMap[sdfsfilename][version_i]
			if dstIP == localIP {
				dstIP = "127.0.0.1"
			}
			SendFile(dstIP, verisonfilename, localfilename)

		case "upload":
			res := strings.Split(message, "%")
			sdfsfilename := res[1]
			candidates := uploadCandidates(sdfsfilename)
			var wg sync.WaitGroup
			for _, member := range candidates {
				wg.Add(1) //add a work count
				go func(member string) {
					send(member, "giveyoufile", localIP, message)
					wg.Done()
				}(member)
				wg.Wait() //wait until all works are done
			}

		case "giveyoufile":
			res := strings.Split(message, "%")
			localfilename := res[0]
			sdfsfilename := res[1]
			dstIP := res[2]
			versionFilename := vesionControl(sdfsfilename)

			send(dstIP, "request", localIP, localfilename+"%"+versionFilename)

		case "request":
			res := strings.Split(message, "%")
			localfilename := res[0]
			versionFilename := res[1]
			SendFile(ip, localfilename, versionFilename)

		case "delete":
			// if isMaster {
			for k, v := range machineMap {
				machineMap[k] = basic.Delete(v, message)
			}
			delete(fileMap, message)

			// } else {
			delete(versionMap, message)
			for _, file := range versionMap[message] {
				os.Remove(FILELOCAITON + file)
			}
			// }

		case "ls":
			targetIPs := fileMap[message]
			send(ip, "lsReturn", localIP, strings.Join(targetIPs, "\n"))
		case "store":
			storedFiles := machineMap[message]
			send(ip, "storeReturn", localIP, strings.Join(storedFiles, "\n"))

		case "lsReturn":
			lsDone <- message

		case "storeReturn":
			storeDone <- message

		case "downloadReturn":
			if message == "false" {
				fmt.Println("not accessable!")
			} else {
				fmt.Println("download success!")

			}
		case "restore":
			res := strings.Split(message, "%")
			file := res[0]
			versionList := []string{}
			if len(res) > 1 {
				versionList = res[1:]
			}
			versionMap[file] = versionList

		case "helpit":
			res := strings.Split(message, "%")
			newRepo := res[0]
			file := res[1]
			send(newRepo, "restore", localIP, file+"%"+strings.Join(versionMap[file], "%"))
			for _, file := range versionMap[file] {
				SendFile(newRepo, file, file)
			}
		case "newmaster":
			masterIP = ip
			if masterIP == localIP {
				fmt.Println("I'm newmaster" + masterIP)
				isMaster = true
			} else {
				fmt.Println("get the newmaster" + masterIP)

			}
		case "orderIP":
			orderAddr = strings.Split(message, "%")
			for _, item := range orderAddr {
				if item == localIP {
					isOrder = true
				} else {
					isOrder = false
				}
			}
		// case "election":
		// 	election()
		case "retriveorder":
			isOrder = true
			machineIdx := strings.Index(message, "machineMap\n")
			fileStr := res[9:machineIdx]
			machineStr := res[machineIdx+11:]
			//file
			fileMap = makeMap(fileStr)
			machineMap = makeMap(machineStr)
		case "clientIP":
			if message != localIP {
				clientIP = message
				fmt.Println("client:" + clientIP)
			} else {
				isClient = true
				fmt.Println("I'm client")
			}
		case "STANDBYIP":
			STANDBYIP = message
			if message != localIP {
				fmt.Println("STANDBYIP :" + STANDBYIP)
			} else {
				isStandBy = true
				fmt.Println("I'm STANDBYIP")
				go listenMaster()

			}
		case "setSlices":
			totalNum, _ := strconv.Atoi(message)
			go secondLevel(totalNum)
		}

	}
}

func listenIntroducer() {
	for {
		if statues && !isIntroducer {
			send(INTRODUCEIP, "register", localIP, "")
		}
		time.Sleep(2 * time.Second)
	}
}

func main() {
	go listenUDP()
	go ping()
	go listenIntroducer()
	go SendOrder()
	go firstLevel()
	go keepMaster()

	//go func(){
	//	time.Sleep(10* time.Second)
	//election()
	//
	//}()
	//basic.CleanDir(FILELOCAITON)
	exec.Command("/bin/sh", "-c", "rm -rf ../files/")
	exec.Command("/bin/sh", "-c", "mkdir ../files/")

	if localIP == INTRODUCEIP {
		//time.Sleep(5 * time.Second)
		fmt.Println("I'm introducer")
		isIntroducer = true
		// isMaster = true
		statues = true
	} else {
		fmt.Println("I'm a common people")
		send(INTRODUCEIP, "intro", localIP, "")
		statues = true
	}
	// if localIP == debugIP {
	// 	time.Sleep(5 * time.Second)
	// 	listenMaster()
	// 	isStandBy = true
	// }
	for {
		// wait for user inputing
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter the command(" + localIP + "): ")
		rawInput, _ := reader.ReadString('\n')
		args := strings.Split(strings.TrimSpace(rawInput), " ")
		cmd := args[0]

		switch cmd {
		case "start":
			setTopo()
			election()
			setStandBy()
			zeroLevel(0)
			fmt.Println("ok")
		case "join":
			send(INTRODUCEIP, "intro", localIP, "")
			statues = true

		case "leave":
			for desIP := range memberMap {
				send(desIP, "leave", localIP, "")
			}
			// clean map
			memberMap = make(map[string]string)
			fmt.Println(basic.GetTime() + "\tleave group")
			statues = false

		case "introducer":
			isIntroducer = true
			// isMaster = true
			statues = true

		case "show":
			basic.ShowList(memberMap)
		case "show-f":
			basic.ShowMap(fileMap)
		case "show-m":
			basic.ShowMap(machineMap)
		case "show-v":
			basic.ShowMap(versionMap)

		case "master":
			// isMaster = true
			fmt.Println(masterIP)

		case "put":
			localfilename := args[1]
			sdfsfilename := args[2]
			send(masterIP, "upload", localIP, localfilename+"%"+sdfsfilename+"%"+localIP)

		case "get":
			localfilename := args[1]
			sdfsfilename := args[2]
			send(masterIP, "download", localIP, localfilename+"%"+sdfsfilename+"%0")

		case "get-versions":
			localfilename := args[1]
			sdfsfilename := args[2]
			numVersions := args[3]
			send(masterIP, "download", localIP, localfilename+"%"+sdfsfilename+"%"+numVersions)

		case "delete":
			sdfsfilename := args[1]
			for ip := range memberMap {
				send(ip, "delete", localIP, sdfsfilename)
			}

		case "ls":
			sdfsfilename := args[1]
			res := List(sdfsfilename)
			fmt.Println(res)

		case "election":
			election()
		case "set":
			setTopo()
		case "store":
			dstIP := args[1]
			res := Store(dstIP)
			fmt.Println(res)

		case "verbose":
			if verbose == true {
				verbose = false
			} else {
				verbose = true
			}

		case "stopreceive":
			stopreceive = true
		case "standby":
			go listenMaster()
		case "result":
			fmt.Println(result)
		case "save":
			localfilename := basic.GetTimeStamp() + ".log"
			sdfsfilename := localfilename
			basic.WriteToFile(FILELOCAITON+basic.GetTimeStamp(), result)
			// fmt.Println(count_str)
			send(masterIP, "upload", localIP, localfilename+"%"+sdfsfilename+"%"+localIP)
		default:
			fmt.Println("invalid command!")
		}

	}
}
