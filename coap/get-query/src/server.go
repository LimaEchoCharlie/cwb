package main

import (
	"context"
	"github.com/go-ocf/go-coap"
	"log"
	"time"
)

func reverseHandler(w coap.ResponseWriter, req *coap.Request) {
	// get query
	query := req.Msg.Query()
	if len(query) == 0 {
		w.SetCode(coap.BadRequest)
		return
	}
	log.Printf("Received \"%s\"", query[0])

	// reverse first query
	runes := []rune(query[0])
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	reply := string(runes)

	// send response
	log.Printf("Replying \"%s\"", reply)
	w.SetContentFormat(coap.TextPlain)
	ctx, cancel := context.WithTimeout(req.Ctx, time.Second)
	defer cancel()
	if _, err := w.WriteWithContext(ctx, []byte(reply)); err != nil {
		log.Printf("Cannot send response: %v", err)
	}
}

func main() {
	mux := coap.NewServeMux()
	mux.Handle("/reverse", coap.HandlerFunc(reverseHandler))
	log.Println("Starting COAP server...")

	log.Fatal(coap.ListenAndServe("udp", ":5688", mux))
}
