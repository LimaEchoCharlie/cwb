package main

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"syml"
	"context"
	"encoding/json"
)

type simpleHandler struct {
}

func (p *simpleHandler) GetString(ctx context.Context, id string) (r string, err error){
	return fmt.Sprintf("Hello from %s", id), nil
}

func (p *simpleHandler) RunCustomCommand(ctx context.Context, id string, cmd *syml.Command) (r string, err error){
	fmt.Printf("name - %s, parameters = %v\n", cmd.Name, cmd.Parameters)
	if cmd.Name != "unpack" {
		simpleErr := syml.NewSimpleError()
		simpleErr.Message = fmt.Sprintf("Unexpected command name \"%s\"", cmd.Name)
		return "", simpleErr
	}
	var parameters exampleParameters
	if !cmd.IsSetParameters() {
		fmt.Println("nil array")
	} else if err:=json.Unmarshal(cmd.Parameters, &parameters); err != nil {
		return "", err
	}
	return fmt.Sprintf("The meaning of life? %v",parameters.N), nil
}

func (p *simpleHandler) Ping(ctx context.Context) (err error) {
	fmt.Print("ping()\n")
	return nil
}

func runServer(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string) error {
	transport, err := thrift.NewTServerSocket(addr)
	if err != nil {
		return err
	}
	fmt.Printf("%T\n", transport)
	processor := syml.NewSimpleServiceProcessor(&simpleHandler{})
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)

	fmt.Println("Starting the simple server... on ", addr)
	return server.Serve()
}
