package main

import (
	"github.com/eclipse/paho.mqtt.golang"
	"log"
)

const (
	username = "cottontail"
	password = "password"
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

	// subscribe
	client.Subscribe(topic, 1, func(_ mqtt.Client, message mqtt.Message) {
		log.Printf("subscriber received: \"%s\" on topic %s", string(message.Payload()), message.Topic())
	}).Wait()

	// wait forever
	select {}
}
