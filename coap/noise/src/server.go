package main

import (
	"context"
	"github.com/go-ocf/go-coap"
	"log"
	"time"
	noiseHandshake "github.com/limaechocharlie/cwb/shared/noise"
	"github.com/flynn/noise"
)

var (
	encrypter *noise.CipherState = nil
	decrypter *noise.CipherState = nil
)

// coapServerMessenger satisfies the ServerMessenger in the noise wrapper library
type coapServerMessenger struct {
	w coap.ResponseWriter
	req *coap.Request
}

func (c coapServerMessenger) Receive() (message []byte, err error) {
	return c.req.Msg.Payload(), nil
}

func (c coapServerMessenger) Send(message []byte)( err error) {
	c.w.SetContentFormat(coap.TextPlain)
	ctx, cancel := context.WithTimeout(c.req.Ctx, time.Second)
	defer cancel()
	_, err = c.w.WriteWithContext(ctx, []byte(message))
	return err
}

func handshakeHandler(w coap.ResponseWriter, req *coap.Request) {
	log.Println("Client has initiated handshake")
	var err error
	decrypter, encrypter, err = noiseHandshake.ServerHandshake(coapServerMessenger{w,req})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Handshake with client completed")
}

func reverseHandler(w coap.ResponseWriter, req *coap.Request) {
	if decrypter == nil || encrypter == nil {
		w.SetCode(coap.Unauthorized)
		return
	}
	// get query
	query := req.Msg.Query()
	if len(query) == 0 {
		w.SetCode(coap.BadRequest)
		return
	}
	msg, err := decrypter.Decrypt(nil, nil, []byte(query[0]))
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
	encryptedReply := encrypter.Encrypt(nil, nil, msg)

	// send response
	log.Printf("Replying \"%s\", encrypted %q", string(msg), encryptedReply)
	w.SetContentFormat(coap.TextPlain)
	ctx, cancel := context.WithTimeout(req.Ctx, time.Second)
	defer cancel()
	if _, err := w.WriteWithContext(ctx, encryptedReply); err != nil {
		log.Printf("Cannot send response: %v", err)
	}
}

func main() {
	mux := coap.NewServeMux()
	mux.Handle("/handshake", coap.HandlerFunc(handshakeHandler))
	mux.Handle("/reverse", coap.HandlerFunc(reverseHandler))
	log.Println("Starting COAP server...")

	log.Fatal(coap.ListenAndServe("udp", ":5688", mux))
}
