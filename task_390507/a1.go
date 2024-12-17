package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type RateLimiter struct {
	c        chan struct{}
	done     chan struct{}
	interval time.Duration
	capacity int
}

func NewRateLimiter(capacity int, interval time.Duration) *RateLimiter {
	c := make(chan struct{}, capacity)
	for i := 0; i < capacity; i++ {
		c <- struct{}{} // Fill the bucket to capacity
	}

	limiter := &RateLimiter{
		c:        c,
		done:     make(chan struct{}),
		interval: interval,
		capacity: capacity,
	}

	go limiter.leak()
	return limiter
}

func (rl *RateLimiter) leak() {
	ticker := time.NewTicker(rl.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			select {
			case rl.c <- struct{}{}: // Refill the bucket at defined intervals
			default:
				// Bucket is full; do nothing
			}
		case <-rl.done:
			return
		}
	}
}

func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.c:
		return true // Controlled task allowed
	default:
		return false // Rate limit reached; request denied
	}
}

// Shutdown the leaky bucket
func (rl *RateLimiter) Shutdown() {
	close(rl.done)
}

func processTask(id int) {
	fmt.Printf("Processing task %d\n", id)
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	fmt.Printf("Completed task %d\n", id)
}

func taskWorker(id int, rateLimiter *RateLimiter, wg *sync.WaitGroup) {
	defer wg.Done()
	if rateLimiter.Allow() {
		processTask(id)
	} else {
		fmt.Printf("Task %d rejected due to rate limiting\n", id)
	}
}

func main() {
	const (
		taskCount  = 20
		bucketSize = 5
		interval   = 500 * time.Millisecond // Rate of leak
	)

	rateLimiter := NewRateLimiter(bucketSize, interval)
	var wg sync.WaitGroup

	for i := 1; i <= taskCount; i++ {
		wg.Add(1)
		go taskWorker(i, rateLimiter, &wg)
	}

	wg.Wait()
	rateLimiter.Shutdown()
	fmt.Println("All tasks processed.")
}