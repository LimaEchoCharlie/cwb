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
	"os"
	"flag"
	"image"
	"sync"
	"time"
	"crypto/tls"
	"io/ioutil"
	"crypto/x509"
)

var defaultCtx = context.Background()

func fullHandler(client *syml.SimpleServiceClient) (err error) {
	const id = "holl"
	var reply string
	fmt.Println("run custom command")
	b, _ := json.Marshal(image.Rect(1,2,3,5))
	if reply, err = client.CustomCommand(defaultCtx, id, &syml.Command{"area", b}); err != nil {
		return err
	}
	fmt.Println(reply)

	fmt.Println("run custom command with unexpected command name")
	_, expectedErr := client.CustomCommand(defaultCtx, id, &syml.Command{"wrong", b})
	switch v := expectedErr.(type){
	case *syml.SimpleError:
		fmt.Println(v.Message)
	default:
		fmt.Println(expectedErr)
	}

	fmt.Println("run snooze")
	if err = client.Snooze(defaultCtx, id, 2); err != nil {
		return err
	}
	return err
}

func createSnoozeHandler(i , secs int) func(client *syml.SimpleServiceClient) error {
	return func(client *syml.SimpleServiceClient) (err error) {
		fmt.Printf("start snooze %d\n", i)
		if err = client.Snooze(defaultCtx, fmt.Sprintf("%d", i), int64(secs)); err != nil {
			return err
		}
		fmt.Printf("end snooze %d\n", i)
		return err
	}
}
func runClient(handler func(client *syml.SimpleServiceClient) error, transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string) error {
	var transport thrift.TTransport
	var err error

	// load client certificate and key
	clientCert, err := tls.LoadX509KeyPair("testdata/client-cert.pem", "testdata/client-key.pem")
	if err != nil {
		return err
	}

	// load server certificate and add to certificate pool
	serverBytes, err := ioutil.ReadFile("testdata/server-cert.pem")
	if err != nil {
		return err
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(serverBytes)
	if !ok {
		fmt.Errorf("failed to append cert from PEM")
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs: certPool,
	}
	transport, err = thrift.NewTSSLSocket(addr, cfg)
	if err != nil {
		return err
	}
	transport, err = transportFactory.GetTransport(transport)
	if err != nil {
		return err
	}
	if err := transport.Open(); err != nil {
		return err
	}
	defer transport.Close()
	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)
	return handler(syml.NewSimpleServiceClient(thrift.NewTStandardClient(iprot, oprot)))
}

func clientUsage() {
	fmt.Fprintln(os.Stderr, "Usage of client example:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "If nClients is zero, then a single client runs through all the server functions")
	fmt.Fprintln(os.Stderr, "Otherwise, clients are created every second (up to nClients) that each call the")
	fmt.Fprintln(os.Stderr, "snooze function of the server")
	fmt.Fprintln(os.Stderr, "")
}

func main() {
	flag.Usage = clientUsage
	addr := flag.String("addr", "localhost:9090", "Address to listen to")
	nClients := flag.Int("multi", 0, "Number of clients")

	flag.Parse()

	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transportFactory := thrift.NewTTransportFactory()

	if *nClients == 0 {
		if err := runClient(fullHandler, transportFactory, protocolFactory, *addr); err != nil {
			fmt.Println("error running client:", err)
		}
	} else {
		var wg sync.WaitGroup
		wg.Add(*nClients)
		for i := 0; i < *nClients; i++ {
			go func(i int) {
				defer wg.Done()
				runClient(createSnoozeHandler(i, 10), transportFactory, protocolFactory, *addr)
			}(i)
			time.Sleep(time.Second)
		}
		wg.Wait()
	}
}
