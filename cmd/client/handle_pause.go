package main

import (
	"github.com/kblasti/peril/internal/gamelogic"
	"github.com/kblasti/peril/internal/routing"
	"github.com/kblasti/peril/internal/pubsub"
	
	"fmt"
)

func handlerPause(gs *gamelogic.GameState) (func(routing.PlayingState) pubsub.AckType){
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}