package main

import (
	"fmt"
	"math/rand/v2"
	"thanhldt060802/common/cache"
	"thanhldt060802/model"
	"time"

	"github.com/google/uuid"
)

var EXAMPLE_NUM int = 7
var EXAMPLES map[int]func()

func init() {
	EXAMPLES = map[int]func(){
		1: Example1,
		2: Example2,
		3: Example3,
		4: Example4,
		5: Example5,
		6: Example6,
		7: Example7,
	}
}

func main() {

	EXAMPLES[EXAMPLE_NUM]()

}

// Example for Set() and Get().
// Element is setted by Set() with default TTL is 3 seconds.
// Get() will be called per 1 second (is less than defaultTTL).
func Example1() {
	cache.MemoryCacheInstance1 = cache.NewMemoryCache[string, string](3 * time.Second)

	cache.MemoryCacheInstance1.Set("my-key", "my-value")

	for {
		value, ok := cache.MemoryCacheInstance1.Get("my-key")
		if !ok {
			fmt.Println("my-key doesn't exists in cache")
			return
		}
		fmt.Println("my-key :", value)

		time.Sleep(1 * time.Second)
	}
}

// Example for Set() and Get().
// Element is setted by Set() with default TTL is 3 seconds.
// Get() will be called per 5 seconds (is greater defaultTTL).
func Example2() {
	cache.MemoryCacheInstance1 = cache.NewMemoryCache[string, string](3 * time.Second)

	cache.MemoryCacheInstance1.Set("my-key", "my-value")

	for {
		value, ok := cache.MemoryCacheInstance1.Get("my-key")
		if !ok {
			fmt.Println("my-key doesn't exists in cache")
			return
		}
		fmt.Println("my-key :", value)

		time.Sleep(5 * time.Second)
	}
}

// Example for SetTTL() and Get()
// Element is setted by SetTTL() with other TTL is 5 seconds
// Get() will be called per 1 second (is less than setted TTL)
func Example3() {
	cache.MemoryCacheInstance1 = cache.NewMemoryCache[string, string](3 * time.Second)

	cache.MemoryCacheInstance1.SetTTL("my-key", "my-value", 5*time.Second)

	for {
		value, ok := cache.MemoryCacheInstance1.Get("my-key")
		if !ok {
			fmt.Println("my-key doesn't exists in cache")
			return
		}
		fmt.Println("my-key :", value)

		time.Sleep(1 * time.Second)
	}
}

// Example for SetTTL() and Get().
// Element is setted by SetTTL() with other TTL is 6 seconds.
// Get() will be called per 4 second (is less than setted TTL).
// It means the default TLL is no effect to other setted TTL for element.
func Example4() {
	cache.MemoryCacheInstance1 = cache.NewMemoryCache[string, string](3 * time.Second)

	cache.MemoryCacheInstance1.SetTTL("my-key", "my-value", 6*time.Second)

	for {
		value, ok := cache.MemoryCacheInstance1.Get("my-key")
		if !ok {
			fmt.Println("my-key doesn't exists in cache")
			return
		}
		fmt.Println("my-key :", value)

		time.Sleep(4 * time.Second)
	}
}

// Example for SetTTL() and Get().
// Element is setted by SetTTL() with other TTL is 6 seconds.
// Get() will be called per 10 second (is greater setted TTL).
func Example5() {
	cache.MemoryCacheInstance1 = cache.NewMemoryCache[string, string](3 * time.Second)

	cache.MemoryCacheInstance1.SetTTL("my-key", "my-value", 6*time.Second)

	for {
		value, ok := cache.MemoryCacheInstance1.Get("my-key")
		if !ok {
			fmt.Println("my-key doesn't exists in cache")
			return
		}
		fmt.Println("my-key :", value)

		time.Sleep(10 * time.Second)
	}
}

// Example for Set(), Get() and Del().
// Element is setted by Set() with default TTL is 3 seconds.
// Get() will be called per 1 second (is less than defaultTTL).
// When Get() is called 5 times, the element will be deleted.
func Example6() {
	cache.MemoryCacheInstance1 = cache.NewMemoryCache[string, string](3 * time.Second)

	cache.MemoryCacheInstance1.Set("my-key", "my-value")
	count := 0

	for {
		value, ok := cache.MemoryCacheInstance1.Get("my-key")
		if !ok {
			fmt.Println("my-key doesn't exists in cache")
			return
		}
		fmt.Println("my-key :", value)

		count++
		if count == 5 {
			cache.MemoryCacheInstance1.Del("my-key")
		}
		time.Sleep(1 * time.Second)
	}
}

// Ref: Example6(), use data struct.
func Example7() {
	cache.MemoryCacheInstance2 = cache.NewMemoryCache[string, *model.DataStruct](3 * time.Second)

	data := model.DataStruct{
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
	cache.MemoryCacheInstance2.Set("my-key", &data)
	count := 0

	for {
		value, ok := cache.MemoryCacheInstance2.Get("my-key")
		if !ok {
			fmt.Println("my-key doesn't exists in cache")
			return
		}
		fmt.Println("my-key :", *value)

		count++
		if count == 5 {
			cache.MemoryCacheInstance2.Del("my-key")
		}
		time.Sleep(1 * time.Second)
	}
}
