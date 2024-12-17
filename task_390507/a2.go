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
	ticker   *time.Ticker
	capacity int
}

// NewRateLimiter returns a new RateLimiter with a specified capacity and interval.
func NewRateLimiter(capacity int, interval time.Duration) *RateLimiter {
	c := make(chan struct{}, capacity)
	rl := &RateLimiter{
		c:        c,
		done:     make(chan struct{}),
		ticker:   time.NewTicker(interval),
		capacity: capacity,
	}

	// Fill the bucket to its capacity initially.
	for i := 0; i < capacity; i++ {
		c <- struct{}{}
	}

	go rl.leak()
	return rl
}

// leak manages the rate at which tokens are released into the bucket.
func (rl *RateLimiter) leak() {
	for {
		select {
		case <-rl.ticker.C:
			select {
			case rl.c <- struct{}{}:
				// Successfully added a token back into the bucket
			default:
				// The bucket is full; do nothing
			}
		case <-rl.done:
			rl.ticker.Stop()
			return
		}
	}
}

// Allow returns true if a task can proceed; false otherwise.
func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.c:
		return true
	default:
		return false
	}
}

// Shutdown stops the leaking process and closes the channel.
func (rl *RateLimiter) Shutdown() {
	close(rl.done)
}

// processTask simulates a task processing function.
func processTask(id int) {
	fmt.Printf("Task %d is being processed...\n", id)
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	fmt.Printf("Task %d completed.\n", id)
}

// taskWorker represents a worker that processes tasks with rate limiting.
func taskWorker(id int, rateLimiter *RateLimiter, wg *sync.WaitGroup) {
	defer wg.Done()

	if rateLimiter.Allow() {
		processTask(id)
	} else {
		fmt.Printf("Task %d rejected due to rate limiting.\n", id)
	}
}

func main() {
	const (
		taskCount  = 20
		bucketSize = 5
		interval   = 500 * time.Millisecond // Token refill interval
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