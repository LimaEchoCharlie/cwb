package main

import (
	"log"
	zmq "github.com/pebbe/zmq4/draft"
	"github.com/limaechocharlie/cwb/shared/noise"
)

type zmqServerMessenger struct {
	*zmq.Socket
	routingId zmq.OptRoutingId
}

func (z *zmqServerMessenger) Send(message []byte) (err error)  {
	_, err = z.SendBytes(message,0, z.routingId)
	return
}

func main() {
	log.Println("Zeromq Server")
	zmqContext, err := zmq.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	defer zmqContext.Term()

	socket, err := zmqContext.NewSocket(zmq.SERVER)
	if err != nil {
		log.Fatal(err)
	}
	defer socket.Close()

	if err := socket.Bind("tcp://127.0.0.1:5556"); err != nil {
		log.Fatal(err)
	}

	clients := make(map[zmq.OptRoutingId]noise.CipherStatePair)

	for {
		if encryptedMessage, opts, err := socket.RecvBytesWithOpts(0, zmq.OptRoutingId(0)); err == nil {
			routingId, ok := opts[0].(zmq.OptRoutingId)
			if !ok {
				log.Printf("%T is not of type OptRoutingId", opts[0])
				continue
			}
			cipherStates, ok := clients[routingId]
			if !ok {
				cipherStates, err = noise.ServerHandshake(&zmqServerMessenger{socket, routingId}, encryptedMessage)
				if err != nil {
					log.Fatal(err)
				}
				clients[routingId] = cipherStates
				continue
			}
			message, err := cipherStates.Decrypter.Decrypt(nil, nil, encryptedMessage)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Received %q, decrypted \"%s\"", string(encryptedMessage), string(message))
			for i, j := 0, len(message)-1; i < j; i, j = i+1, j-1 {
				message[i], message[j] = message[j], message[i]
			}
			encryptedReply := cipherStates.Encrypter.Encrypt(nil, nil, message)
			log.Printf("Replying \"%s\", encrypted %q", string(message), string(encryptedReply))
			socket.SendBytes(encryptedReply,0, routingId)
		}
	}
}
