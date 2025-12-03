package main

import (
	"fmt"
	"math/rand/v2"
	"thanhldt060802/common/queuedisk"
	"thanhldt060802/model"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var EXAMPLE_NUM int = 1
var EXAMPLES map[int]func()

func init() {
	EXAMPLES = map[int]func(){
		1: Example1,
		2: Example2,
		3: Example3,
		4: Example4,
		5: Example5,
		6: Example6,
	}
}

func main() {

	EXAMPLES[EXAMPLE_NUM]()

}

// Example for Enqueue() and Dequeue() with Queue Disk.
func Example1() {
	queuedisk.QueueDiskInstance1 = queuedisk.NewQueueDisk[string]("disk_storage")

	for i := 1; i <= 30; i++ {
		dataEnq := fmt.Sprintf("message %v", i)
		if err := queuedisk.QueueDiskInstance1.Enqueue(dataEnq); err != nil {
			log.Errorf("Enqueue failed: %v", err.Error())
			break
		}
	}

	for {
		dataDeq, err := queuedisk.QueueDiskInstance1.Dequeue()
		if err != nil {
			log.Errorf("Dequeue failed: %v", err.Error())
			break
		}
		fmt.Println(dataDeq)
	}

	queuedisk.QueueDiskInstance1.Close()
}

// Ref: Example1(), use data struct.
func Example2() {
	queuedisk.QueueDiskInstance2 = queuedisk.NewQueueDisk[*model.DataStruct]("disk_storage")

	for i := 1; i <= 30; i++ {
		dataEnq := model.DataStruct{
			Field1: uuid.New().String(),
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
		if err := queuedisk.QueueDiskInstance2.Enqueue(&dataEnq); err != nil {
			log.Errorf("Enqueue failed: %v", err.Error())
			break
		}
	}

	for {
		dataDeq, err := queuedisk.QueueDiskInstance2.Dequeue()
		if err != nil {
			log.Errorf("Dequeue failed: %v", err.Error())
			break
		}
		fmt.Println(*dataDeq)
	}

	queuedisk.QueueDiskInstance2.Close()
}

// Example for Enqueue() and Dequeue() with Queue Disk.
// Calculate time for performance when handle 10000 element.
func Example3() {
	queuedisk.QueueDiskInstance1 = queuedisk.NewQueueDisk[string]("disk_storage")

	{
		dataEnqs := make([]string, 10000)
		for i := 0; i < len(dataEnqs); i++ {
			dataEnqs[i] = fmt.Sprintf("message %v", i)
		}

		count := 0
		startTime := time.Now()
		for _, dataEnq := range dataEnqs {
			if err := queuedisk.QueueDiskInstance1.Enqueue(dataEnq); err != nil {
				log.Errorf("Enqueue failed: %v", err.Error())
				break
			}
			count++
		}
		endTime := time.Now()
		log.Infof("Total time for enqueue %v elements: %v", count, endTime.Sub(startTime))
	}

	{
		count := 0
		startTime := time.Now()
		for {
			_, err := queuedisk.QueueDiskInstance1.Dequeue()
			if err != nil {
				log.Errorf("Dequeue failed: %v", err.Error())
				break
			}
			count++
		}
		endTime := time.Now()
		log.Infof("Total time for dequeue %v elements: %v", count, endTime.Sub(startTime))
	}

	queuedisk.QueueDiskInstance1.Close()
}

// Example for Enqueue() and Dequeue() with Batch Queue Disk.
func Example4() {
	queuedisk.BatchQueueDiskInstance1 = queuedisk.NewBatchQueueDisk[string]("disk_storage", 8)

	for i := 1; i <= 30; i++ {
		dataEnq := fmt.Sprintf("message %v", i)
		if err := queuedisk.BatchQueueDiskInstance1.Enqueue(dataEnq); err != nil {
			log.Errorf("Enqueue failed: %v", err.Error())
			break
		}
	}

	for {
		dataDeqs, err := queuedisk.BatchQueueDiskInstance1.Dequeue()
		if err != nil {
			log.Errorf("Dequeue failed: %v", err.Error())
			break
		}
		for _, dataDeq := range dataDeqs {
			fmt.Println(dataDeq)
		}
	}

	queuedisk.BatchQueueDiskInstance1.Close()
}

// Ref: Example4(), use data struct.
func Example5() {
	queuedisk.BatchQueueDiskInstance2 = queuedisk.NewBatchQueueDisk[*model.DataStruct]("disk_storage", 8)

	for i := 1; i <= 30; i++ {
		dataEnq := model.DataStruct{
			Field1: uuid.New().String(),
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
		if err := queuedisk.BatchQueueDiskInstance2.Enqueue(&dataEnq); err != nil {
			log.Errorf("Enqueue failed: %v", err.Error())
			break
		}
	}

	for {
		dataDeqs, err := queuedisk.BatchQueueDiskInstance2.Dequeue()
		if err != nil {
			log.Errorf("Dequeue failed: %v", err.Error())
			break
		}
		for _, dataDeq := range dataDeqs {
			fmt.Println(*dataDeq)
		}
	}

	queuedisk.BatchQueueDiskInstance2.Close()
}

// Example for Enqueue() and Dequeue() with Batch Queue Disk.
// Calculate time for performance when handle 10000 element.
func Example6() {
	queuedisk.BatchQueueDiskInstance1 = queuedisk.NewBatchQueueDisk[string]("disk_storage", 33)

	{
		dataEnqs := make([]string, 10000)
		for i := 0; i < len(dataEnqs); i++ {
			dataEnqs[i] = fmt.Sprintf("message %v", i)
		}

		count := 0
		startTime := time.Now()
		for _, dataEnq := range dataEnqs {
			if err := queuedisk.BatchQueueDiskInstance1.Enqueue(dataEnq); err != nil {
				log.Errorf("Enqueue failed: %v", err.Error())
				break
			}
			count++
		}
		endTime := time.Now()
		log.Printf("Total time for enqueue %v elements: %v", count, endTime.Sub(startTime))
	}

	{
		count := 0
		startTime := time.Now()
		for {
			dataDeqs, err := queuedisk.BatchQueueDiskInstance1.Dequeue()
			if err != nil {
				log.Errorf("Dequeue failed: %v", err.Error())
				break
			}
			count += len(dataDeqs)
		}
		endTime := time.Now()
		log.Printf("Total time for dequeue %v elements: %v", count, endTime.Sub(startTime))
	}

	queuedisk.BatchQueueDiskInstance1.Close()
}
