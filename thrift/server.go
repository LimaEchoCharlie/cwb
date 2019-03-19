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
	"time"
	"os"
	"flag"
	"image"
	"crypto/tls"
	"io/ioutil"
	"crypto/x509"
)

type simpleHandler struct {
}

func (p *simpleHandler) GetString(ctx context.Context, id string) (r string, err error){
	return fmt.Sprintf("Hello from %s", id), nil
}

func (p *simpleHandler) RunCustomCommand(ctx context.Context, id string, cmd *syml.Command) (r string, err error){
	fmt.Printf("name - %s, parameters = %v\n", cmd.Name, cmd.Parameters)
	if cmd.Name != "area" {
		simpleErr := syml.NewSimpleError()
		simpleErr.Message = fmt.Sprintf("Unexpected command name \"%s\"", cmd.Name)
		return "", simpleErr
	}
	var rect image.Rectangle
	if !cmd.IsSetParameters() {
		fmt.Println("nil array")
	} else if err:=json.Unmarshal(cmd.Parameters, &rect); err != nil {
		return "", err
	}
	return fmt.Sprintf("The area of the rectangle is %d", rect.Dx() * rect.Dy()), nil
}

func (p *simpleHandler) Ping(ctx context.Context) (err error) {
	fmt.Print("ping()\n")
	return nil
}

func (p *simpleHandler) Snooze(ctx context.Context, id string, secs int64) (err error){
	fmt.Printf("snooze (%s) in:  %s\n", id, time.Now().Format("15:04:05"))
	time.Sleep(time.Duration(secs) * time.Second)
	fmt.Printf("snooze (%s) out: %s\n", id, time.Now().Format("15:04:05"))
	return nil
}

func runServer(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string) error {
	// load server certificate
	serverCert, err := tls.LoadX509KeyPair("testdata/server-cert.pem", "testdata/server-key.pem")
	if err != nil {
		return err
	}

	// load client certificate and add to certificate pool
	clientBytes, err := ioutil.ReadFile("testdata/client-cert.pem")
	if err != nil {
		return err
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(clientBytes)
	if !ok {
		fmt.Errorf("failed to append cert from PEM")
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth: tls.RequireAndVerifyClientCert, // set the server's policy for TLS Client Authentication
		ClientCAs: certPool,
	}
	transport, err := thrift.NewTSSLServerSocket(addr, cfg)
	if err != nil {
		return err
	}
	fmt.Printf("%T\n", transport)
	processor := syml.NewSimpleServiceProcessor(&simpleHandler{})
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)

	fmt.Println("Starting the simple server... on ", addr)
	return server.Serve()
}

func serverUsage() {
	fmt.Fprintln(os.Stderr, "Usage of server example:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "")
}

func main() {
	flag.Usage = serverUsage
	addr := flag.String("addr", "localhost:9090", "Address to listen to")

	flag.Parse()

	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transportFactory := thrift.NewTTransportFactory()

	if err := runServer(transportFactory, protocolFactory, *addr); err != nil {
		fmt.Println("error running server:", err)
	}
}
