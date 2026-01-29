package main

import (
	"github.com/kblasti/peril/internal/gamelogic"
	"github.com/kblasti/peril/internal/pubsub"
	"github.com/kblasti/peril/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	
	"fmt"
)

func handlerMove(gs *gamelogic.GameState, publishCh *amqp.Channel) (func(gamelogic.ArmyMove) pubsub.AckType){
	return func(mv gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		mvOutcome := gs.HandleMove(mv)
		switch mvOutcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix + "." + gs.GetUsername(), gamelogic.RecognitionOfWar{Attacker: mv.Player, Defender: gs.GetPlayerSnap(),})
			if err != nil {
				return pubsub.NackReque
			} else {
				return pubsub.Ack
			}
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}