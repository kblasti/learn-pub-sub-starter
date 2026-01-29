package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"fmt"
	"log"
	"strconv"

	"github.com/kblasti/peril/internal/gamelogic"
	"github.com/kblasti/peril/internal/pubsub"
	"github.com/kblasti/peril/internal/routing"
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

	fmt.Println("Starting Peril client...")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	gs := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		"pause." + username,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix + "." + username,
		routing.ArmyMovesPrefix + ".*",
		pubsub.Transient,
		handlerMove(gs, channel),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		"war",
		routing.WarRecognitionsPrefix + ".*",
		pubsub.Durable,
		handlerWar(gs, channel),
	)
	if err != nil {
		log.Fatal(err)
	}

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "spawn":
			err = gs.CommandSpawn(input)
			if err != nil {
				fmt.Println(err)
			}
		case "move":
			mv, err := gs.CommandMove(input)
			if err != nil {
				fmt.Println(err)
			} else {
				err = pubsub.PublishJSON(channel, routing.ExchangePerilTopic, routing.ArmyMovesPrefix + "." + username, mv)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("Move published succesfully")
				}
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			if len(input) < 2 {
				fmt.Println("Number of spams needed")
				continue
			}
			n, err := strconv.Atoi(input[1])
			if err != nil {
				fmt.Println("Error converting input to int")
				continue
			}
			for i := 0; i < n; i++ {
				msg := gamelogic.GetMaliciousLog()
				err = publishGameLog(channel, msg, username)
				if err != nil {
					fmt.Println("Error posting spam")
					break
				}
			}
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}
