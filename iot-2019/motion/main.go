package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	uuid "github.com/satori/go.uuid"
	"github.com/stianeikeland/go-rpio"
)

type Config struct {
	Pin    int
	Broker string
	Topic  string
}

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	config := loadConfig("config.json")
	log.Print("Configured pin: ", config.Pin)
	log.Print("Configured mqtt broker: ", config.Broker)
	log.Print("Configured mqtt topic: ", config.Topic)

	err := rpio.Open()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Opened GPIO")
	go waitForInterrupt() // wait for stop/kill signal asynchronously to close GPIO

	pin := rpio.Pin(config.Pin)
	pin.Input() // set GPIO as input

	client := setupMQTTClient(config.Broker)

	lastMotion := 0
	for { // read from sensor indefinitely
		if pin.Read() == 1 {
			if lastMotion == 0 {
				lastMotion = 1

				log.Print("detected motion")
				client.Publish(config.Topic, 0, false, "motion")

			}
		} else {
			lastMotion = 0
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func loadConfig(path string) Config {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func setupMQTTClient(endpoint string) mqtt.Client {
	opts := mqtt.NewClientOptions().AddBroker(endpoint)
	opts.SetClientID("light-"+uuid.Must(uuid.NewV4()).String())

	opts.SetOnConnectHandler(func(pahoClient mqtt.Client) {
		log.Printf("MQTT: Connected")
	})
	opts.SetConnectionLostHandler(func(pahoClient mqtt.Client, err error) {
		log.Printf("MQTT: Disconnected: %s", err)
	})

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	return c

}

func waitForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Print("Shutting down...")
	rpio.Close()
	os.Exit(1)
}
