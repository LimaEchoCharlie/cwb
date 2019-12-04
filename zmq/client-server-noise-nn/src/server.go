package main

import (
	"log"
	zmq "github.com/pebbe/zmq4/draft"
	"github.com/limaechocharlie/cwb/shared/noise"
	"math/rand"
	"encoding/json"
)

type zmqServerMessenger struct {
	*zmq.Socket
	routingId zmq.OptRoutingId
}

func (z *zmqServerMessenger) Send(message []byte) (err error)  {
	_, err = z.SendBytes(message,0, z.routingId)
	return
}

type inboundMessage struct {
	SessionID noise.EncryptionSessionID
	Payload   []byte
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

	clients := make(map[noise.EncryptionSessionID]noise.CipherStatePair)

	for {
		if b, opts, err := socket.RecvBytesWithOpts(0, zmq.OptRoutingId(0)); err == nil {
			routingId, ok := opts[0].(zmq.OptRoutingId)
			if !ok {
				log.Printf("%T is not of type OptRoutingId", opts[0])
				continue
			}
			inbound := inboundMessage{}
			if err := json.Unmarshal(b, &inbound); err != nil {
				log.Println(err)
				continue
			}
			var cipherStates noise.CipherStatePair
			if inbound.SessionID == noise.HandshakeSessionID {
				id := noise.EncryptionSessionID(rand.Uint32())
				cipherStates, err = noise.ServerHandshake(
					&zmqServerMessenger{socket, routingId},
					id,
					inbound.Payload)
				if err != nil {
					log.Println(err)
					continue
				}
				clients[id] = cipherStates
				continue
			} else if cipherStates, ok = clients[inbound.SessionID]; !ok {
				log.Printf("Can't find encryption session %d", inbound.SessionID)
				continue
			}
			payload, err := cipherStates.Decrypter.Decrypt(nil, nil, inbound.Payload)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Received %q, decrypted \"%s\"", string(b), string(payload))
			for i, j := 0, len(payload)-1; i < j; i, j = i+1, j-1 {
				payload[i], payload[j] = payload[j], payload[i]
			}
			encryptedReply := cipherStates.Encrypter.Encrypt(nil, nil, payload)
			log.Printf("Replying \"%s\", encrypted %q", string(payload), string(encryptedReply))
			socket.SendBytes(encryptedReply,0, routingId)
		}
	}
}
