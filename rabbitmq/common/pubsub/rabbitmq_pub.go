package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"thanhldt060802/internal/rabbitmqclient"
	"thanhldt060802/model"
	"time"

	"github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

var RabbitMqPubInstance1 IRabbitMqPub[string]
var RabbitMqPubInstance2 IRabbitMqPub[*model.DataStruct]

type IRabbitMqPub[T any] interface {
	PublishWithRetry(ctx context.Context, exchange string, routingKey string, data T)
}

type RabbitMqPub[T any] struct {
	channel *amqp091.Channel

	confirmCh <-chan amqp091.Confirmation
	lock      sync.Mutex
}

func NewRabbitMqPub[T any]() (IRabbitMqPub[T], error) {
	if channel, err := rabbitmqclient.RabbitMQClientConnInstance.NewChannel(); err != nil {
		log.Errorf("Create new channel failed: %v", err.Error())
		return nil, err
	} else {
		if err := channel.Confirm(false); err != nil {
			log.Errorf("Turn off confirm option of channel failed: %v", err.Error())
			return nil, err
		}

		return &RabbitMqPub[T]{
			channel:   channel,
			confirmCh: channel.NotifyPublish(make(chan amqp091.Confirmation, 1)),
		}, nil
	}
}

func (rabbitMqPub *RabbitMqPub[T]) PublishWithRetry(ctx context.Context, exchange string, routingKey string, data T) {
	body, err := json.Marshal(data)
	if err != nil {
		log.Errorf("Marshal data failed: %v", err.Error())
		return
	}

	rabbitMqPub.lock.Lock()
	defer rabbitMqPub.lock.Unlock()

	closeChan := rabbitMqPub.channel.NotifyClose(make(chan *amqp091.Error, 1))

	for {
		err := rabbitMqPub.startPublishWithConfirm(ctx, exchange, routingKey, body)
		if err == nil {
			log.Infof("Publish %v to %v of %v successful", data, routingKey, exchange)
			return
		}

		log.Errorf("Publish %v to %v of %v failed: %v", data, routingKey, exchange, err.Error())

		select {
		case <-ctx.Done():
			log.Infof("Context canceled, stop publishing to %v of %v", routingKey, exchange)
			return
		case rabbitErr := <-closeChan:
			if rabbitErr != nil {
				log.Errorf("Channel of publisher for %v of %v closed: %v. Retry in 5s...", routingKey, exchange, rabbitErr.Error())
			} else {
				log.Warnf("Channel of publisher for %v of %v closed cleanly. Retry in 5s...", routingKey, exchange)
			}
			time.Sleep(5 * time.Second)

			for {
				newCh, chErr := rabbitmqclient.RabbitMQClientConnInstance.NewChannel()
				if chErr != nil {
					log.Errorf("Create new channel failed: %v. Retry in 5s...", chErr.Error())
					time.Sleep(5 * time.Second)
					continue
				}
				rabbitMqPub.channel = newCh

				if err := newCh.Confirm(false); err != nil {
					log.Errorf("Turn off confirm option of channel failed: %v, Retry in 5s...", err.Error())
					time.Sleep(5 * time.Second)
					continue
				}

				rabbitMqPub.confirmCh = newCh.NotifyPublish(make(chan amqp091.Confirmation, 1))

				closeChan = rabbitMqPub.channel.NotifyClose(make(chan *amqp091.Error, 1))
				break
			}
		default:
			log.Warnf("Publish %v to %v of %v failed but channel still open. Retry in 5s...", data, routingKey, exchange)
			time.Sleep(5 * time.Second)
		}
	}
}

func (rabbitMqPub *RabbitMqPub[T]) startPublishWithConfirm(ctx context.Context, exchange string, routingKey string, body []byte) error {
	err := rabbitMqPub.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return err
	}

	select {
	case confirm, ok := <-rabbitMqPub.confirmCh:
		if !ok {
			return fmt.Errorf("confirm channel closed")
		}
		if confirm.Ack {
			return nil
		}
		return fmt.Errorf("nack received for delivery tag: %d", confirm.DeliveryTag)
	case <-ctx.Done():
		return fmt.Errorf("publish confirm canceled")
	}

}
