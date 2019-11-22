package main

import (
	"log"
	zmq "github.com/pebbe/zmq4/draft"
	"bufio"
	"os"
)

func main() {
	log.Println("Zeromq Client")
	const endpoint = "tcp://127.0.0.1:5556"
 	socket, err := zmq.NewSocket(zmq.CLIENT)
	if err != nil {
		log.Fatal(err)
	}
	defer socket.Close()

	// set immediate so that messages shall be queued only to completed connections (avoid lost messages)
	err = socket.SetImmediate(true)
	if err != nil {
		log.Fatal(err)
	}

	if err := socket.Connect(endpoint); err != nil {
		log.Fatal(err)
	}
	defer socket.Disconnect(endpoint)

	log.Println("Type messages to send, enter 'q' to exit.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "q" {
			break
		}
		socket.SendBytes(scanner.Bytes(),0)
		reply, err := socket.Recv(0)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Received repky '%s'", reply)
	}

	log.Println("Exiting...")
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
