package main

import (
	"github.com/go-redis/redis"
	"main/lock"
	"strconv"
	"sync"
)

func worker(db *redis.Client, mutex lock.Locker, wg *sync.WaitGroup, times, amount int, varKey string) {
	defer wg.Done()
	for i := 0; i < times; i++ {
		mutex.Lock()
		v, err := db.Get(varKey).Result()
		if err != nil {
			panic(err)
		}
		vInt, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}
		_, err = db.Set(varKey, vInt+amount, 0).Result()
		if err != nil {
			return
		}
		mutex.Unlock()
	}
}
