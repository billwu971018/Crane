package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	qr "../query"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

//build a counter for counting num of lines from result
type counter struct {
	linenum  int
	filename string
}

const (
	port = "50055"
)

var (
	address = [...]string{"fa18-cs425-g61-01.cs.illinois.edu",
		"fa18-cs425-g61-02.cs.illinois.edu",
		"fa18-cs425-g61-03.cs.illinois.edu",
		"fa18-cs425-g61-04.cs.illinois.edu",
		"fa18-cs425-g61-05.cs.illinois.edu",
		"fa18-cs425-g61-06.cs.illinois.edu",
		"fa18-cs425-g61-07.cs.illinois.edu",
		"fa18-cs425-g61-08.cs.illinois.edu",
		"fa18-cs425-g61-09.cs.illinois.edu",
		"fa18-cs425-g61-10.cs.illinois.edu"}
	//address = [...]string{"127.0.0.1"} // uncomment for local debug
)

var wg sync.WaitGroup //define a wait group for sync

// define a global counter
// do not forget to assign channel size!
var gcount = make(chan counter, 0)

func main() {
	// init connection pool
	machine_num := len(address)
	var g_conn = make([]*grpc.ClientConn, machine_num)
	var g_c = make([]qr.LogQueryClient, machine_num)

	// init all address in your address pool
	for i, addr := range address {
		// concat addr and port making it looks like "fa18-cs425-g61-01.cs.illinois.edu:50051"
		addr = addr + ":" + port
		wg.Add(1) //add a work count
		go func(i int, addr string) {
			// Set up a connection to the server.
			// larger the upper boundary of recv and send size
			conn, err := grpc.Dial(addr, grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(256<<20),
				grpc.MaxCallSendMsgSize(256<<20)), grpc.WithInsecure())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			// create a client
			c := qr.NewLogQueryClient(conn)
			// save your connections and client into a save place
			g_conn[i] = conn
			g_c[i] = c
			wg.Done() //decrease a count to declare that 'this work is done'
		}(i, addr) // you must post all para you need into this func because this is a closure
	}
	wg.Wait() //join until all works are done
	for {
		// wait for user inputing
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter the string to grep: ")
		word, _ := reader.ReadString('\n')
		// start Distribute Querying
		DistributeQuery(g_c, word)
	}
}

/*
Design for distribute querying
param: 1. all client for sending requert to server 2. the word you want query
*/
func DistributeQuery(g_c []qr.LogQueryClient, word string) {
	gcount = make(chan counter, 10)
	for _, c := range g_c {
		wg.Add(1) //add a work count
		//create a corountine for each individual query server
		go HandleConn(c, word)
	}
	wg.Wait() //wait until all works are done
	// you must close the channel before iterate it
	close(gcount)

	// iterate channel and print the filename and the total num of results in each valid file
	fmt.Printf("\n")
	for count, ok := <-gcount; ok; {
		fmt.Printf("The file %s has %d results\n", count.filename, count.linenum)
		count, ok = <-gcount
	}
}

/*
Design for single querying
param: 1. all client for sending requert to server 2. the word you want query
return: No return, all result will be printed in real-time and sent into channel `gcount`
*/
func HandleConn(c qr.LogQueryClient, word string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r, err := c.ReturnResult(ctx, &qr.Query{Name: word})
	if err != nil {
		fmt.Printf("could not grep: %v", err)
		wg.Done() //decrease a count to declare that 'this work has done'
		return
	}
	fmt.Printf(r.Message)
	res := string(r.Message)
	if res != "" {
		// if the result is valid, calculate to total num of result found in this machine
		linenum := strings.Count(string(res), "\n") + 1
		filename := strings.SplitN(string(res), ":", 2)[0]
		a := counter{linenum, filename}
		// send the filename and the total num of results in each valid file into global channel `gcount`
		gcount <- a
	}

	wg.Done() //decrease a count to declare that 'this work has done'
}
