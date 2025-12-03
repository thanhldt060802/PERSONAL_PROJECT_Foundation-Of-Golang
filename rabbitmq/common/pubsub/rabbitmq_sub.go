package pubsub

import (
	"context"
	"encoding/json"
	"reflect"
	"thanhldt060802/internal/rabbitmqclient"
	"thanhldt060802/model"
	"time"

	"github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

var RabbitMqSubInstance1 IRabbitMqSub[string]
var RabbitMqSubInstance2 IRabbitMqSub[*model.DataStruct]

type IRabbitMqSub[T any] interface {
	ConsumeWithRetry(ctx context.Context, exchange string, queue string, routingKey string, prefetchCount int, handler func(data T) error, dlxTable amqp091.Table)
}

type RabbitMqSub[T any] struct {
	channel *amqp091.Channel
}

func NewRabbitMqSub[T any]() (IRabbitMqSub[T], error) {
	if channel, err := rabbitmqclient.RabbitMQClientConnInstance.NewChannel(); err != nil {
		log.Errorf("Create new channel failed: %v", err.Error())
		return nil, err
	} else {
		return &RabbitMqSub[T]{
			channel: channel,
		}, nil
	}
}

func (rabbitMqSub *RabbitMqSub[T]) ConsumeWithRetry(ctx context.Context, exchange string, queue string, routingKey string, prefetchCount int, handler func(data T) error, dlxTable amqp091.Table) {
	go func() {
		closeChan := rabbitMqSub.channel.NotifyClose(make(chan *amqp091.Error, 1))

		for {
			err := rabbitMqSub.startConsume(ctx, exchange, queue, routingKey, prefetchCount, handler, dlxTable)
			if err != nil {
				log.Errorf("Start comsumer on %v for %v of %v failed: %v, Retry in 5s...", queue, routingKey, exchange, err.Error())
				time.Sleep(5 * time.Second)

				for {
					newCh, chErr := rabbitmqclient.RabbitMQClientConnInstance.NewChannel()
					if chErr != nil {
						log.Errorf("Create new channel failed: %v. Retry in 5s...", chErr.Error())
						time.Sleep(5 * time.Second)
						continue
					}
					rabbitMqSub.channel = newCh

					break
				}

				continue
			}

			log.Infof("Start comsumer on %v for %v of %v successful", queue, routingKey, exchange)

			select {
			case <-ctx.Done():
				rabbitMqSub.channel.Close()
				log.Infof("Context canceled, stop consumer on %v for %v of %v", queue, routingKey, exchange)
				return
			case rabbitErr := <-closeChan:
				if rabbitErr != nil {
					log.Errorf("Channel of consumer on %v for %v of %v closed: %v. Retry in 5s...", queue, routingKey, exchange, rabbitErr.Error())
				} else {
					log.Warnf("Channel of consumer on %v for %v of %v closed cleanly. Retry in 5s...", queue, routingKey, exchange)
				}
				time.Sleep(5 * time.Second)

				for {
					newCh, chErr := rabbitmqclient.RabbitMQClientConnInstance.NewChannel()
					if chErr != nil {
						log.Errorf("Create new channel failed: %v. Retry in 5s...", chErr.Error())
						time.Sleep(5 * time.Second)
						continue
					}
					rabbitMqSub.channel = newCh

					closeChan = rabbitMqSub.channel.NotifyClose(make(chan *amqp091.Error, 1))
					break
				}
			}
		}
	}()
}

func (rabbitMqSub *RabbitMqSub[T]) startConsume(ctx context.Context, exchange string, queue string, routingKey string, prefetchCount int, handler func(data T) error, dlxTable amqp091.Table) error {
	if err := rabbitMqSub.channel.Qos(prefetchCount, 0, false); err != nil {
		log.Errorf("Set QoS for channel failed: %v", err.Error())
		return err
	}

	if _, err := rabbitMqSub.channel.QueueDeclare(queue, true, false, false, false, dlxTable); err != nil {
		log.Errorf("Declare queue %v for %v of %v failed: %v", queue, routingKey, exchange, err.Error())
		return err
	}

	if err := rabbitMqSub.channel.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
		log.Errorf("Bind queue %v for %v of %v failed: %v", queue, routingKey, exchange, err.Error())
		return err
	}

	ch, err := rabbitMqSub.channel.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		log.Errorf("Start consumer on %v for %v of %v failed: %v", queue, routingKey, exchange, err.Error())
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case message, ok := <-ch:
				if !ok {
					log.Errorf("Channel of consumer on %v for %v of %v closed", queue, routingKey, exchange)
					return
				}

				log.Infof("Consume from %v of %v successful", message.RoutingKey, message.Exchange)

				var value T
				t := reflect.TypeOf(value)

				var instance any
				if t.Kind() == reflect.Ptr {
					// T is pointer to struct: create *Struct
					instance = reflect.New(t.Elem()).Interface()
				} else {
					// T is value: create pointer to value (e.g., *int, *string)
					instance = reflect.New(t).Interface()
				}

				if err := json.Unmarshal([]byte(message.Body), instance); err != nil {
					log.Errorf("Unmarshal %v failed: %v", message.Body, err.Error())
					message.Nack(false, false) // Xử lý bị lỗi sẽ không requeue mà đưa vào unacked list
					continue
				}

				var data T
				if t.Kind() == reflect.Ptr {
					// T is pointer already
					data = instance.(T)
				} else {
					// T is value, dereference pointer
					data = reflect.ValueOf(instance).Elem().Interface().(T)
				}

				if err := handler(data); err != nil {
					log.Errorf("Handle message failed: %v", err.Error())
					message.Nack(false, false) // Xử lý bị lỗi sẽ không requeue mà đưa vào unacked list
				} else {
					log.Infof("Handle message successful")
					message.Ack(false)
				}
			}
		}
	}()

	return nil
}
