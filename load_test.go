package main

import (
	"conserver/pkg/mysql"
	"conserver/pkg/redis"
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/rand"
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"
)

func insert() {
	intn := rand.Intn(100000000)
	price := rand.Intn(1000)
	amount := rand.Intn(1000)
	_, err := http.Get(fmt.Sprintf("http://127.0.0.1:8888/order/test/8/%d/1/%d/%d", intn, price, amount))
	if err != nil {
		panic(err)
	}
}

func BenchmarkMatch(b *testing.B) {
	worker := 2
	runTime := 30 * time.Second
	ctx, cancelFunc := context.WithTimeout(context.Background(), runTime)
	defer cancelFunc()

	var wg sync.WaitGroup
	wg.Add(worker)

	// 用于记录每次insert操作的持续时间
	var mu sync.Mutex
	insertTimes := make([]int64, 100000)
	s := time.Now()
	counter := 0
	for i := 0; i < worker; i++ {
		go func() {
			//tempCounter := 0
			tempTime := make([]int64, 100000)
			tempCounter := 0
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					// ctx过期时退出
					mu.Lock()
					insertTimes = append(insertTimes, tempTime...)
					counter += tempCounter
					mu.Unlock()
					return
				default:
					start := time.Now()
					insert()
					tempCounter += 1
					duration := time.Since(start).Milliseconds()
					fmt.Println(duration)
					//
					//mu.Lock()
					tempTime = append(tempTime, duration)
					//counter += 1
					//mu.Unlock()
				}
			}
		}()
	}

	wg.Wait() // 等待所有goroutine完成
	total := time.Since(s).Milliseconds()
	sort.Slice(insertTimes, func(i, j int) bool {
		return insertTimes[i] < insertTimes[j]
	})
	fmt.Printf("插入总数：%d, 运行时间：%d, p50: %d, p90: %d, p99: %d, avg: %d\n", counter, total, insertTimes[len(insertTimes)/2], insertTimes[int(float64(len(insertTimes))*0.9)], insertTimes[int(float64(len(insertTimes))*0.99)], int(float64(total)/float64(counter)))
}

func TestMySQL(t *testing.T) {
	ch := make(chan int)
	mysql.GetInstancePool().Init(&mysql.InstancePoolConfig{
		PoolSize:     5,
		InitInstance: nil,
	})
	<-ch
}

func TestRedis(t *testing.T) {
	ch := make(chan int)
	redis.GetInstancePool().Init(&redis.InstancePoolConfig{
		PoolSize:     4,
		InitInstance: nil,
	})
	<-ch
}
