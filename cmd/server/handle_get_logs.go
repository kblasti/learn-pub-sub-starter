package main

import (
	"github.com/kblasti/peril/internal/pubsub"
	"github.com/kblasti/peril/internal/routing"
	"github.com/kblasti/peril/internal/gamelogic"
	
	"fmt"
)

func handleGameLog(gl routing.GameLog) pubsub.AckType {
	defer fmt.Print("> ")
	err := gamelogic.WriteLog(gl)
	if err != nil {
		fmt.Println(err)
		return pubsub.NackDiscard
	}
	return pubsub.Ack
}