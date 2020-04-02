package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"../basic"
)

// implement upload function
func vesionControl(sdfsfilename string) string {
	// maintain deque
	timestampe := basic.GetTimeStamp()
	versionFilename := timestampe + "#" + sdfsfilename
	// only 5 backup allowed
	if len(versionMap[sdfsfilename]) >= 5 {
		os.Remove(FILELOCAITON + versionMap[sdfsfilename][4])
		versionMap[sdfsfilename] = append([]string{versionFilename}, versionMap[sdfsfilename][:4]...)
	} else {
		versionMap[sdfsfilename] = append([]string{versionFilename}, versionMap[sdfsfilename][:]...)
	}
	return versionFilename
}

func uploadCandidates(sdfsfilename string) []string {
	// forward file
	members := []string{}
	if _, ok := fileMap[sdfsfilename]; ok {
		// file exist
		members = fileMap[sdfsfilename]
	} else {
		// new file
		keys := basic.KeysView2(memberMap)
		if len(keys) >= 4 {
			members = basic.Shuffle(keys)[:4]
		} else {
			members = basic.Shuffle(keys)
		}

		// upload into 4 machines
		for _, memberIP := range members {
			if !basic.Contains(machineMap[memberIP], sdfsfilename) {
				machineMap[memberIP] = append(machineMap[memberIP], sdfsfilename)
			}
			if !basic.Contains(fileMap[sdfsfilename], memberIP) {
				fileMap[sdfsfilename] = append(fileMap[sdfsfilename], memberIP)
			}
			fmt.Println("transfer upload file " + sdfsfilename + " to " + memberIP)
		}

	}

	return members

}

// implement download function
func SendFile(dstIP, srcfile, dstfile string) {

	// fill out the params into your template
	cmd_str := "scp -i /root/.ssh/id_rsa " + FILELOCAITON + srcfile + "  root@" + dstIP + ":/home/mp4/files/" + dstfile
	fmt.Println(cmd_str)

	// execute command
	cmd := exec.Command("/bin/sh", "-c", cmd_str)
	output := new(bytes.Buffer)
	cmd.Stdout = output
	cmd.Stderr = output
	err := cmd.Start()
	if err != nil {
		fmt.Println("Failed to start " + err.Error())
	}
}
