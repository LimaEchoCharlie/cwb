package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"image"
	"log"
	"syml"
	"sync"
	"time"
	"google.golang.org/grpc/credentials"
	"crypto/tls"
)

var defaultContext = context.Background()

func mustDial(addr string) *grpc.ClientConn {
	// set InsecureSkipVerify to true so that TLS accepts any certificate presented by the server
	cfg := &tls.Config{InsecureSkipVerify: true}
	// connect to the server
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(credentials.NewTLS(cfg)))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	return conn
}

func runFullWorkflow(client syml.SimpleServiceClient) (err error) {
	const ID = "holl"
	fmt.Println("run custom command")
	var cmdResponse *syml.CommandResponse
	b, _ := json.Marshal(image.Rect(1, 2, 3, 5))
	if cmdResponse, err = client.CustomCommand(defaultContext, &syml.CommandRequest{Id: ID, Name: "area", Parameters: b}); err != nil {
		return err
	}
	fmt.Println(cmdResponse)

	fmt.Println("run custom command with unexpected command name")
	_, expectedErr := client.CustomCommand(defaultContext, &syml.CommandRequest{Id: ID, Name: "wrong", Parameters: b})
	fmt.Println(expectedErr)

	fmt.Println("run snooze")
	if _, err = client.Snooze(defaultContext, &syml.SnoozeRequest{Id: ID, Secs: 2}); err != nil {
		return err
	}
	return nil

}

func main() {
	addr := flag.String("addr", "localhost:9090", "Address to listen to")
	nClients := flag.Int("multi", 0, "Number of clients")
	flag.Parse()

	var wg sync.WaitGroup
	wg.Add(*nClients)
	if *nClients == 0 {
		conn := mustDial(*addr)
		defer conn.Close()
		// create a client and call snooze
		client := syml.NewSimpleServiceClient(conn)
		err := runFullWorkflow(client)
		if err != nil {
			log.Fatalf("could not run full workflow: %v", err)
		}
	} else {
		for i := 0; i < *nClients; i++ {

			go func(i int) {
				defer wg.Done()
				conn := mustDial(*addr)
				defer conn.Close()
				// create a client and call snooze
				client := syml.NewSimpleServiceClient(conn)
				_, err := client.Snooze(defaultContext, &syml.SnoozeRequest{Id: fmt.Sprint(i), Secs: 10})
				if err != nil {
					log.Fatalf("could not snooze: %v", err)
				}
			}(i)
			time.Sleep(time.Second)
		}
		wg.Wait()
	}
}
