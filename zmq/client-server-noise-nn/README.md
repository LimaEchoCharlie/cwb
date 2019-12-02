# ZeroMQ client-server with Noise Protocol example

Build and run the docker container:

    docker build -t zmq-cs-noise .
    docker run -it -d --name zmq-cs-noise zmq-cs-noise
    
Run server:

    docker exec -it zmq-cs-noise bash
    go run src/server.go

Run client:

    docker exec -it zmq-cs-noise bash
    go run src/client.go
