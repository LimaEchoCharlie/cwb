package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"image"
	"log"
	"net"
	"syml"
	"time"
	"io/ioutil"
	"crypto/x509"
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
	fmt.Printf("custom command (%s) name:  %s\n", in.Id, in.Name)
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

	// load server certificate and key
	serverCert, err := tls.LoadX509KeyPair("testdata/server-cert.pem", "testdata/server-key.pem")
	if err != nil {
		log.Fatalf("failed to load server cert: %v", err)
	}

	// load client certificate and add to certificate pool
	clientBytes, err := ioutil.ReadFile("testdata/client-cert.pem")
	if err != nil {
		log.Fatalf("failed to read client cert: %v", err)
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(clientBytes)
	if !ok {
		fmt.Errorf("failed to append cert from PEM")
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert, // set the server's policy for TLS Client Authentication
		ClientCAs:    certPool,
	}
	s := grpc.NewServer(grpc.Creds(credentials.NewTLS(cfg)))
	syml.RegisterSimpleServiceServer(s, &server{})

	fmt.Println("Starting the gRPC simple server... on ", *addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
