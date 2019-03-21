package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"image"
	"log"
	"net"
	"syml"
	"time"
)

// server will implement the syml.SimpleServiceServer interface
type server struct{}

func (s *server) Snooze(ctx context.Context, in *syml.SnoozeRequest) (*syml.Empty, error) {
	fmt.Printf("snooze (%s) in:  %s\n", in.Id, time.Now().Format("15:04:05"))
	time.Sleep(time.Duration(in.Secs) * time.Second)
	fmt.Printf("snooze (%s) out: %s\n", in.Id, time.Now().Format("15:04:05"))
	return &syml.Empty{}, nil
}

func (s *server) CustomCommand(ctx context.Context, in *syml.CommandRequest) (*syml.CommandResponse, error) {
	response := new(syml.CommandResponse)
	if in.Name != "area" {
		return response, fmt.Errorf("Unexpected command name \"%s\"", in.Name)
	}
	var rect image.Rectangle
	if err := json.Unmarshal(in.Parameters, &rect); err != nil {
		return response, err
	}
	response.Message = fmt.Sprintf("The area of the rectangle is %d", rect.Dx()*rect.Dy())
	return response, nil
}

func main() {
	addr := flag.String("addr", "localhost:9090", "Address to listen to")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	syml.RegisterSimpleServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
