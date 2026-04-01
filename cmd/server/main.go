package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	connString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalf("can not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connection successful!")

	publishChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("can not create channel: %v", err)
	}

	/*
		_, queue, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.SimpleQueueDurable)
		if err != nil {
			log.Fatalf("can not bind topic queue: %v", err)
		}
		fmt.Printf("Queue %s declared and bound!\n", queue.Name)
	*/
	err = pubsub.SubscribeGob(conn, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.SimpleQueueDurable, handlerLogs)
	if err != nil {
		log.Fatalf("could not start consuming logs: %v", err)
	}

	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause":
			fmt.Println("Sending pause message...")
			setPause(publishChan, true)
		case "resume":
			fmt.Println("Sending resume message...")
			setPause(publishChan, false)
		case "quit":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Printf("Invalid command!")
		}
	}

	/*

	 */
	/*
	   // wait for ctrl+c
	   signalChan := make(chan os.Signal, 1)
	   signal.Notify(signalChan, os.Interrupt)
	   <-signalChan

	   fmt.Println("Server shutting down...")
	*/
}

func setPause(publishChan *amqp.Channel, paused bool) {

	data := routing.PlayingState{
		IsPaused: paused,
	}
	err := pubsub.PublishJSON(publishChan, routing.ExchangePerilDirect, routing.PauseKey, data)
	if err != nil {
		log.Fatalf("could not publish time: %v", err)
	}
}
