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
	token []byte	// hold the token used during the handshake
}

func (c *coapClientMessenger) Exchange(message []byte) (reply []byte, err error) {
	replyMessage, err := c.clientConn.Post("/handshake", coap.TextPlain, bytes.NewReader(message))
	if replyMessage.Code() != coap.Changed {
		err = fmt.Errorf("unexpected status response: %s", replyMessage.Code())
	}
	c.token = replyMessage.Token()
	return replyMessage.Payload(), err
}

// newCOAPClientMessenger creates a new COAP client messenger
func newCOAPClientMessenger(conn *coap.ClientConn) *coapClientMessenger {
	return &coapClientMessenger{clientConn:conn}
}

func main() {
	clientConn, err := coap.Dial("udp", "localhost:5688")
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}
	log.Println("Initialising handshake...")
	messenger := newCOAPClientMessenger(clientConn)
	csPair, err := noise.ClientHandshake(messenger)
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
		ctx, cancel = context.WithTimeout(context.Background(), 5 * time.Second)
		message, err := clientConn.NewGetRequest("/reverse")
		if err != nil {
			log.Printf("Error creating request: %v", err)
			continue
		}
		encryptedText := csPair.Encrypter.Encrypt(nil, nil, scanner.Bytes())
		log.Printf("Sending: \"%s\", encrypted %q", scanner.Text(), encryptedText)
		message.SetQueryString(string(encryptedText))

		// re-use the token from the handshake exchange so the server knows which cipher states to use
		// should not be used for concurrent requests
		message.SetToken(messenger.token)

		response, err := clientConn.ExchangeWithContext(ctx, message)
		if err != nil {
			log.Printf("Error sending request: %v", err)
			continue
		}
		if response.Code() != coap.Content {
			log.Printf("Unexpected code: \"%s\"", response.Code())
			continue
		}

		decryptedReply, err := csPair.Decrypter.Decrypt(nil, nil, response.Payload())
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
