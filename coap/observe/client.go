package main

import (
	"github.com/go-ocf/go-coap"
	"log"
	"time"
	"fmt"
	"context"
	"bufio"
	"os"
	"math"
)

var twentyFourBitLimit = uint64(math.Pow(2, 23))

// checks whether the new observe value from the server is truly newer than the old one
// see reordering in spec
func newerMessage(oldValue, newValue uint64) bool {
	return ((oldValue < newValue) && (newValue-oldValue) < twentyFourBitLimit) ||
	((oldValue > newValue) && (oldValue-newValue > twentyFourBitLimit))
	//|| (T2 > T1 + 128 seconds)
}

// observes the resource at the given path
// exits if the context has been cancelled or the held representation has become too old
// see freshness in spec
func observe(ctx context.Context, co *coap.ClientConn, path string, defaultMaxAge time.Duration)  {
	var prevSequence uint64 = 0
	t := time.NewTimer(defaultMaxAge)
	obs, err := co.Observe(path, func(req *coap.Request) {
		// check whether the current message is newer than the one received previously
		if !newerMessage(prevSequence, req.Sequence) {
			log.Println("Ignoring stale message")
			fmt.Println(prevSequence, req.Sequence)
			return
		}
		prevSequence = req.Sequence

		// get max age from message, if it has been set, and use the value to reset the timer
		if maxAge, ok := req.Msg.Option(coap.MaxAge).(uint32); ok {
			t.Reset(time.Duration(maxAge)*time.Second)
		}
		log.Printf("[%x] obs \"%s\"", req.Msg.Token(), req.Msg.Payload())
	})
	if err != nil {
		log.Fatalf("Unexpected error '%v'", err)
	}
	defer obs.Cancel()

	select {
	case <-ctx.Done():
		log.Println("Caller has cancelled Observe")
	case <-t.C:
		log.Println("Data has become stale")
	}
}

func main() {
	path := "/device/config"
	client := &coap.Client{}

	co, err := client.Dial("localhost:5688")
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}
	// do a normal GET
	if msg, err := co.Get(path); err != nil {
		log.Fatalf("Unexpected error '%v'", err)
	} else {
		log.Printf("[%x] got \"%s\"", msg.Token(), msg.Payload())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(500 * time.Millisecond)
	}()

	go func() {
		for  {
			select {
			case <-ctx.Done():
				return
			default:
				observe(ctx, co, path, time.Hour)
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// wait for a key press and then exit
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
}
