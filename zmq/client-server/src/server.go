package main

import (
	"log"
	zmq "github.com/pebbe/zmq4/draft"
)

func main() {
	log.Println("Zeromq Server")
	zmqContext, err := zmq.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	defer zmqContext.Term()

	soc, err := zmqContext.NewSocket(zmq.SERVER)
	if err != nil {
		log.Fatal(err)
	}
	defer soc.Close()

	if err := soc.Bind("tcp://127.0.0.1:5556"); err != nil {
		log.Fatal(err)
	}

	for {
		if msg, opts, err := soc.RecvBytesWithOpts(0, zmq.OptRoutingId(0)); err == nil {
			routingId, ok := opts[0].(zmq.OptRoutingId)
			if !ok {
				log.Fatalf("%T is not of type OptRoutingId", opts[0])
			}
			log.Printf("Received message '%s', replying with reversed message", string(msg))
			for i, j := 0, len(msg)-1; i < j; i, j = i+1, j-1 {
				msg[i], msg[j] = msg[j], msg[i]
			}
			soc.SendBytes(msg,0, routingId)
		}
	}
}
