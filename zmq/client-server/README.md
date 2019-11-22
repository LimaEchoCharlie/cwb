# ZeroMQ client-server example

Build and run the docker container:

    docker build -t zmq-cs .
    docker run -it -d --name zmq-cs zmq-cs
    
Run server:

    docker exec -it zmq-cs bash
    go run src/server.go

Run client:

    docker exec -it zmq-cs bash
    go run src/client.go
