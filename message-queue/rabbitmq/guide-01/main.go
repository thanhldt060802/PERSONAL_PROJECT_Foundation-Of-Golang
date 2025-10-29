package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"thanhldt060802/common/pubsub"
	"thanhldt060802/internal/rabbitmqclient"
	"thanhldt060802/model"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

var EXAMPLE_NUM = 1
var EXAMPLES map[int]func()

func init() {
	EXAMPLES = map[int]func(){
		1: Example1,
		2: Example2,
	}
}

func main() {

	EXAMPLES[EXAMPLE_NUM]()

}

func Example1() {
	mainExchange := "ldtt-exchange"
	mainQueue := "ldtt-service-handle-for-" + mainExchange
	mainRoutingKey := "service.ldtt.*.create"
	// mainParamsRoutingKey := "service.ldtt.%v.create"

	dlxExchange := "ldtt-dlx-exchange"
	dlxQueue := mainQueue + ".dlx"
	dlxRoutingKey := mainQueue // Trùng với tên queue của main queue

	rabbitmqclient.RabbitMQClientConnInstance = rabbitmqclient.NewRabbitMQClient(rabbitmqclient.RabbitMQConfig{
		Host:     "pitel-cx.dev.tel4vn.com",
		Port:     5672,
		Username: "pitel",
		Password: "Pitel@8229",
	})
	if err := rabbitmqclient.RabbitMQClientConnInstance.DeclareExchange(mainExchange, "topic"); err != nil {
		log.Fatalf("Failed to declare new exchange: %v", err.Error())
	}
	if err := rabbitmqclient.RabbitMQClientConnInstance.DeclareExchange(dlxExchange, "direct"); err != nil {
		log.Fatalf("Failed to declare new dlx exchange: %v", err.Error())
	}
	if rabbitMqPub, err := pubsub.NewRabbitMqPub[string](); err != nil {
		log.Fatalf("Failed to create new publisher: %v", err.Error())
	} else {
		pubsub.RabbitMqPubInstance1 = rabbitMqPub
	}
	if rabbitMqDlx, err := pubsub.NewRabbitMqDlx[any](); err != nil {
		log.Fatalf("Failed to create new dlx subscriber: %v", err.Error())
	} else {
		pubsub.RabbitMqDlxInstance1 = rabbitMqDlx
	}
	if rabbitMqSub, err := pubsub.NewRabbitMqSub[string](); err != nil {
		log.Fatalf("Failed to create new subscriber: %v", err.Error())
	} else {
		pubsub.RabbitMqSubInstance1 = rabbitMqSub
	}

	pubsub.RabbitMqDlxInstance1.ConsumeWithRetry(context.Background(), dlxExchange, dlxQueue, dlxRoutingKey, 1, nil)

	go func() {
		count := 0

		dlxTable := amqp091.Table{
			"x-dead-letter-exchange":    dlxExchange,
			"x-dead-letter-routing-key": mainQueue,
		}

		pubsub.RabbitMqSubInstance1.ConsumeWithRetry(context.Background(), mainExchange, mainQueue, mainRoutingKey, 1, func(data string) error {
			fmt.Println(data)

			count++
			fmt.Printf("Count: %v\n", count)

			time.Sleep(1 * time.Second)

			if rand.IntN(2) == 0 {
				return fmt.Errorf("error simulation")
			}

			return nil
		}, dlxTable)
	}()

	// go func() {
	// 	for i := 1; ; i++ {
	// 		data := fmt.Sprintf("my-payload-%v", i)
	// 		pubsub.RabbitMqPubInstance1.PublishWithRetry(context.Background(), mainExchange, fmt.Sprintf(mainParamsRoutingKey, i), data)
	// 		// go pubsub.RabbitMqPubInstance1.PublishWithRetry(context.Background(), mainExchange, fmt.Sprintf(mainParamsRoutingKey, i), data)

	// 		time.Sleep(200 * time.Millisecond)
	// 	}
	// }()

	select {}
}

func Example2() {
	mainExchange := "ldtt-exchange"
	mainQueue := "ldtt-service-handle-for-" + mainExchange
	mainRoutingKey := "service.ldtt.*.create"
	mainParamsRoutingKey := "service.ldtt.%v.create"

	dlxExchange := "ldtt-dlx-exchange"
	dlxQueue := mainQueue + ".dlx"
	dlxRoutingKey := mainQueue // Trùng với tên queue của main queue

	rabbitmqclient.RabbitMQClientConnInstance = rabbitmqclient.NewRabbitMQClient(rabbitmqclient.RabbitMQConfig{
		Host:     "pitel-cx.dev.tel4vn.com",
		Port:     5672,
		Username: "pitel",
		Password: "Pitel@8229",
	})
	if err := rabbitmqclient.RabbitMQClientConnInstance.DeclareExchange(mainExchange, "topic"); err != nil {
		log.Fatalf("Failed to declare new exchange: %v", err.Error())
	}
	if err := rabbitmqclient.RabbitMQClientConnInstance.DeclareExchange(dlxExchange, "direct"); err != nil {
		log.Fatalf("Failed to declare new dlx exchange: %v", err.Error())
	}
	if rabbitMqPub, err := pubsub.NewRabbitMqPub[*model.DataStruct](); err != nil {
		log.Fatalf("Failed to create new publisher: %v", err.Error())
	} else {
		pubsub.RabbitMqPubInstance2 = rabbitMqPub
	}
	if rabbitMqDlx, err := pubsub.NewRabbitMqDlx[any](); err != nil {
		log.Fatalf("Failed to create new dlx subscriber: %v", err.Error())
	} else {
		pubsub.RabbitMqDlxInstance2 = rabbitMqDlx
	}
	if rabbitMqSub, err := pubsub.NewRabbitMqSub[*model.DataStruct](); err != nil {
		log.Fatalf("Failed to create new subscriber: %v", err.Error())
	} else {
		pubsub.RabbitMqSubInstance2 = rabbitMqSub
	}

	pubsub.RabbitMqDlxInstance2.ConsumeWithRetry(context.Background(), dlxExchange, dlxQueue, dlxRoutingKey, 1, nil)

	go func() {
		count := 0

		dlxTable := amqp091.Table{
			"x-dead-letter-exchange":    dlxExchange,
			"x-dead-letter-routing-key": mainQueue,
		}

		pubsub.RabbitMqSubInstance2.ConsumeWithRetry(context.Background(), mainExchange, mainQueue, mainRoutingKey, 1, func(data *model.DataStruct) error {
			fmt.Println(data)

			count++
			fmt.Printf("Count: %v\n", count)

			time.Sleep(1 * time.Second)

			if rand.IntN(2) == 0 {
				return fmt.Errorf("error simulation")
			}

			return nil
		}, dlxTable)
	}()

	go func() {
		for i := 1; ; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			data := model.DataStruct{
				Field1: fmt.Sprintf("my-payload-%v", i),
				Field2: rand.Int32(),
				Field3: rand.Int64(),
				Field4: rand.Float32(),
				Field5: rand.Float64(),
				Field6: time.Now(),
				Field7: model.SubDataStruct{
					Field1: uuid.New().String(),
					Field2: rand.Int32(),
					Field3: rand.Int64(),
				},
			}
			pubsub.RabbitMqPubInstance2.PublishWithRetry(ctx, mainExchange, fmt.Sprintf(mainParamsRoutingKey, i), &data)

			cancel()
			time.Sleep(200 * time.Millisecond)
		}
	}()

	select {}
}
