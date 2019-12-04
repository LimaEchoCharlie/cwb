package main

import (
	"log"
	zmq "github.com/pebbe/zmq4/draft"
	"bufio"
	"os"
	"github.com/limaechocharlie/cwb/shared/noise"
	"encoding/json"
)

type requestMessage struct {
	SessionID noise.EncryptionSessionID
	Payload   []byte
}

type zmqClientMessenger struct {
	*zmq.Socket
}

func (z zmqClientMessenger) Exchange(message []byte) (reply []byte, err error) {
	request, err := json.Marshal(requestMessage{SessionID: noise.HandshakeSessionID, Payload:message})
	if err != nil {
		return
	}
	_, err = z.SendBytes(request,0)
	if err != nil {
		return
	}
	return z.RecvBytes(0)
}

func main() {
	log.Println("Zeromq Client")
	const endpoint = "tcp://127.0.0.1:5556"
 	socket, err := zmq.NewSocket(zmq.CLIENT)
	if err != nil {
		log.Fatal(err)
	}
	defer socket.Close()

	// set immediate so that messages shall be queued only to completed connections (avoid lost messages)
	err = socket.SetImmediate(true)
	if err != nil {
		log.Fatal(err)
	}

	if err := socket.Connect(endpoint); err != nil {
		log.Fatal(err)
	}
	defer socket.Disconnect(endpoint)

	log.Printf("Initiate client handshake")
	sessionID, csPair, err := noise.ClientHandshake(zmqClientMessenger{socket})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Handshake completed (ID:%d)", sessionID)

	log.Println("Type messages to send, enter 'q' to exit.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "q" {
			break
		}
		encryptedMessage := csPair.Encrypter.Encrypt(nil, nil, scanner.Bytes())
		log.Printf("Sending \"%s\", encrypted %q", scanner.Text(), encryptedMessage)
		request, err := json.Marshal(requestMessage{SessionID:sessionID, Payload:encryptedMessage})
		if err != nil {
			log.Fatal(err)
		}

		_, err = socket.SendBytes(request,0)
		if err != nil {
			log.Fatal(err)
		}

		encryptedReply, err := socket.RecvBytes(0)
		if err != nil {
			log.Fatal(err)
		}

		reply, err := csPair.Decrypter.Decrypt(nil, nil, encryptedReply)
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
