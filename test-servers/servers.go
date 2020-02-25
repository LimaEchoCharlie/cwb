package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-ocf/go-coap"
	"github.com/gorilla/mux"
	"github.com/ugorji/go/codec"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"time"
)

// cartesian point
type cartesian struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (c cartesian) String() string {
	return fmt.Sprintf("( x:%v, y:%v )", c.X, c.Y)
}

// polar point
type polar struct {
	R     float64 `json:"r"`
	Theta float64 `json:"theta"`
}

func (p polar) String() string {
	return fmt.Sprintf("( r:%v, theta:%v )", p.R, p.Theta)
}

func cartesianToPolar(c cartesian) polar {
	return polar{R: math.Sqrt(c.X*c.X + c.Y*c.Y), Theta: math.Atan2(c.Y, c.X)}
}

func startHTTPServer(port string) error {
	router := mux.NewRouter()
	router.HandleFunc("/cartesian-to-polar",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error reading body, %s", err.Error()), http.StatusBadRequest)
				return
			}
			var c cartesian
			if err := json.Unmarshal(body, &c); err != nil {
				http.Error(w, fmt.Sprintf("Error reading body, %s", err.Error()), http.StatusBadRequest)
				return
			}
			log.Println("HTTP unmarshalled request:", c)
			p := cartesianToPolar(c)
			b, err := json.Marshal(p)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error marshalling response, %s", err.Error()), http.StatusBadRequest)
				return
			}
			log.Println("HTTP response", string(b))
			w.Write(b)
		})
	return http.ListenAndServe(fmt.Sprintf(":%s", port), router)
}

func startCOAPServer(port string) error {
	mux := coap.NewServeMux()
	mux.Handle("/cartesian-to-polar", coap.HandlerFunc(func(w coap.ResponseWriter, req *coap.Request) {
		handle := new(codec.CborHandle)
		if mt, ok := req.Msg.Option(coap.ContentFormat).(coap.MediaType); ok {
			log.Println("COAP request media type:", mt)
		} else {
			log.Println("COAP request NO media type")
		}

		payload := req.Msg.Payload()
		var c cartesian
		d := codec.NewDecoderBytes(payload, handle)
		if err := d.Decode(&c); err != nil {
			w.SetCode(coap.BadRequest)
			return
		}
		log.Println("COAP unmarshalled request:", c)
		p := cartesianToPolar(c)

		w.SetContentFormat(coap.AppCBOR)
		ctx, cancel := context.WithTimeout(req.Ctx, time.Second)
		defer cancel()

		var buf bytes.Buffer
		e := codec.NewEncoder(&buf, handle)
		if err := e.Encode(p); err != nil {
			w.SetCode(coap.BadRequest)
			return
		}
		log.Println("COAP response:", buf.Bytes())
		if _, err := w.WriteWithContext(ctx, buf.Bytes()); err != nil {
			log.Printf("Cannot send response: %v", err)
		}
	}))
	return coap.ListenAndServe("udp", ":"+port, mux)
}

func main() {
	httpPort := flag.String("http-port", "8001", "Port on which HTTP server listens")
	coapPort := flag.String("coap-port", "5688", "Port on which COAP server listens")

	log.Println("Starting HTTP server")
	go func() {
		log.Fatal(startHTTPServer(*httpPort))
	}()
	log.Println("HTTP server started")
	log.Println("Starting COAP server...")
	go func() {
		log.Fatal(startCOAPServer(*coapPort))
	}()
	log.Println("COAP server started")

	select {}
}
