package goloader

import (
	"fmt"
	"sync"
	"time"
)

var requests = [][]func(*Loader){{y1, y2}, {y2, y3}, {y3}}

var data = map[string]string{
	"1": "a",
	"2": "b",
	"3": "c",
}

func main() {
	fmt.Print(123)
	for _, request := range requests {
		wg := sync.WaitGroup{}
		loader := NewDataloader(WhereIn)
		for _, routine := range request {
			wg.Add(1)
			go func() {
				routine(loader)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func WhereIn(ids []string) []interface{} {
	time.Sleep(1000)
	values := make([]interface{}, len(ids))
	for index, id := range ids {
		values[index] = data[id]
	}
	return values
}

type Loader struct {
	cache    map[string]interface{}
	channels map[string][]chan interface{}
	fn       func([]string) []interface{}
}

func NewDataloader(fn func([]string) []interface{}) *Loader {
	return &Loader{
		cache:    make(map[string]interface{}),
		channels: make(map[string][]chan interface{}),
		fn:       fn,
	}
}

func (loader *Loader) Load(id string) chan interface{} {
	channel := make(chan interface{}, 1)
	value, ok := loader.cache[id]
	if ok {
		// 如果不给 channel 缓存，还得新创建一个 goroutine，在里面给 channel 添值，不然没法返回 channel
		channel <- value
		return channel
	}
	loader.channels[id] = make([]chan interface{}, 2)
	loader.channels[id] = append(loader.channels[id], channel)
	return channel
}

func y1(loader *Loader) {
	fmt.Println("something")
	value := (<-loader.Load("1")).(string)
	fmt.Println(value)
	value2 := (<-loader.Load("2")).(string)
	fmt.Println(value2)
}

func y2(loader *Loader) {
	fmt.Println("something")
	value := (<-loader.Load("2")).(string)
	fmt.Println(value)
}

func y3(loader *Loader) {
	fmt.Println("something")
}

func z() {

}
