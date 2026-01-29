package pubsub

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"context"
	"fmt"
	"encoding/gob"
	"bytes"
)

type SimpleQueueType string

const (
    Durable   SimpleQueueType = "durable"
    Transient SimpleQueueType = "transient"
)

type AckType string

const (
	Ack AckType = "ack"
	NackReque AckType = "nackreque"
	NackDiscard AckType = "nackdiscard"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonVal, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), 
	exchange, 
	key, 
	false, 
	false, 
	amqp.Publishing{
		ContentType:	"application/json",
		Body:			jsonVal,
	})
	if err != nil {
		return err
	}

	return nil
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}
	durable := queueType == Durable
	autoDelete := queueType == Transient
	exclusive := queueType == Transient

	queue, err := channel.QueueDeclare(queueName, durable, autoDelete, exclusive, false, amqp.Table{"x-dead-letter-exchange": "peril_dlx"})
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	return channel, queue, nil
}

func SubscribeJSON[T any](
    conn *amqp.Connection,
    exchange,
    queueName,
    key string,
    queueType SimpleQueueType, // an enum to represent "durable" or "transient"
    handler func(T) AckType,
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	deliveryChan, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range deliveryChan {
			var msg T
			err = json.Unmarshal(d.Body, &msg)
			if err != nil {
				fmt.Println(err)
			}
			acktype := handler(msg)
			switch acktype {
			case "ack":
				err = d.Ack(false)
				fmt.Println("message acknowledged")
				if err != nil {
					fmt.Println(err)
				}
			case "nackreque":
				err = d.Nack(false, true)
				fmt.Println("message not acknowledged, requeing")
				if err != nil {
					fmt.Println(err)
				}
			case "nackdiscard":
				err = d.Nack(false, false)
				fmt.Println("message not acknowledged, discarding")
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}()
	return nil
}

func PublishGob[T any] (ch *amqp.Channel, exchange, key string, val T) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), 
	exchange, 
	key, 
	false, 
	false, 
	amqp.Publishing{
		ContentType:	"application/gob",
		Body:			buf.Bytes(),
	})
	if err != nil {
		return err
	}

	return nil
}