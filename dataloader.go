package main

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
	fmt.Println(1234)
	for _, request := range requests {
		wg := sync.WaitGroup{}
		loader := NewDataloader(WhereIn)
		for _, routine := range request {
			wg.Add(1)
			loader.rwmutex.RLock()
			a := routine
			go func() {
				a(loader)
				loader.rwmutex.RUnlock()
				wg.Done()
			}()
		}
		go z(loader)
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

type cacheStruct struct {
	value interface{}
	chans []chan interface{}
}

type Loader struct {
	rwmutex sync.RWMutex
	mutex   sync.Mutex
	caches  map[string]cacheStruct
	fn      func([]string) []interface{}
}

func NewDataloader(fn func([]string) []interface{}) *Loader {
	return &Loader{
		caches: make(map[string]cacheStruct),
		fn:     fn,
	}
}

func (loader *Loader) Load(id string) chan interface{} {
	loader.rwmutex.RUnlock()
	loader.mutex.Lock()
	channel := make(chan interface{}, 1)
	cache, ok := loader.caches[id]
	// value, ok := loader.caches[id].value
	if ok {
		if cache.value != nil {
			// 如果不给 channel 缓存，还得新创建一个 goroutine，在里面给 channel 添值，不然没法返回 channel
			channel <- cache.value
			loader.mutex.Unlock()
			loader.rwmutex.RLock()
			return channel
		}
		cache.chans = append(cache.chans, channel)
		loader.mutex.Unlock()
		loader.rwmutex.RLock()
		return channel
	}
	cacheChans := make([]chan interface{}, 0)
	cacheChans = append(cacheChans, channel)
	loader.caches[id] = cacheStruct{
		chans: cacheChans,
	}
	loader.mutex.Unlock()
	loader.rwmutex.RLock()
	return channel
}

func y1(loader *Loader) {
	fmt.Println("something y1")
	value := (<-loader.Load("1")).(string)
	fmt.Println("1" + value)
	// value2 := (<-loader.Load("2")).(string)
	// fmt.Println("2" + value2)
}

func y2(loader *Loader) {
	fmt.Println("something y2")
	value := (<-loader.Load("2")).(string)
	fmt.Println("2" + value)
}

func y3(loader *Loader) {
	fmt.Println("something y3")
}

func z(loader *Loader) {
	for {
		loader.rwmutex.Lock()
		loader.mutex.Lock()
		ids := make([]string, 0)
		// fmt.Println(len(ids))
		// fmt.Println(loader.caches)
		for id, cache := range loader.caches {
			// fmt.Println(len(cache.chans))
			if len(cache.chans) > 0 {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			loader.mutex.Unlock()
			loader.rwmutex.Unlock()
			break
		}
		// fmt.Println(ids)
		// fmt.Println(len(ids))
		values := loader.fn(ids)
		// fmt.Println(values)
		for index, id := range ids {
			cache := loader.caches[id]
			cache.value = values[index]
			for _, channel := range cache.chans {
				channel <- values[index]
			}
			cache.chans = cache.chans[:0]
		}
		loader.mutex.Unlock()
		loader.rwmutex.Unlock()
	}
}
