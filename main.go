package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"main/lock"
	"sync"
	"time"
)

//OP nums: 10000
//[LIST] expect: 2000 get: 2000 use time: 12.7059663
//[NX] expect: 2000 get: 2000 use time: 16.3814946

func main() {
	// 准备工作
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // redis地址
		Password: "",               // redis密码，没有则留空
		DB:       0,                // 默认数据库，默认是0
	})
	varKey := "balanceKey"
	varLock := "balanceLock"
	plus, minus := 6000, 4000
	fmt.Println("OP nums:", plus+minus)
	var wg sync.WaitGroup
	if err := redisClient.Set(varKey, 0, 0).Err(); err != nil {
		return
	}
	// 测试ListMutex
	listMutex1, err := lock.NewListMutex(redisClient, varLock, 0)
	if err != nil {
		panic(err)
	}
	listMutex2, err := lock.NewListMutex(redisClient, varLock, 0)
	if err != nil {
		panic(err)
	}
	before := time.Now()
	wg.Add(2)
	go worker(redisClient, listMutex1, &wg, plus, 1, varKey)
	go worker(redisClient, listMutex2, &wg, minus, -1, varKey)
	wg.Wait()
	v, err := redisClient.Get(varKey).Result()
	if err != nil {
		panic(err)
	}
	after := time.Now()
	fmt.Println("[LIST] expect:", plus-minus, "get:", v, "use time:", after.Sub(before).Seconds())

	// flush
	//if err = redisClient.FlushDB().Err(); err != nil {
	//	return
	//}
	if err = redisClient.Set(varKey, 0, 0).Err(); err != nil {
		return
	}

	// 测试NxMutex
	nxMutex1, err := lock.NewNxMutex(redisClient, varLock, 0)
	if err != nil {
		panic(err)
	}
	nxMutex2, err := lock.NewNxMutex(redisClient, varLock, 0)
	if err != nil {
		panic(err)
	}

	before = time.Now()
	wg.Add(2)
	go worker(redisClient, nxMutex1, &wg, plus, 1, varKey)
	go worker(redisClient, nxMutex2, &wg, minus, -1, varKey)
	wg.Wait()
	v, err = redisClient.Get(varKey).Result()
	if err != nil {
		panic(err)
	}
	after = time.Now()
	fmt.Println("[NX] expect:", plus-minus, "get:", v, "use time:", after.Sub(before).Seconds())
}
