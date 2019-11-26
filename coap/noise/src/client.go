package main

import (
	"github.com/go-ocf/go-coap"
	"log"
	"os"
	"bufio"
	"time"
	"context"
	"bytes"
	"github.com/limaechocharlie/cwb/shared/noise"
	"fmt"
)

// coapClientMessenger satisfies the ClientMessenger in the noise wrapper library
type coapClientMessenger struct {
	clientConn *coap.ClientConn
}

func (c coapClientMessenger) SendReceive(message []byte) (reply []byte, err error) {
	replyMessage, err := c.clientConn.Post("/handshake", coap.TextPlain, bytes.NewReader(message))
	if replyMessage.Code() != coap.Changed {
		err = fmt.Errorf("unexpected status response: %s", replyMessage.Code())
	}
	return replyMessage.Payload(), err
}

func main() {
	clientConn, err := coap.Dial("udp", "localhost:5688")
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}
	log.Println("Initialising handshake...")
	encrypter, decrypter, err := noise.ClientHandshake(coapClientMessenger{clientConn:clientConn})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Handshake complete")
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
		message, err := clientConn.NewGetRequest("/reverse")
		if err != nil {
			log.Printf("Error creating request: %v", err)
			continue
		}
		encryptedText := encrypter.Encrypt(nil, nil, scanner.Bytes())
		log.Printf("Sending: \"%s\", encrypted %q", scanner.Text(), encryptedText)
		message.SetQueryString(string(encryptedText))

		response, err := clientConn.ExchangeWithContext(ctx, message)
		if err != nil {
			log.Printf("Error sending request: %v", err)
			continue
		}
		if response.Code() != coap.Content {
			log.Printf("Unexpected code: \"%s\"", response.Code())
			continue
		}

		decryptedReply, err := decrypter.Decrypt(nil, nil, response.Payload())
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("Received: %q, decrypted \"%s\"", response.Payload(), decryptedReply)
	}

	log.Println("Exiting...")
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
