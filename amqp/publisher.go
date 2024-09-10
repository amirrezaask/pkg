package amqp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(exchange string, key string, message []byte)
}

type _Publisher struct {
	PublishingChannel chan PublishingPayload
	durationHistogram *prometheus.HistogramVec
}

type PublishingPayload struct {
	Exchange string
	Key      string
	Message  []byte
}

func NewAMQPPublisher(appName string, name string, rabbitMQURI string) Publisher {
	var a _Publisher
	var amqpCloseNotifyC chan *amqp091.Error
	var conn *amqp091.Connection
	durationHistogram := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: appName,
			Name:      fmt.Sprintf("amqp_publisher_%s", name),
			Help:      "",
			Buckets:   prometheusDurationBuckets,
		}, []string{"exchange", "routing_key"})

	a.durationHistogram = durationHistogram

	a.PublishingChannel = make(chan PublishingPayload, 100)
	restartConnection := func() {
		fmt.Printf("[Re]Starting publisher connection\n")
		var err error
		cfg := amqp091.Config{
			Properties: amqp091.NewConnectionProperties(),
		}
		conn, err = amqp091.DialConfig(rabbitMQURI, cfg)
		if err != nil {
			slog.Error("cannot connect to rabbit", "err", err)
			for range time.NewTicker(time.Second * 1).C {
				conn, err = amqp091.DialConfig(rabbitMQURI, cfg)
				if err != nil {
					slog.Error("cannot connect to rabbit", "err", err)
					continue
				}
				break
			}

		}
		amqpCloseNotifyC = make(chan *amqp091.Error, 100)
		conn.NotifyClose(amqpCloseNotifyC)
	}

	restartConnection()

	go func() {
		for {
			select {
			case <-amqpCloseNotifyC:
				fmt.Printf("notified of close connection for amqp publishing.")
				restartConnection()

			case p := <-a.PublishingChannel:
				go func() {
					timer := prometheus.NewTimer(durationHistogram.WithLabelValues(p.Exchange, p.Key))
					timer.ObserveDuration()
					ch, err := conn.Channel()
					if err != nil {
						slog.Error("cannot get amqp channel in handling publishing channel", "err", err)
						time.Sleep(time.Second * 1)
						go a.Publish(p.Exchange, p.Key, p.Message)
						restartConnection()
					}
					defer ch.Close()
					err = ch.PublishWithContext(context.Background(), p.Exchange, p.Key, false, false, amqp091.Publishing{
						Timestamp: time.Now(),
						Body:      p.Message,
					})
					if err != nil {
						slog.Error("error in publishing into amqp", "err", err)
					}
				}()

			}

		}
	}()

	return &a
}
func (a *_Publisher) Publish(exchange string, key string, message []byte) {
	go func() {
		a.PublishingChannel <- PublishingPayload{
			Exchange: exchange,
			Key:      key,
			Message:  message,
		}
	}()
}
