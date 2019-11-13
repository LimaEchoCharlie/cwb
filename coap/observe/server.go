package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-ocf/go-coap"
	"math/rand"
	"context"
)

var observers = make(map[string]context.CancelFunc)

// send a single response with the server uptime in the payload
func sendResponse(w coap.ResponseWriter, req *coap.Request, startTime time.Time) error {
	resp := w.NewResponse(coap.Content)
	resp.SetOption(coap.ContentFormat, coap.TextPlain)
	resp.SetOption(coap.MaxAge, 10)
	resp.SetPayload([]byte(fmt.Sprintf("uptime %v", time.Since(startTime))))
	return w.WriteMsg(resp)
}

// repeatably transmits on the same channel with random time gaps between the responses
func randomTransmitter(ctx context.Context, w coap.ResponseWriter, req *coap.Request, startTime time.Time) {
	s1 := rand.NewSource(2)
	r1 := rand.New(s1)
	t := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			// context has been cancelled, stop process
			log.Printf("[%x] cancel process", string(req.Msg.Token()))
			t.Stop()
			return

		case <- t.C:
			// timer has fired, send a response
			err := sendResponse(w, req, startTime)
			if err != nil {
				log.Printf("Error on transmitter, stopping: %v", err)
				return
			}
			log.Printf("[%x] notification sent", string(req.Msg.Token()))
			t.Reset(time.Second * time.Duration(r1.Intn(15)))
		}
	}
}

// checks whether the message has the observe option set
// register is set to true if the observer wants to register its interest in a resource
// register is set to false if the observer wants to unregister its interest in a resource
func observeAction(msg coap.Message)(register, ok bool){
	if msg.Code() != coap.GET {
		return
	}
	v, ok := msg.Option(coap.Observe).(uint32)
	if !ok {
		return
	}
	switch v {
	case 0:
		return true, true
	case 1:
		return false, true
	default:
		return false, false
	}
}

func main() {
	startTime := time.Now()
	mux := coap.NewServeMux()
	mux.HandleFunc("/device/config",
		coap.HandlerFunc(func(w coap.ResponseWriter, req *coap.Request) {
			// only support GET requests
			if req.Msg.Code() != coap.GET {
				w.SetCode(coap.BadRequest)
				return
			}
			strToken := string(req.Msg.Token())
			log.Printf("[%x] received GET request", strToken)

			// switch behaviour on type of GET request
			// GET, observe = register; register observer, start a random transmitter
			// GET, observe = deregister; deregister observer, stop random transmitter, send one-off response
			// GET, no observe option; send one-off response
			registerObserver, observeRequest := observeAction(req.Msg)
			switch {
			case observeRequest && registerObserver:
				// start random transmitter and add token to the observer list
				log.Printf("[%x] register observer", strToken)
				ctx, cancel := context.WithCancel(context.Background())
				observers[strToken] = cancel
				go randomTransmitter(ctx, w, req, startTime)
				return
			case observeRequest && !registerObserver:
				// cancel random the transmitter associated with the token and remove from observer list
				if cancel, ok := observers[strToken]; ok {
					log.Printf("[%x] unregister observer", strToken)
					cancel()
					delete(observers, strToken)
				}
			}
			log.Printf("[%x] sending one-off response", strToken)
			err := sendResponse(w, req, startTime)
			if err != nil {
				log.Printf("[%x] error on transmitter: %v", strToken, err)
			}
		}))

	log.Fatal(coap.ListenAndServe("udp", ":5688", mux))
}
