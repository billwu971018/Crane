package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"

	qr "../query"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50055"
)

// server is used to implement QueryServer.
type server struct{}

// Implemented ReturnResult method for LogQuery Server
func (s *server) ReturnResult(ctx context.Context, in *qr.Query) (*qr.Reply, error) {
	// trim the work to be queried
	word := strings.TrimSpace(in.Name)

	// create local command template for grep
	// define the location of log files
	log_location := "/home/mp2/node/"

	// fill out the params into your template
	cmd_str := "find " + log_location + "  -name '*.log' -exec grep -Hn " + word + " {} \\;"
	fmt.Println(cmd_str)

	// execute command
	cmd := exec.Command("/bin/sh", "-c", cmd_str)
	output := new(bytes.Buffer)
	cmd.Stdout = output
	cmd.Stderr = output
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Failed to start " + err.Error())
	}

	// wait for result, when the work is done, there will be signal sent to channel `done`
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()
	// Wait for the process to finish or kill it when client disconnect
	select {
	case <-ctx.Done():
		log.Print("writer closed")
		if err := cmd.Process.Kill(); err != nil {
			log.Fatal("failed to kill process: ", err)
		}
		log.Println("process killed as client disconnect")
	case err := <-done:
		if err != nil {
			log.Fatalf("process finished with error = %v", err)
		}
		log.Print("process finished successfully")
	}

	res := output.String()

	return &qr.Reply{Message: res}, err
}

/*
main func
*/
func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	qr.RegisterLogQueryServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
