package queuedisk

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"thanhldt060802/model"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dgraph-io/badger/v4"
)

var BatchQueueDiskInstance1 IBatchQueueDisk[string]
var BatchQueueDiskInstance2 IBatchQueueDisk[*model.DataStruct]

type BatchQueueDisk[T any] struct {
	db      *badger.DB
	counter int64

	batchSize int

	batchEnqueue            []T
	currentBatchEnqueueSize int

	batchDequeue []T
}

type IBatchQueueDisk[T any] interface {
	Enqueue(data T) error
	Dequeue() ([]T, error)
	Close() error
}

func NewBatchQueueDisk[T any](path string, batchSize int) IBatchQueueDisk[T] {
	opts := badger.DefaultOptions(path)
	// opts.WithSyncWrites(true)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}

	bqd := &BatchQueueDisk[T]{
		db:      db,
		counter: 0,

		batchSize: batchSize,

		batchEnqueue:            make([]T, batchSize),
		currentBatchEnqueueSize: 0,

		batchDequeue: make([]T, batchSize),
	}
	go bqd.GarbageCollection()

	return bqd
}

func (bqd *BatchQueueDisk[T]) GarbageCollection() {
	if err := bqd.db.RunValueLogGC(0.5); err != nil && err != badger.ErrNoRewrite {
		log.Printf("GC error: %v", err)
	}

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := bqd.db.RunValueLogGC(0.5); err != nil && err != badger.ErrNoRewrite {
			log.Printf("GC error: %v", err)
		}
	}
}

func (bqd *BatchQueueDisk[T]) Close() error {
	return bqd.db.Close()
}

func (bqd *BatchQueueDisk[T]) Enqueue(data T) error {
	bqd.batchEnqueue[bqd.currentBatchEnqueueSize] = data
	bqd.currentBatchEnqueueSize++

	if bqd.currentBatchEnqueueSize >= bqd.batchSize {
		return bqd.db.Update(func(txn *badger.Txn) error {
			for _, dataEnq := range bqd.batchEnqueue {
				key := []byte(fmt.Sprintf("%020d", bqd.counter))
				bqd.counter++

				payload, err := json.Marshal(dataEnq)
				if err != nil {
					log.Errorf("Marshal data failed: %v", err.Error())
					return err
				}

				if err := txn.Set(key, payload); err != nil {
					return err
				}
			}

			bqd.currentBatchEnqueueSize = 0

			return nil
		})
	}

	return nil
}

func (bqd *BatchQueueDisk[T]) Dequeue() ([]T, error) {
	var keysToDelete [][]byte
	var dataDeqs []T

	err := bqd.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		currentBatchDequeueSize := 0
		for it.Rewind(); it.Valid() && currentBatchDequeueSize < bqd.batchSize; it.Next() {
			item := it.Item()
			k := item.KeyCopy(nil)
			v, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

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

			if err := json.Unmarshal([]byte(v), instance); err != nil {
				log.Errorf("Unmarshal %v failed: %v", v, err.Error())
				continue
			}

			if t.Kind() == reflect.Ptr {
				// T is pointer already
				bqd.batchDequeue[currentBatchDequeueSize] = instance.(T)
			} else {
				// T is value, dereference pointer
				bqd.batchDequeue[currentBatchDequeueSize] = reflect.ValueOf(instance).Elem().Interface().(T)
			}
			keysToDelete = append(keysToDelete, k)

			currentBatchDequeueSize++
		}
		dataDeqs = bqd.batchDequeue[:currentBatchDequeueSize]

		if len(dataDeqs) == 0 {
			if bqd.currentBatchEnqueueSize > 0 {
				copyEnqueueRemaining := make([]T, bqd.currentBatchEnqueueSize)
				copy(copyEnqueueRemaining, bqd.batchEnqueue[:bqd.currentBatchEnqueueSize])
				dataDeqs = copyEnqueueRemaining
				bqd.currentBatchEnqueueSize = 0

				return nil
			}

			return errors.New("queue empty")
		}

		for _, key := range keysToDelete {
			if err := txn.Delete(key); err != nil {
				return err
			}
		}

		return nil
	})

	return dataDeqs, err
}
