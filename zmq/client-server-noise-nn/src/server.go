package main

import (
	"log"
	zmq "github.com/pebbe/zmq4/draft"
	"github.com/limaechocharlie/cwb/shared/noise"
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
	MessageType int // o = handshake, 1 = reverse
	ChannelID   noise.ChannelID
	Payload     []byte
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

	clients := make(map[uint64]noise.CipherStatePair)

forLoop:
	for {
		if b, opts, err := socket.RecvBytesWithOpts(0, zmq.OptRoutingId(0)); err == nil {
			routingId, ok := opts[0].(zmq.OptRoutingId)
			if !ok {
				log.Printf("%T is not of type OptRoutingId", opts[0])
				continue forLoop
			}
			inbound := inboundMessage{}
			if err := json.Unmarshal(b, &inbound); err != nil {
				log.Println(err)
				continue forLoop
			}
			switch inbound.MessageType {
			case 0:
				// handshake
				log.Println("Client has initiated handshake")
				channelID, csPair, err := noise.ServerHandshake(
					&zmqServerMessenger{socket, routingId},
					inbound.Payload)
				if err != nil {
					log.Println(err)
					continue forLoop
				}

				id, ok := channelID.UInt64()
				if !ok {
					log.Println("Unable to encode channel ID into an integer")
					continue forLoop
				}
				clients[id] = csPair
				log.Printf("Handshake with client completed [id: %d]", id)
			case 1:
				// reverse
				id, ok := inbound.ChannelID.UInt64()
				if !ok {
					log.Println("Unable to encode channel ID into an integer")
					continue forLoop
				}
				csPair, ok := clients[id]
				if!ok {
					log.Printf("Can't find channel id %d", id)
					continue forLoop
				}
				payload, err := csPair.Decrypter.Decrypt(nil, nil, inbound.Payload)
				if err != nil {
					log.Printf("Failed to decrypt payload; %s", err)
					continue forLoop
				}
				log.Printf("Received %q, decrypted \"%s\"", string(b), string(payload))
				for i, j := 0, len(payload)-1; i < j; i, j = i+1, j-1 {
					payload[i], payload[j] = payload[j], payload[i]
				}
				encryptedReply := csPair.Encrypter.Encrypt(nil, nil, payload)
				log.Printf("Replying \"%s\", encrypted %q", string(payload), string(encryptedReply))
				socket.SendBytes(encryptedReply,0, routingId)
			default:
				log.Printf("Unknown message type %d", inbound.MessageType)
			}
		}
	}
}
