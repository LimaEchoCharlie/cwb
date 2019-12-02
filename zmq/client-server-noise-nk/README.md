# ZeroMQ client-server with Noise Protocol NK example

Example uses the NK Noise protocol for the message encryption.

* **N**o static key for client.
* The server's static key is **K**nown to the client.

This handshake consists of a single request and response. 
Since the client has pre-knowledge of the server's static key, we can use zero round trip encryption
and encrypt the client request in the first handshake payload.

## Build and run

Build and run the docker container:

    docker build -t zmq-cs-nk .
    docker run -it -d --name zmq-cs-nk zmq-cs-nk
    
Run server:

    docker exec -it zmq-cs-nk bash
    go run src/server.go

Run client:

    docker exec -it zmq-cs-nk bash
    go run src/client.go
