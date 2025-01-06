package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type RateLimiter struct {
	c           chan struct{}
	done        chan struct{}
	interval    time.Duration
	capacity    int
	mu          sync.Mutex
	queueSize   atomic.Int32
	requestFreq atomic.Int32
}

func NewRateLimiter(initialCapacity int, interval time.Duration) *RateLimiter {
	c := make(chan struct{}, initialCapacity)
	for i := 0; i < initialCapacity; i++ {
		c <- struct{}{}
	}

	limiter := &RateLimiter{
		c:           c,
		done:        make(chan struct{}),
		interval:    interval,
		capacity:    initialCapacity,
		queueSize:   atomic.NewInt32(0),
		requestFreq: atomic.NewInt32(0),
	}

	go limiter.leak()
	go limiter.monitorMetrics()
	return limiter
}

func (rl *RateLimiter) leak() {
	ticker := time.NewTicker(rl.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			// Scale capacity based on queue size and request frequency
			currentCapacity := rl.capacity
			if atomic.LoadInt32(&rl.queueSize) > 10 {
				currentCapacity = max(currentCapacity-1, 1)
			} else if atomic.LoadInt32(&rl.requestFreq) > 100 {
				currentCapacity += 1
			}
			if currentCapacity != rl.capacity {
				rl.capacity = currentCapacity
				fmt.Printf("Scaling bucket capacity to %d\n", currentCapacity)
				for i := 0; i < currentCapacity-rl.capacity; i++ {
					rl.c <- struct{}{}
				}
				for i := 0; i < rl.capacity-currentCapacity; i++ {
					<-rl.c
				}
			}
			rl.mu.Unlock()

		case <-rl.done:
			return
		}
	}
}

func (rl *RateLimiter) Allow(priority int) bool {
	rl.queueSize.Inc()
	defer rl.queueSize.Dec()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if priority == 1 { // High priority: skip the queue
		select {
		case rl.c <- struct{}{}:
			return true
		default:
			return false
		}
	}

	select {
	case <-rl.c:
		return true
	default:
		return false
	}
}

func (rl *RateLimiter) monitorMetrics() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			atomic.StoreInt32(&rl.requestFreq, 0)
			fmt.Printf("Queue Size: %d, Request Frequency: %d\n", atomic.LoadInt32(&rl.queueSize), rl.requestFreq)
		case <-rl.done:
			return
		}
	}
}

func processTask(id int, priority int, rateLimiter *RateLimiter, wg *sync.WaitGroup) {
	defer wg.Done()
	if rateLimiter.Allow(priority) {
		start := time.Now()
		fmt.Printf("Processing task %d (priority %d)\n", id, priority)
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		fmt.Printf("Completed task %d (priority %d) in %s\n", id, priority, time.Since(start))

		atomic.AddInt32(&rateLimiter.requestFreq, 1)
	} else {
		fmt.Printf("Task %d (priority %d) rejected due to rate limiting\n", id, priority)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	const (
		taskCount      = 50
		highPriorityCount = 10
		initialCapacity = 5
		interval       = 500 * time.Millisecond
	)

	rateLimiter := NewRateLimiter(initialCapacity, interval)
	var wg sync.WaitGroup

	for i := 1; i <= taskCount; i++ {
		priority := 1
		if i > highPriorityCount {
			priority = 0
		}
		wg.Add(1)
		go processTask(i, priority, rateLimiter, &wg)
	}

	wg.Wait()
	rateLimiter.Shutdown()
	fmt.Println("All tasks processed.")
}