package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"thanhldt060802/common/pubsub"
	"thanhldt060802/internal/redisclient"
	"thanhldt060802/model"
	"time"

	"github.com/google/uuid"
)

var EXAMPLE_NUM = 2
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

// Example for Subscribe() and Publish().
// Subscribe() will listen for Payloads sent over TCP Socket.
// Publish() will send Payloads to TCP Socket.
func Example1() {
	redisclient.RedisClientConnInstance = redisclient.NewRedisClient(redisclient.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Database: 0,
		Password: "12345678",
	})
	pubsub.RedisPubInstance1 = pubsub.NewRedisPub[string](redisclient.RedisClientConnInstance.GetClient())
	pubsub.RedisSubInstance1 = pubsub.NewRedisSub[string](redisclient.RedisClientConnInstance.GetClient())

	go func() {
		count := 0

		pubsub.RedisSubInstance1.Subscribe(context.Background(), "my-channel", func(data string) {
			fmt.Println(data)

			count++
			fmt.Printf("Count: %v\n", count)

			time.Sleep(1 * time.Second)
		})
	}()

	go func() {
		for i := 1; ; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			data := fmt.Sprintf("my-payload-%v", i)
			if err := pubsub.RedisPubInstance1.Publish(ctx, "my-channel", data); err != nil {
				cancel()
				return
			}

			cancel()
			time.Sleep(200 * time.Millisecond)
		}
	}()

	select {}
}

// Ref: Example1(), use data struct
func Example2() {
	redisclient.RedisClientConnInstance = redisclient.NewRedisClient(redisclient.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Database: 0,
		Password: "12345678",
	})
	pubsub.RedisPubInstance2 = pubsub.NewRedisPub[*model.DataStruct](redisclient.RedisClientConnInstance.GetClient())
	pubsub.RedisSubInstance2 = pubsub.NewRedisSub[*model.DataStruct](redisclient.RedisClientConnInstance.GetClient())

	go func() {
		count := 0

		pubsub.RedisSubInstance2.Subscribe(context.Background(), "my-channel", func(data *model.DataStruct) {
			fmt.Println(*data)

			count++
			fmt.Printf("Count: %v\n", count)

			time.Sleep(1 * time.Second)
		})
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
			if err := pubsub.RedisPubInstance2.Publish(ctx, "my-channel", &data); err != nil {
				cancel()
				return
			}

			cancel()
			time.Sleep(200 * time.Millisecond)
		}
	}()

	select {}
}
