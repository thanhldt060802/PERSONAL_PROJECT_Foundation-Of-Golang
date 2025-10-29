package queuedisk

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"thanhldt060802/model"
	"time"

	"github.com/dgraph-io/badger/v4"
	log "github.com/sirupsen/logrus"
)

var QueueDiskInstance1 IQueueDisk[string]
var QueueDiskInstance2 IQueueDisk[*model.DataStruct]

type QueueDisk[T any] struct {
	db      *badger.DB
	counter int64
}

type IQueueDisk[T any] interface {
	Enqueue(data T) error
	Dequeue() (T, error)
	Close() error
}

func NewQueueDisk[T any](path string) IQueueDisk[T] {
	opts := badger.DefaultOptions(path)
	// opts.WithSyncWrites(true)  // No effect on Window
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}

	qd := &QueueDisk[T]{
		db:      db,
		counter: 0,
	}
	go qd.garbageCollection()

	return qd
}

func (qd *QueueDisk[T]) garbageCollection() {
	if err := qd.db.RunValueLogGC(0.5); err != nil && err != badger.ErrNoRewrite {
		log.Errorf("GC error: %v", err)
	}

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := qd.db.RunValueLogGC(0.5); err != nil && err != badger.ErrNoRewrite {
			log.Errorf("GC error: %v", err)
		}
	}
}

func (qd *QueueDisk[T]) Enqueue(data T) error {
	key := []byte(fmt.Sprintf("%020d", qd.counter))
	qd.counter++

	payload, err := json.Marshal(data)
	if err != nil {
		log.Errorf("Marshal data failed: %v", err.Error())
		return err
	}

	return qd.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, payload)
	})
}

func (qd *QueueDisk[T]) Dequeue() (T, error) {
	var keyToDelete []byte
	var data T

	err := qd.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
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
				data = instance.(T)
			} else {
				// T is value, dereference pointer
				data = reflect.ValueOf(instance).Elem().Interface().(T)
			}
			keyToDelete = k

			break
		}

		if keyToDelete == nil {
			return errors.New("queue empty")
		}

		return txn.Delete(keyToDelete)
	})

	return data, err
}

func (qd *QueueDisk[T]) Close() error {
	return qd.db.Close()
}
