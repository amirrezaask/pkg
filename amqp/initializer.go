package amqp

import (
	"log/slog"

	"github.com/rabbitmq/amqp091-go"
)

type ExchangeInitializeConfig struct {
	//
	// Exchange name
	//
	Name string

	//
	// direct|fanout
	//
	Type string
}

type QueueInitializeConfig struct {
	Name       string
	Exchange   string
	RoutingKey string
}

type RabbitInitializeConfig struct {
	ConnectionString string
	Exchange         []ExchangeInitializeConfig
	Queue            []QueueInitializeConfig
}

func InitializeRabbit(rcfgs []RabbitInitializeConfig) {
	for _, rcfg := range rcfgs {
		conn, err := amqp091.DialConfig(rcfg.ConnectionString, amqp091.Config{
			Properties: amqp091.NewConnectionProperties(),
		})
		if err != nil {
			panic(err)
		}

		ch, err := conn.Channel()
		if err != nil {
			panic(err)
		}
		defer ch.Close()
		for _, ex := range rcfg.Exchange {
			err = ch.ExchangeDeclare(ex.Name, ex.Type, true, false, false, false, amqp091.Table{})
			if err != nil {
				slog.Error("cannot declare exchange in amqp.InitializeRabbit", "err", err)
				panic(err)
			}
		}

		for _, qu := range rcfg.Queue {
			_, err = ch.QueueDeclare(qu.Name, true, false, false, false, amqp091.Table{})
			if err != nil {
				slog.Error("cannot declare queue in amqp.InitializeRabbit", "err", err)
				panic(err)
			}

			err = ch.QueueBind(
				qu.Name,
				qu.RoutingKey,
				qu.Exchange, false, amqp091.Table{})
			if err != nil {
				slog.Error("cannot bind queue to financial exchange in amqp.InitFinancialExchangeAndQueues", "err", err)
				panic(err)
			}
		}

	}

}
