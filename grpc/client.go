package main

import (
	pb "./syml"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"sync"
	"time"
)

func main() {
	addr := flag.String("addr", "localhost:9090", "Address to listen to")
	nClients := flag.Int("multi", 0, "Number of clients")
	flag.Parse()

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(*nClients)
	for i := 0; i < *nClients; i++ {
		go func(i int) {
			defer wg.Done()
			// connect to the server
			conn, err := grpc.Dial(*addr, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			defer conn.Close()
			// create a client and call snooze
			client := pb.NewSimpleServiceClient(conn)
			_, err = client.Snooze(ctx, &pb.SnoozeRequest{Id: fmt.Sprint(i), Secs: 10})
			if err != nil {
				log.Fatalf("could not greet: %v", err)
			}
		}(i)
		time.Sleep(time.Second)
	}
	wg.Wait()
}
