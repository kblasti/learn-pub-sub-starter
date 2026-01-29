package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/kblasti/peril/internal/pubsub"
	"github.com/kblasti/peril/internal/routing"
	
	"time"
)

func publishGameLog(ch *amqp.Channel, message, username string) error {
	gl := routing.GameLog{
		CurrentTime:	time.Now(),
		Message:		message,
		Username:		username,
	}

	err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug + "." + username, gl)
	if err != nil {
		return err
	}

	return nil
}

