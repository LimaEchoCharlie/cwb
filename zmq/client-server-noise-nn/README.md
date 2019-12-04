# ZeroMQ client-server with Noise Protocol NN example

Example uses the NN Noise protocol for the message encryption.

* **N**o static key for client.
* **N**o static key for server.

## Build and run

Build and run the docker container:

    docker build -t zmq-cs-noise-nn .
    docker run -it -d --name zmq-cs-noise-nn zmq-cs-noise-nn
    
Run server:

    docker exec -it zmq-cs-noise-nn bash
    go run src/server.go

Run client:

    docker exec -it zmq-cs-noise-nn bash
    go run src/client.go
