package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	const connString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game client connected to RabbitMQ!")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not get register username: %v", err)
	}

	/*
		_, queue, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.SimpleQueueTransient)
		if err != nil {
			log.Fatalf("could not subscribe to pause: %v", err)
		}
		fmt.Printf("Queue %v declared and bound!\n", queue.Name)
	*/

	gamestate := gamelogic.NewGameState(username)

	pubChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+gamestate.GetUsername(), routing.PauseKey, pubsub.SimpleQueueTransient, handlerPause(gamestate))
	if err != nil {
		log.Fatalf("couldn't subscribe to the JSON queue: %v", err)
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+gamestate.GetUsername(), routing.ArmyMovesPrefix+".*", pubsub.SimpleQueueTransient, handlerMove(gamestate, pubChan))
	if err != nil {
		log.Fatalf("could not subscribe to army-move queue: %v", err)
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+".*", pubsub.SimpleQueueDurable, handlerWar(gamestate, pubChan))
	if err != nil {
		log.Fatalf("could not subscribe to war declarations queue: %v", err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			err := gamestate.CommandSpawn(words)
			if err != nil {
				fmt.Println("invalid spawn command")
				continue
			}
		case "move":
			mv, err := gamestate.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}

			err = pubsub.PublishJSON(pubChan, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+gamestate.GetUsername(), mv)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				continue
			}
			fmt.Println("Move published successfully!")
		case "status":
			gamestate.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Invalid command!")
			continue
		}

	}
	/*
	   // wait for ctrl+c
	   signalChan := make(chan os.Signal, 1)
	   signal.Notify(signalChan, os.Interrupt)
	   <-signalChan
	   fmt.Println("RabbitMQ connection closed.")
	*/
}

func publshGameLog(ch *amqp.Channel, msg, usr string) error {
	gl := routing.GameLog{
		CurrentTime: time.Now(),
		Message:     msg,
		Username:    usr,
	}

	return pubsub.PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+usr, gl)
}
