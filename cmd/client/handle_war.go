package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/kblasti/peril/internal/gamelogic"
	"github.com/kblasti/peril/internal/pubsub"
	
	"fmt"
)

func handlerWar(gs *gamelogic.GameState, publishCh *amqp.Channel) (func(gamelogic.RecognitionOfWar) pubsub.AckType) {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackReque
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			msg := fmt.Sprintf("%v won a war against %v", winner, loser)
			err := publishGameLog(publishCh, msg, gs.GetUsername())
			if err != nil {
				return pubsub.NackReque
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			msg := fmt.Sprintf("%v won a war against %v", winner, loser)
			err := publishGameLog(publishCh, msg, gs.GetUsername())
			if err != nil {
				return pubsub.NackReque
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			msg := fmt.Sprintf("A war between %v and %v resulted in a draw", winner, loser)
			err := publishGameLog(publishCh, msg, gs.GetUsername())
			if err != nil {
				return pubsub.NackReque
			}
			return pubsub.Ack
		default:
			fmt.Println("Invalid war outcome response")
			return pubsub.NackDiscard
		}
	}
}