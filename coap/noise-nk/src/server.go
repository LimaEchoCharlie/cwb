package main

import (
	"context"
	"github.com/go-ocf/go-coap"
	"log"
	"time"
	"crypto/rand"
	"github.com/flynn/noise"
)

// handle requests for the public key of the server
func keyHandler(key []byte) coap.HandlerFunc {
	return func(w coap.ResponseWriter, req *coap.Request) {
		log.Println("Client has requested key")
		w.SetContentFormat(coap.TextPlain)
		ctx, cancel := context.WithTimeout(req.Ctx, time.Second)
		defer cancel()
		_, err := w.WriteWithContext(ctx, key)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// handle requests to reverse the text in the payload
// the text will arrive encrypted and response with be encrypted before it is sent back
func reverseHandler(cs noise.CipherSuite, staticKey noise.DHKey) coap.HandlerFunc {
	return func(w coap.ResponseWriter, req *coap.Request) {
		if req.Msg.Code() != coap.POST {
			log.Printf("Expected a POST but got %s", req.Msg.Code())
			w.SetCode(coap.BadRequest)
			return
		}
		payload := req.Msg.Payload()
		hs, err := noise.NewHandshakeState(noise.Config{
			CipherSuite:   cs,
			Random:        rand.Reader,
			Pattern:       noise.HandshakeNK,
			StaticKeypair: staticKey,
		})
		if err != nil {
			log.Fatal(err)
		}

		message, _, _, err := hs.ReadMessage(nil, payload)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Received %q, decrypted \"%s\"", string(payload), string(message))
		for i, j := 0, len(message)-1; i < j; i, j = i+1, j-1 {
			message[i], message[j] = message[j], message[i]
		}
		reply, _, _, err := hs.WriteMessage(nil, message)
		if err != nil {
			log.Fatal(err)
		}
		// send response
		log.Printf("Replying \"%s\", encrypted %q", string(message), reply)
		w.SetContentFormat(coap.TextPlain)
		ctx, cancel := context.WithTimeout(req.Ctx, time.Second)
		defer cancel()
		if _, err := w.WriteWithContext(ctx, reply); err != nil {
			log.Printf("Cannot send response: %v", err)
		}
	}
}

func main() {
	cs := noise.NewCipherSuite(noise.DH25519, noise.CipherAESGCM, noise.HashSHA256)
	staticKey, err := cs.GenerateKeypair(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	mux := coap.NewServeMux()
	mux.Handle("/key", keyHandler(staticKey.Public))
	mux.Handle("/reverse", reverseHandler(cs, staticKey))
	log.Println("Starting COAP server...")

	log.Fatal(coap.ListenAndServe("udp", ":5688", mux))
}
