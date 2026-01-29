package main

import (
    amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"fmt"

	"github.com/kblasti/peril/internal/routing"
	"github.com/kblasti/peril/internal/pubsub"
	"github.com/kblasti/peril/internal/gamelogic"
)

func main() {
	connectString := "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(connectString)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		log.Fatal(err)
	}

	err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
	if err != nil {
		log.Fatal(err)
	}

	_, queue, err := pubsub.DeclareAndBind(connection, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug + ".*", pubsub.Durable)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	fmt.Println("Starting Peril server...")
	gamelogic.PrintServerHelp()

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "pause":
			fmt.Println("Sending pause message...")
			err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			if err != nil {
				log.Fatal(err)
			}
		case "resume":
			fmt.Println("Sending resume message...")
			err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			if err != nil {
				log.Fatal(err)
			}
		case "quit":
			fmt.Println("Server shutting down...")
			return
		case "help":
			gamelogic.PrintServerHelp()
		default:
			fmt.Println("Unknown command...")
		}
	}
}
