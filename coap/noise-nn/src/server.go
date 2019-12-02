package main

import (
	"context"
	"encoding/binary"
	"github.com/go-ocf/go-coap"
	"github.com/limaechocharlie/cwb/shared/noise"
	"log"
	"time"
)

type clientCiphers map[uint64]noise.CipherStatePair

// coapServerMessenger satisfies the ServerMessenger in the noise wrapper library
type coapServerMessenger struct {
	w   coap.ResponseWriter
	req *coap.Request
}

// getToken retrieves the message token as an integer
func getToken(req *coap.Request) uint64 {
	return binary.BigEndian.Uint64(req.Msg.Token())
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
		var err error
		token := getToken(req)
		ciphers[token], err = noise.ServerHandshake(coapServerMessenger{w, req}, req.Msg.Payload())
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Handshake with client completed")
	}
}

func reverseHandler(ciphers clientCiphers) coap.HandlerFunc {
	return func(w coap.ResponseWriter, req *coap.Request) {
		token := getToken(req)
		csPair, ok := ciphers[token]
		if !ok {
			w.SetCode(coap.Unauthorized)
			return
		}
		// get query
		query := req.Msg.Query()
		if len(query) == 0 {
			w.SetCode(coap.BadRequest)
			return
		}
		msg, err := csPair.Decrypter.Decrypt(nil, nil, []byte(query[0]))
		if err != nil {
			log.Println(err)
			w.SetCode(coap.Unauthorized)
			return
		}
		log.Printf("Received %q, decrypted \"%s\"", query[0], string(msg))

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
