package main

import (
	"bufio"
	"encoding/json"
	"github.com/limaechocharlie/cwb/shared/noise"
	zmq "github.com/pebbe/zmq4/draft"
	"log"
	"os"
)

const (
	handshakeType = 0
	reverseType = 1
)

type requestMessage struct {
	MessageType int // 0 = handshake, 1 = reverse
	ChannelID   noise.ChannelID
	Payload     []byte
}

type zmqClientMessenger struct {
	*zmq.Socket
}

func (z zmqClientMessenger) Exchange(message []byte) (reply []byte, err error) {
	request, err := json.Marshal(requestMessage{MessageType: handshakeType, Payload: message})
	if err != nil {
		return
	}
	_, err = z.SendBytes(request, 0)
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
	channelID, csPair, err := noise.ClientHandshake(zmqClientMessenger{socket})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Handshake complete")

	log.Println("Type messages to send, enter 'q' to exit.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "q" {
			break
		}
		encryptedMessage := csPair.Encrypter.Encrypt(nil, nil, scanner.Bytes())
		log.Printf("Sending \"%s\", encrypted %q", scanner.Text(), encryptedMessage)
		request, err := json.Marshal(requestMessage{MessageType: reverseType, ChannelID: channelID, Payload: encryptedMessage})
		if err != nil {
			log.Fatal(err)
		}

		_, err = socket.SendBytes(request, 0)
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
