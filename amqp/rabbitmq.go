package amqp

import (
	"context"
	"fmt"
	"time"

	"github.com/amirrezaask/go-std/logging"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitConnection struct {
	Conn *amqp091.Connection
}

// Creates a new rabbit connection, remember you should use `NotifyClose` method on Conn field
// to handle rabbit connection issues.
func NewRabbitConnection(rabbitURI string) (*RabbitConnection, error) {
	cfg := amqp091.Config{
		Properties: amqp091.NewConnectionProperties(),
	}

	conn, err := amqp091.DialConfig(rabbitURI, cfg)
	if err != nil {
		logging.Error("cannot connect to rabbit", "err", err)
		return nil, err
	}

	return &RabbitConnection{Conn: conn}, nil
}

func (rp *RabbitConnection) PublishContext(ctx context.Context, exchange string, key string, msg []byte) error {
	ch, err := rp.Conn.Channel()
	if err != nil {
		logging.Error("cannot create channel from rabbit mq connection", "err", err)
		return err
	}
	defer func() {
		err := ch.Close()
		if err != nil {
			logging.Error("cannot close channel from rabbit mq connection", "err", err)
		}
	}()

	err = ch.PublishWithContext(ctx, exchange, key, false, false, amqp091.Publishing{
		Timestamp: time.Now(),
		Body:      msg,
	})
	if err != nil {
		return err
	}
	return nil
}
func (rp *RabbitConnection) ConsumeContext(ctx context.Context, appName string, queueName string, routingKey string, exchangeName string, prefetch int,
) (<-chan amqp091.Delivery, error) {
	ch, err := rp.Conn.Channel()
	if err != nil {
		logging.Error("cannot create channel from rabbit mq connection", "err", err)
		return nil, err
	}
	// defer func() {
	// 	err := ch.Close()
	// 	if err != nil {
	// 		logging.Error("cannot close channel from rabbit mq connection", "err", err)
	// 	}
	// }()

	_, err = ch.QueueDeclare(queueName, true, false, false, false, amqp091.Table{})
	if err != nil {
		logging.Error("cannot declare rabbit queue", "err", err)
		return nil, err
	}

	err = ch.QueueBind(queueName, routingKey, exchangeName, false, amqp091.Table{})
	if err != nil {
		logging.Error("cannot bind rabbit queue", "err", err)
		return nil, err
	}
	if prefetch != 0 {
		err = ch.Qos(
			prefetch, // prefetch count
			0,        // prefetch size
			false,    // global
		)

		if err != nil {
			logging.Error("cannot set qos (prefetch) rabbit queue", "err", err)
			return nil, err
		}
	}

	delivery, err := ch.ConsumeWithContext(ctx, queueName, fmt.Sprintf("consumer-%s-%s", appName, uuid.NewString()),
		false,
		false,
		false,
		false,
		amqp091.Table{})
	if err != nil {
		logging.Error("cannot consume rabbit queue", "err", err)
		return nil, err
	}

	return delivery, nil
}
