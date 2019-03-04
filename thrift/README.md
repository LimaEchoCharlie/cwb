## installation 

	brew install thrift
	go get github.com/apache/thrift/lib/go/thrift/...
	
	# generate the code
	thrift -r --gen go:thrift_import=github.com/apache/thrift/lib/go/thrift syml.thrift
	
	# create a link in your go path to your generated code
	ln -s "$(pwd)/gen-go/syml" "$(pwd)/src/syml"
	
	# run server
	env GOPATH=$GOPATH:"$(pwd)" go run *.go -server=true
	
	# run client
	env GOPATH=$GOPATH:"$(pwd)" go run *.go
