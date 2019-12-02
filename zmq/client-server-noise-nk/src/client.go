package main

import (
	"bufio"
	"crypto/rand"
	"github.com/flynn/noise"
	zmq "github.com/pebbe/zmq4/draft"
	"log"
	"os"
)

func main() {
	log.Println("Zeromq Client")
	const endpoint = "tcp://127.0.0.1:5556"
	socket, err := zmq.NewSocket(zmq.CLIENT)
	if err != nil {
		log.Fatal(err)
	}
	defer socket.Close()

	err = socket.SetImmediate(true)
	if err != nil {
		log.Fatal(err)
	}

	if err := socket.Connect(endpoint); err != nil {
		log.Fatal(err)
	}
	defer socket.Disconnect(endpoint)

	// get public static key of server
	socket.Send("publicKey", 0)
	peerStatic, err := socket.RecvBytes(0)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Got server public key %q", peerStatic)

	cs := noise.NewCipherSuite(noise.DH25519, noise.CipherAESGCM, noise.HashSHA256)

	log.Println("Type messages to send, enter 'q' to exit.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
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
			PeerStatic:  peerStatic,
		})
		if err != nil {
			log.Fatal(err)
		}
		encryptedMessage, _, _, err := hs.WriteMessage(nil, scanner.Bytes())
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Sending \"%s\", encrypted %q", scanner.Text(), encryptedMessage)
		socket.SendBytes(encryptedMessage, 0)
		encryptedReply, err := socket.RecvBytes(0)
		if err != nil {
			log.Fatal(err)
		}
		reply, _, _, err := hs.ReadMessage(nil, encryptedReply)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Received %q, decrypted \"%s\"", encryptedReply, string(reply))
	}

	log.Println("Exiting...")
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
