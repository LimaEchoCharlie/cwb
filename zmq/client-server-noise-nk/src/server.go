package main

import (
	"crypto/rand"
	"github.com/flynn/noise"
	zmq "github.com/pebbe/zmq4/draft"
	"log"
)

func main() {
	log.Println("Zeromq Server")
	zmqContext, err := zmq.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	defer zmqContext.Term()

	soc, err := zmqContext.NewSocket(zmq.SERVER)
	if err != nil {
		log.Fatal(err)
	}
	defer soc.Close()

	if err := soc.Bind("tcp://127.0.0.1:5556"); err != nil {
		log.Fatal(err)
	}

	//
	cs := noise.NewCipherSuite(noise.DH25519, noise.CipherAESGCM, noise.HashSHA256)
	staticKey, err := cs.GenerateKeypair(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	for {
		if rawMessage, opts, err := soc.RecvBytesWithOpts(0, zmq.OptRoutingId(0)); err == nil {
			routingId, ok := opts[0].(zmq.OptRoutingId)
			var reply []byte
			if !ok {
				log.Fatalf("%T is not of type OptRoutingId", opts[0])
			}
			if string(rawMessage) == "publicKey" {
				reply = staticKey.Public
			} else {
				hs, err := noise.NewHandshakeState(noise.Config{
					CipherSuite:   cs,
					Random:        rand.Reader,
					Pattern:       noise.HandshakeNK,
					StaticKeypair: staticKey,
				})
				if err != nil {
					log.Fatal(err)
				}

				message, _, _, err := hs.ReadMessage(nil, rawMessage)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Received %q, decrypted \"%s\"", string(rawMessage), string(message))
				for i, j := 0, len(message)-1; i < j; i, j = i+1, j-1 {
					message[i], message[j] = message[j], message[i]
				}
				reply, _, _, err = hs.WriteMessage(nil, message)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Replying \"%s\", encrypted %q", string(message), string(reply))
			}
			soc.SendBytes(reply, 0, routingId)
		}
	}
}
