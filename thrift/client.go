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
	"context"
	"fmt"
	"syml"

	"github.com/apache/thrift/lib/go/thrift"
	"encoding/json"
)

var defaultCtx = context.Background()

type exampleParameters struct {
	N int `json:"n"`
	B bool `json:"b"`
	F float64 `json:"f"`
}

func handleClient(client *syml.SimpleServiceClient) (err error) {
	var reply string
	fmt.Println("ping")
	client.Ping(defaultCtx)

	fmt.Println("get string")
	if reply, err = client.GetString(defaultCtx, "Bob"); err != nil {
		return err
	}
	fmt.Println(reply)

	fmt.Println("run custom command")
	exampleParams := exampleParameters{42, true, 3.14}
	b, _ := json.Marshal(exampleParams)
	if reply, err = client.RunCustomCommand(defaultCtx, "Bob", &syml.Command{"unpack", b}); err != nil {
		return err
	}
	fmt.Println(reply)

	fmt.Println("run custom command with unexpected command name")
	_, expectedErr := client.RunCustomCommand(defaultCtx, "Bob", &syml.Command{"wrong", b})
	switch v := expectedErr.(type){
	case *syml.SimpleError:
		fmt.Println(v.Message)
	default:
		fmt.Println(expectedErr)
	}
	return err
}

func runClient(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string) error {
	var transport thrift.TTransport
	var err error
	transport, err = thrift.NewTSocket(addr)
	if err != nil {
		fmt.Println("Error opening socket:", err)
		return err
	}
	transport, err = transportFactory.GetTransport(transport)
	if err != nil {
		return err
	}
	defer transport.Close()
	if err := transport.Open(); err != nil {
		return err
	}
	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)
	return handleClient(syml.NewSimpleServiceClient(thrift.NewTStandardClient(iprot, oprot)))
}
