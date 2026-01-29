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

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	err = channel.Qos(10, 0, false)
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
			msg, err = unmarshaller(d.Body)
			if err != nil {
				fmt.Println(err)
			}
			acktype := handler(msg)
			switch acktype {
			case Ack:
				err = d.Ack(false)
				fmt.Println("message acknowledged")
				if err != nil {
					fmt.Println(err)
				}
			case NackReque:
				err = d.Nack(false, true)
				fmt.Println("message not acknowledged, requeing")
				if err != nil {
					fmt.Println(err)
				}
			case NackDiscard:
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

func jsonUnmarshaller[T any](b []byte) (T, error) {
    var v T
    err := json.Unmarshal(b, &v)
    return v, err
}

func gobUnmarshaller[T any](b []byte) (T, error) {
	var v T
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&v)
	return v, err
}

func SubscribeJSON[T any](
    conn *amqp.Connection,
    exchange,
    queueName,
    key string,
    queueType SimpleQueueType,
    handler func(T) AckType,
) error {
	err := subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		jsonUnmarshaller[T],
	)
	return err
}

func SubscribeGob[T any](
    conn *amqp.Connection,
    exchange,
    queueName,
    key string,
    queueType SimpleQueueType,
    handler func(T) AckType,
) error {
	err := subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		gobUnmarshaller[T],
	)
	return err
}