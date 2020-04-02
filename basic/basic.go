package basic

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

var logFile, _ = os.Create("./mp3.log")
var logger = log.New(logFile, "", 0)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func SplitLink(s string) (string, string, string, string) {
	x := strings.Split(s, "@")
	return x[0], x[1], x[2], x[3]
}

// get locate ip
func GetIP() string {
	addrs, err := net.InterfaceAddrs()
	var ip = ""
	if err != nil {
		fmt.Println(err)
	}
	for _, address := range addrs {
		// check whether the ip is loop address
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				break
			}
		}

	}
	if ip == "" {
		fmt.Println("fail to find")
	}
	return ip
}

func GetTime() string {
	t := time.Now()
	timeStr := t.Format("2006-01-0215:04:05")

	return timeStr
}

func GetTimeStamp() string {
	t := time.Now()
	timeStr := t.Format("20060102150405")

	return timeStr
}

func Println(str string) {
	logger.Println(str)
	fmt.Println(str)
}
func LoggerPrintln(str string) {
	logger.Println(str)
}
func ShowList(memberMap map[string]string) {
	fmt.Println("")
	fmt.Println("##############showList##############")

	for desIP, theTime := range memberMap {
		fmt.Println("ip: " + desIP + "\tadd time: " + theTime)
	}
	fmt.Println("####################################")
	fmt.Println("")
}

func ShowMap(memberMap map[string][]string) {
	fmt.Println("")
	fmt.Println("##############showMap##############")

	for k, v := range memberMap {
		fmt.Println(k + ":")
		for _, item := range v {
			fmt.Println(item)
		}
		fmt.Println("")
	}
	fmt.Println("####################################")
	fmt.Println("")
}

func Shuffle(a []string) []string {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
	return a
}

func ReadLocalFile() []string {
	var files []string

	root := "/tmp/"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}
func contains(arr [3]string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func KeysView1(m map[string][]string) []string {
	keys := reflect.ValueOf(m).MapKeys()
	strkeys := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		strkeys[i] = keys[i].String()
	}
	return strkeys
}

func KeysView2(m map[string]string) []string {
	keys := reflect.ValueOf(m).MapKeys()
	strkeys := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		strkeys[i] = keys[i].String()
	}
	return strkeys
}

func Intersection(a, b []string) (c []string) {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			c = append(c, item)
		}
	}
	return c
}

func Difference(slice1 []string, slice2 []string) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

func Serilaize(m []string) io.Reader {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)

	// Encoding the map
	e.Encode(m)
	return b
}

func Deserilaize(b io.Reader) map[string]string {
	decodedMap := map[string]string{}
	d := gob.NewDecoder(b)

	// Decoding the serialized data
	d.Decode(&decodedMap)
	return decodedMap
}

func CleanDir(dir string) {
	os.RemoveAll(dir)
	os.Mkdir(dir, os.ModeDir)
}
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Delete(s []string, e string) []string {
	var na []string
	for _, v := range s {
		if v == e {
			continue
		} else {
			na = append(na, v)
		}
	}
	return na
}
func Lpad(s string, pad string, plength int) string {
	for i := len(s); i < plength; i++ {
		s = pad + s
	}
	return s
}

func WriteToFile(filename string, content string) {

	file, err := os.Create(filename) // Truncates if file already exists, be careful!
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer file.Close() // Make sure to close the file when you're done

	_, err = file.WriteString(content)

	if err != nil {
		log.Fatalf("failed writing to file: %s", err)
	}

}
