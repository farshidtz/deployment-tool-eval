package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stianeikeland/go-rpio"
)

type Config struct {
	Duration int
	Pin      int
	Broker   string
	Topic    string
}

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	config := loadConfig("config.json")
	log.Print("Configured duration: ", config.Duration)
	log.Print("Configured pin: ", config.Pin)
	log.Print("Configured mqtt broker: ", config.Broker)

	if config.Duration == 0 {
		log.Fatal("Duration cannot be zero!")
	}

	err := rpio.Open()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Opened GPIO")

	pin := rpio.Pin(config.Pin)
	pin.Output() // set GPIO as output
	pin.Low()    // turn off by default

	client := setupMQTTClient(config.Broker)

	client.Subscribe(config.Topic, 0,
		func(client mqtt.Client, msg mqtt.Message) {
			log.Printf("received message: %s", msg.Payload())
			pin.High()
			time.Sleep(time.Duration(config.Duration) * time.Second)
			pin.Low()
		})

	waitForInterrupt()
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
