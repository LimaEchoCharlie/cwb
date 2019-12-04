package main

import (
	"context"
	"encoding/json"
	"github.com/go-ocf/go-coap"
	"github.com/limaechocharlie/cwb/shared/noise"
	"log"
	"math/rand"
	"time"
)

type clientCiphers map[noise.EncryptionSessionID]noise.CipherStatePair

type inboundMessage struct {
	SessionID noise.EncryptionSessionID
	Payload   []byte
}

// coapServerMessenger satisfies the ServerMessenger in the noise wrapper library
type coapServerMessenger struct {
	w   coap.ResponseWriter
	req *coap.Request
}

func (c coapServerMessenger) Send(message []byte) (err error) {
	c.w.SetContentFormat(coap.TextPlain)
	ctx, cancel := context.WithTimeout(c.req.Ctx, time.Second)
	defer cancel()
	_, err = c.w.WriteWithContext(ctx, []byte(message))
	return err
}

func handshakeHandler(ciphers clientCiphers) coap.HandlerFunc {
	return func(w coap.ResponseWriter, req *coap.Request) {
		log.Println("Client has initiated handshake")
		sessionID := noise.EncryptionSessionID(rand.Uint32())
		csPair, err := noise.ServerHandshake(coapServerMessenger{w, req}, sessionID, req.Msg.Payload())
		if err != nil {
			log.Fatal(err)
		}
		ciphers[sessionID] = csPair
		log.Println("Handshake with client completed")
	}
}

func reverseHandler(ciphers clientCiphers) coap.HandlerFunc {
	return func(w coap.ResponseWriter, req *coap.Request) {
		if req.Msg.Code() != coap.POST {
			log.Printf("Expected a POST but got %s", req.Msg.Code())
			w.SetCode(coap.BadRequest)
			return
		}
		var inbound inboundMessage
		if err := json.Unmarshal(req.Msg.Payload(), &inbound); err != nil {
			log.Printf("Unable to unmarshall payload; %s", err)
			w.SetCode(coap.BadRequest)
			return
		}

		csPair, ok := ciphers[inbound.SessionID]
		if !ok {
			w.SetCode(coap.Unauthorized)
			return
		}
		msg, err := csPair.Decrypter.Decrypt(nil, nil, inbound.Payload)
		if err != nil {
			log.Println(err)
			w.SetCode(coap.Unauthorized)
			return
		}
		log.Printf("Received %q, decrypted \"%s\"", inbound.Payload, string(msg))

		// reverse first query
		for i, j := 0, len(msg)-1; i < j; i, j = i+1, j-1 {
			msg[i], msg[j] = msg[j], msg[i]
		}

		// encrypt response
		encryptedReply := csPair.Encrypter.Encrypt(nil, nil, msg)

		// send response
		log.Printf("Replying \"%s\", encrypted %q", string(msg), encryptedReply)
		w.SetContentFormat(coap.TextPlain)
		ctx, cancel := context.WithTimeout(req.Ctx, time.Second)
		defer cancel()
		if _, err := w.WriteWithContext(ctx, encryptedReply); err != nil {
			log.Printf("Cannot send response: %v", err)
		}
	}
}

func main() {
	ciphers := make(clientCiphers)
	mux := coap.NewServeMux()
	mux.Handle("/handshake", handshakeHandler(ciphers))
	mux.Handle("/reverse", reverseHandler(ciphers))
	log.Println("Starting COAP server...")

	log.Fatal(coap.ListenAndServe("udp", ":5688", mux))
}
