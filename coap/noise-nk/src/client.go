package main

import (
	"github.com/go-ocf/go-coap"
	"log"
	"bufio"
	"os"
	"context"
	"github.com/flynn/noise"
	"crypto/rand"
	"time"
	"bytes"
)

func main() {
	clientConn, err := coap.Dial("udp", "localhost:5688")
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}
	log.Println("Get server key...")
	keyResponse, err := clientConn.Get("/key")
	if err != nil {
		log.Fatal(err)
	} else if keyResponse.Code() != coap.Content {
		log.Fatalf("Unexpected code %s", keyResponse.Code())
	}
	log.Printf("Server key %q", keyResponse.Payload())

	cs := noise.NewCipherSuite(noise.DH25519, noise.CipherAESGCM, noise.HashSHA256)
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
		// request and response is transported via the handshake payload
		// if the communication requires more than one round trips then cipher states
		// must be used for subsequent communication
		hs, err := noise.NewHandshakeState(noise.Config{
			CipherSuite: cs,
			Random:      rand.Reader,
			Pattern:     noise.HandshakeNK,
			Initiator:   true,
			PeerStatic:  keyResponse.Payload(),
		})
		if err != nil {
			log.Fatal(err)
		}
		encryptedMessage, _, _, err := hs.WriteMessage(nil, scanner.Bytes())
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Sending \"%s\", encrypted %q", scanner.Text(), encryptedMessage)
		ctx, cancel = context.WithTimeout(context.Background(), 5 * time.Second)

		response, err := clientConn.PostWithContext(ctx, "/reverse",coap.TextPlain,bytes.NewBuffer(encryptedMessage))
		if err != nil {
			log.Printf("Error sending request: %v", err)
			continue
		}
		if response.Code() != coap.Changed {
			log.Printf("Unexpected code: \"%s\"", response.Code())
			continue
		}

		decryptedReply, _,_,err := hs.ReadMessage(nil, response.Payload())
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
