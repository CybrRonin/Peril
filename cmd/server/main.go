package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	connString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connection successful!")

	publishChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to create channel: %v", err)
	}

	data := routing.PlayingState{
		IsPaused: true,
	}
	err = pubsub.PublishJSON(publishChan, routing.ExchangePerilDirect, routing.PauseKey, data)
	if err != nil {
		log.Fatalf("Error publishing JSON: %v", err)
	}

	fmt.Println("Pause message sent!")
	/*
	   // wait for ctrl+c
	   signalChan := make(chan os.Signal, 1)
	   signal.Notify(signalChan, os.Interrupt)
	   <-signalChan

	   fmt.Println("Server shutting down...")
	*/
}
