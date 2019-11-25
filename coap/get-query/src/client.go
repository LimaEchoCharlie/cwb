package main

import (
	"github.com/go-ocf/go-coap"
	"log"
	"os"
	"bufio"
	"time"
	"context"
)

func main() {
	co, err := coap.Dial("udp", "localhost:5688")
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}
	log.Println("Type messages to send, enter 'q' to exit.")
	scanner := bufio.NewScanner(os.Stdin)
	var ctx context.Context
	var cancel context.CancelFunc = nil
	for scanner.Scan() {
		if cancel != nil {
			cancel()
		}
		if scanner.Text() == "q" {
			break
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)

		message, err := co.NewGetRequest("/reverse")
		if err != nil {
			log.Printf("Error creating request: %v", err)
			continue
		}
		log.Printf("Sending: \"%s\"", scanner.Text())
		message.SetQueryString(scanner.Text())

		response, err := co.ExchangeWithContext(ctx, message)
		if err != nil {
			log.Printf("Error sending request: %v", err)
			continue
		}
		if response.Code() == coap.Content {
			log.Printf("Received: \"%s\"", string(response.Payload()))
		} else {
			log.Printf("Unexpected code: \"%s\"", response.Code())
		}
	}

	log.Println("Exiting...")
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
