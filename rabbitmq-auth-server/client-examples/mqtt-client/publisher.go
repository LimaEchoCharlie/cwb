package main

import (
	"github.com/eclipse/paho.mqtt.golang"
	"log"
)

const (
	username = "chinchilla"
	//password = "password"
	password = "BdbpqBhtmF0Yc6ZOsWEo6bbBLsI"
	topic = "mqtt-example"
	serverURI = "tcp://127.0.0.1:1883"
)

func main()  {
	// create client
	opts := mqtt.NewClientOptions()
	opts.AddBroker(serverURI).SetUsername(username).SetPassword(password)
	client := mqtt.NewClient(opts)

	// connect to server
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// publish message
	msg := "hello from " + username
	if ok := client.Publish(topic, 1, true, msg).Wait(); !ok {
		panic("Publish failed")
	}
	log.Printf("publisher sent \"%s\"", msg)
}
