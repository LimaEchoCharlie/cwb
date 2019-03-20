package main

import (
	pb "./syml"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

// server will implement the pb.SimpleServiceServer interface
type server struct{}

func (s *server) Snooze(ctx context.Context, in *pb.SnoozeRequest) (*pb.Empty, error) {
	fmt.Printf("snooze (%s) in:  %s\n", in.Id, time.Now().Format("15:04:05"))
	time.Sleep(time.Duration(in.Secs) * time.Second)
	fmt.Printf("snooze (%s) out: %s\n", in.Id, time.Now().Format("15:04:05"))
	return &pb.Empty{}, nil
}

func main() {
	addr := flag.String("addr", "localhost:9090", "Address to listen to")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSimpleServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
