# gRPC and Protobuf
## running
	# generate code
	protoc --go_out=plugins=grpc:src/syml syml.proto

	# run server
	env GOPATH=$GOPATH:"$(pwd)" go run server.go

	# run client
	env GOPATH=$GOPATH:"$(pwd)" go run client.go
