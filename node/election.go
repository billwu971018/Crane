package main

import (
	"fmt"
	"sort"
	"strings"

	"../basic"
)

func makeMap(msg string) map[string][]string {
	ret := make(map[string][]string)
	lastIdx := 0
	for {
		keyIdx := strings.Index(msg[lastIdx:len(msg)], "@")
		currKey := msg[lastIdx:keyIdx]
		var values []string
		lastValueIdx := keyIdx + 1
		for {
			valueIdx := strings.Index(msg[lastValueIdx:len(msg)], "#")
			if valueIdx > keyIdx {
				break
			}
			value := msg[lastValueIdx:valueIdx]
			values = append(values, value)
			lastValueIdx = valueIdx + 1
			if lastValueIdx == len(msg) {
				break
			}
		}
		lastIdx = lastValueIdx + 1
		if lastIdx >= len(msg) {
			break
		}
		ret[currKey] = values
	}
	return ret
}

func mapStr(currMap map[string][]string) string {
	ret := ""
	for k, v := range currMap {
		ret = ret + k + "@"
		value := ""
		for _, val := range v {
			value = value + val + "#"
		}
		ret = ret + value
	}
	return ret
}

func chooseRepIP() {
	memberList := basic.KeysView2(memberMap)
	sort.Strings(memberList)
	orderAddr = []string{}
	if len(memberList) < 3 {
		orderAddr = memberList
	} else {
		orderAddr = memberList[:3]
	}
	//broadcasting
	for currIP, _ := range memberMap {
		send(currIP, "orderIP", localIP, strings.Join(orderAddr, "%"))
	}
}

func SendOrder() {
	fileStr := mapStr(fileMap)
	machineStr := mapStr(machineMap)
	text := "fileMap\n" + fileStr + "machineMap\n" + machineStr

	for _, member := range orderAddr {
		send(member, "retrieveorder", localIP, text)
	}
}
func election() {
	//new master will be the one with minimum IP address
	// maxIP := ""
	// memberList := basic.KeysView2(memberMap)
	// for _, currIP := range memberList {
	// 	if currIP > maxIP && currIP != "" {
	// 		maxIP = currIP
	// 	}
	// }
	masterIP = localIP
	fmt.Println("new master" + masterIP)

	//broadcasting
	for desIP, _ := range memberMap {
		if desIP != localIP {
			send(desIP, "newmaster", masterIP, "")
		}
	}
	chooseRepIP()
	SendOrder()
	fileStr := mapStr(fileMap)
	machineStr := mapStr(machineMap)
	text := "fileMap\n" + fileStr + "machineMap\n" + machineStr
	send(masterIP, "retrieveorder", localIP, text)

}
