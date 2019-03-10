## installation 

	brew install thrift
	go get github.com/apache/thrift/lib/go/thrift/...
	
	# generate the code
	thrift -r --gen go:thrift_import=github.com/apache/thrift/lib/go/thrift syml.thrift
	
	# create a link in your go path to your generated code
	ln -s "$(pwd)/gen-go/syml" "$(pwd)/src/syml"
	
	# run server
	env GOPATH=$GOPATH:"$(pwd)" go run server.go
	
	# run client
	env GOPATH=$GOPATH:"$(pwd)" go run client.go

## Golang translations
Examples produced using Thrift Compiler 0.11.0
### basic types and containers
|Thrift|Go|
|:---:|:---:|
|bool|bool|
|binary|[]byte|
|i8|int8|
|i16|int16|
|i32|int32|
|i64|int64|
|double|float64|
|string|string|
|list\<t1\> |[]t1|
|set\<t1\>|[]t1|
|map<t1,t2>|map[t1]t2|

### struct
	// Thrift
	struct name {
		1: bool elementOne,
		2: binary elementTwo
	}

	// Go
	type TypeName struct {
		ElementOne bool
		ElementTwo []byte
	}

### union
	// Thrift
	union unionName {
		1: bool elementOne,
		2: i8 elementTwo
	}

	// Go
	type UnionName struct {
		ElementOne *bool
		ElementTwo *int8
	}

### exception
	// Thrift
	exception errorName {
		1: string elementOne,
		2: int64 elementTwo
	}

	// Go
	type ErrorName struct {
		ElementOne string
		ElementTwo int64
	}
	func (p *ErrorName) Error() string {
	  return p.String()
	}

### enum
	// Thrift
	enum enumName {
		stop,
		walk,
		run = 10,
		dance
	}

	// Go
	type EnumName int64
	const (
		EnumName_stop EnumName = 0
		EnumName_walk EnumName = 1
		EnumName_run EnumName = 10
		EnumName_dance EnumName = 11
	)

	// Go helper functions
	func (p EnumName) String() string {
		switch p {
			case EnumName_stop: return "stop"
			case EnumName_walk: return "walk"
			case EnumName_run: return "run"
			case EnumName_dance: return "dance"
		}
		return "<UNSET>"
	}

	func EnumNameFromString(s string) (EnumName, error) {
		switch s {
			case "stop": return EnumName_stop, nil
			case "walk": return EnumName_walk, nil
			case "run": return EnumName_run, nil
			case "dance": return EnumName_dance, nil
		}
		return EnumName(0), fmt.Errorf("not a valid EnumName string")
	}

### constants
	// Thrift
	const i8 meaningOfLife = 42

	// Go
	const MeaningOfLife = 42

### typedef
	// Thrift
	typedef i64 id

	// Go
	type ID int64

#### comments
	# thrift comment
	/*
	 * Thrift also supports multi-line C comments
	 */
	//  ... and simple C comments
