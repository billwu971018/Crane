package main

import (
	"time"

	"../basic"
	"github.com/adam-hanna/arrayOperations"
)

func ReassignMachine(ip string) {
	files := machineMap[ip]
	allIPS := basic.KeysView2(memberMap)
	for _, file := range files {
		fileMap[file] = basic.Delete(fileMap[file], ip)
		backupIPs := fileMap[file]
		diff, _ := arrayOperations.Difference(backupIPs, allIPS)
		diffSlice :=  diff.Interface().([]string)
		if len(diffSlice) != 0 {
			newRepo := basic.Shuffle(diffSlice)[0]
			send(backupIPs[0], "helpit", localIP, newRepo+"%"+file)
		}

	}
}

func List(sdfsfilename string) string {
	send(masterIP, "ls", localIP, sdfsfilename)
	select {
	case <-time.After(MESSAGETIMEOUT * time.Millisecond):
		return "Timeout!"

	case ls := <-lsDone:
		return ls
	}
}
func Store(dstIP string) string {
	send(masterIP, "store", localIP, dstIP)

	select {
	case <-time.After(MESSAGETIMEOUT * time.Millisecond):
		return "Timeout!"
	case store := <-storeDone:
		return store
	}
}
