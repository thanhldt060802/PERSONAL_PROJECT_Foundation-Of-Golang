package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

/*
SORTED SET
asynq:servers: Ghi lại thông tin Server đang chạy Worker
asynq:workers: Ghi lại thông tin Worker đang chạy trên Server
asynq:{<queue>}:scheduled: Chứa các task đang trong lập lịch mà chưa đạt lập lịch
asynq:{<queue>}:lease: Chứa các task đang được Worker xử lý để tránh các Worker khác lấy nhầm
asynq:{<queue>}:retry: Chứa các task bị lỗi trong quá trình xử lý để bắn về lại cho Worker
asynq:{<queue>}:archived: Chứa các task bị lỗi trong quá trình xử sau số lần max retry đã quy định

SET
asynq:queues: Chứa các khai báo queue chứa task

LIST
asynq:{<queue>}:active: Chứa danh sách các task đang được Worker xử lý
asynq:{<queue>}:pending: Chứa danh sách các task cần xử lý khi đã đạt thời gian lập lịch

STRING
asynq:{<queue>}:processed: Chứa số lượng task đã xử lý thành công
asynq:{<queue>}:failed: Chứa số lượng task đã xử lý thất bại

*** Lưu ý: Đang asynq:{<queue>}:active mà tắt Service thì asynq:servers khác sẽ khoi phục task nếu timeout trong asynq:{<queue>}:lease về asynq:{<queue>}:retry
*/

func main() {

	go func() {
		srv := asynq.NewServer(
			asynq.RedisClientOpt{Addr: "127.0.0.1:6379", Password: "12345678", DB: 1},
			asynq.Config{
				Concurrency: 3, // distributed worker = chạy nhiều instance
				Queues: map[string]int{
					"mytask": 3,
				},
				RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
					return 5 * time.Second
				},
			},
		)

		mux := asynq.NewServeMux()

		// Handler for queue task
		mux.HandleFunc("myqueuetask:hello", func(ctx context.Context, t *asynq.Task) error {
			// time.Sleep(5 * time.Second)
			var data map[string]interface{}
			json.Unmarshal(t.Payload(), &data)
			// if rand.IntN(2) == 0 {
			// 	fmt.Printf("[myqueuetask:hello - task: %s] Payload: %v - FAILED\n", t.ResultWriter().TaskID(), data)
			// 	return errors.New("simulate error")
			// }
			fmt.Printf("[1- myqueuetask:hello - task: %s] Payload: %v - SUCCESS\n", t.ResultWriter().TaskID(), data)
			return nil
		})

		// Handler for schedule task
		mux.HandleFunc("myscheduletask:goodbye", func(ctx context.Context, t *asynq.Task) error {
			// time.Sleep(5 * time.Second)
			var data map[string]interface{}
			json.Unmarshal(t.Payload(), &data)
			// if rand.IntN(2) == 0 {
			// 	fmt.Printf("[myscheduletask:goodbye] Payload: %v - FAILED\n", data)
			// 	return errors.New("simulate error")
			// }
			fmt.Printf("[myscheduletask:goodbye] Payload: %v - SUCCESS\n", data)
			return nil
		})

		log.Println("Worker started...")
		if err := srv.Run(mux); err != nil {
			log.Fatal(err)
		}

		select {}
	}()

	go func() {
		srv := asynq.NewServer(
			asynq.RedisClientOpt{Addr: "127.0.0.1:6379", Password: "12345678", DB: 1},
			asynq.Config{
				Concurrency: 3, // distributed worker = chạy nhiều instance
				Queues: map[string]int{
					"mytask": 3,
				},
				RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
					return 5 * time.Second
				},
			},
		)

		mux := asynq.NewServeMux()

		// Handler for queue task
		mux.HandleFunc("myqueuetask:hello", func(ctx context.Context, t *asynq.Task) error {
			// time.Sleep(5 * time.Second)
			var data map[string]interface{}
			json.Unmarshal(t.Payload(), &data)
			// if rand.IntN(2) == 0 {
			// 	fmt.Printf("[myqueuetask:hello - task: %s] Payload: %v - FAILED\n", t.ResultWriter().TaskID(), data)
			// 	return errors.New("simulate error")
			// }
			fmt.Printf("[2 - myqueuetask:hello - task: %s] Payload: %v - SUCCESS\n", t.ResultWriter().TaskID(), data)
			return nil
		})

		// Handler for schedule task
		mux.HandleFunc("myscheduletask:goodbye", func(ctx context.Context, t *asynq.Task) error {
			// time.Sleep(5 * time.Second)
			var data map[string]interface{}
			json.Unmarshal(t.Payload(), &data)
			// if rand.IntN(2) == 0 {
			// 	fmt.Printf("[myscheduletask:goodbye] Payload: %v - FAILED\n", data)
			// 	return errors.New("simulate error")
			// }
			fmt.Printf("[myscheduletask:goodbye] Payload: %v - SUCCESS\n", data)
			return nil
		})

		log.Println("Worker started...")
		if err := srv.Run(mux); err != nil {
			log.Fatal(err)
		}

		select {}
	}()

	// 1️⃣ Creating a normal job
	go func() {
		log.Println("Test normal task...")

		client := asynq.NewClient(
			asynq.RedisClientOpt{Addr: "127.0.0.1:6379", Password: "12345678", DB: 1},
		)
		defer client.Close()

		{
			count := 1
			for {
				dataBytes, _ := json.Marshal(map[string]interface{}{
					"count": count,
				})
				task := asynq.NewTask("myqueuetask:hello", dataBytes, asynq.Queue("mytask"))
				_, err := client.Enqueue(task)
				if err != nil {
					log.Fatal(err)
				}
				count++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// 2️⃣ Delay job — chạy sau 10 giây
	// go func() {
	// 	log.Println("Test normal delay task...")

	// 	client := asynq.NewClient(
	// 		asynq.RedisClientOpt{Addr: "127.0.0.1:6379", Password: "12345678", DB: 1},
	// 	)
	// 	defer client.Close()

	// 	{
	// 		count := 1
	// 		for {
	// 			dataBytes, _ := json.Marshal(map[string]interface{}{
	// 				"count": count,
	// 			})
	// 			delayTask := asynq.NewTask("myqueuetask:hello", dataBytes, asynq.Queue("mytask"))
	// 			_, err := client.Enqueue(delayTask, asynq.ProcessIn(10*time.Second))
	// 			if err != nil {
	// 				log.Fatal(err)
	// 			}
	// 			count++
	// 			time.Sleep(1 * time.Second)
	// 		}
	// 	}
	// }()

	// 3️⃣ Retry policy (5 lần, backoff mặc định)
	// go func() {
	// 	log.Println("Test normal retry task...")

	// 	client := asynq.NewClient(
	// 		asynq.RedisClientOpt{Addr: "127.0.0.1:6379", Password: "12345678", DB: 1},
	// 	)
	// 	defer client.Close()

	// 	{
	// 		count := 1
	// 		for {
	// 			dataBytes, _ := json.Marshal(map[string]interface{}{
	// 				"count": count,
	// 			})
	// 			retryTask := asynq.NewTask("myqueuetask:hello", dataBytes, asynq.Queue("mytask"))
	// 			_, err := client.Enqueue(retryTask, asynq.ProcessIn(5*time.Second), asynq.MaxRetry(1))
	// 			if err != nil {
	// 				log.Fatal(err)
	// 			}
	// 			count++
	// 			time.Sleep(2 * time.Second)
	// 		}
	// 	}
	// }()

	// 4️⃣ Scheduling with CRON (ví dụ chạy daily)
	// go func() {
	// 	log.Println("Test normal schedule task...")

	// 	scheduler := asynq.NewScheduler(
	// 		asynq.RedisClientOpt{Addr: "127.0.0.1:6379", Password: "12345678", DB: 1},
	// 		&asynq.SchedulerOpts{},
	// 	)

	// 	go func() {
	// 		// Start scheduler in background
	// 		log.Println("Scheduler started...")
	// 		if err := scheduler.Run(); err != nil {
	// 			log.Fatal(err)
	// 		}

	// 		select {}
	// 	}()

	// 	_, err := scheduler.Register("@every 5s", asynq.NewTask("myscheduletask:goodbye", nil, asynq.Queue("mytask"), asynq.MaxRetry(0), asynq.Timeout(1*time.Hour)))
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	select {}

}
