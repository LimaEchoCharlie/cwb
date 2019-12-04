package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-ocf/go-coap"
	"github.com/limaechocharlie/cwb/shared/noise"
	"log"
	"os"
	"time"
)

// coapClientMessenger satisfies the ClientMessenger in the noise wrapper library
type coapClientMessenger struct {
	clientConn *coap.ClientConn
	token      []byte // hold the token used during the handshake
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
	return &coapClientMessenger{clientConn: conn}
}

type reverseRequest struct {
	SessionID noise.EncryptionSessionID
	Payload   []byte
}

func main() {
	clientConn, err := coap.Dial("udp", "localhost:5688")
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}
	log.Println("Initialising handshake...")
	messenger := newCOAPClientMessenger(clientConn)
	sessionID, csPair, err := noise.ClientHandshake(messenger)
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
		encryptedText := csPair.Encrypter.Encrypt(nil, nil, scanner.Bytes())
		log.Printf("Sending: \"%s\", encrypted %q", scanner.Text(), encryptedText)

		request, err := json.Marshal(reverseRequest{SessionID: sessionID, Payload: encryptedText})
		if err != nil {
			log.Printf("Error marshalling request: %v", err)
			continue
		}

		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		response, err := clientConn.PostWithContext(ctx, "/reverse", coap.TextPlain, bytes.NewBuffer(request))
		if err != nil {
			log.Printf("Error sending request: %v", err)
			continue
		}

		if response.Code() != coap.Changed {
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
