package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type RateLimiter struct {
	tokenCh  chan struct{} // Channel for managing tokens
	stopCh   chan struct{} // Channel to stop the token replenishment
	capacity int           // Maximum number of tokens in the bucket
	rate     float64       // Tokens per second
	wg       sync.WaitGroup // Wait group to synchronize goroutines
}

// NewRateLimiter initializes a new RateLimiter with the given capacity and rate
func NewRateLimiter(capacity int, rate float64) *RateLimiter {
	rand.Seed(time.Now().UnixNano())
	limiter := &RateLimiter{
		tokenCh:  make(chan struct{}, capacity),
		stopCh:   make(chan struct{}),
		capacity: capacity,
		rate:     rate,
	}

	// Start the goroutine to replenish tokens
	go limiter.replenishTokens()
	return limiter
}

// replenishTokens continuously adds tokens to the bucket at the specified rate
func (rl *RateLimiter) replenishTokens() {
	defer rl.wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Calculate the number of tokens to add
			tokensToAdd := int(rl.rate)
			for tokensToAdd > 0 && len(rl.tokenCh) < rl.capacity {
				rl.tokenCh <- struct{}{}
				tokensToAdd--
			}
		case <-rl.stopCh:
			return
		}
	}
}

// Acquire attempts to acquire a token, blocking if the bucket is empty
func (rl *RateLimiter) Acquire() {
	rl.wg.Add(1)
	defer rl.wg.Done()
	<-rl.tokenCh // Wait for a token
}

// Release returns a token to the bucket
func (rl *RateLimiter) Release() {
	select {
	case rl.tokenCh <- struct{}{}: // Return the token
	default:
		// Bucket is full, do nothing (this should rarely happen with correct rate calculation)
	}
}

// Shutdown stops the token replenishment goroutine
func (rl *RateLimiter) Shutdown() {
	close(rl.stopCh)
	rl.wg.Wait()
}

func processTask(id int, limiter *RateLimiter) {
	limiter.Acquire() // Acquire a token to process the task
	defer limiter.Release()

	fmt.Printf("Processing task %d\n", id)
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	fmt.Printf("Completed task %d\n", id)
}

func main() {
	const (
		taskCount    = 50
		bucketSize   = 5
		ratePerSec   = 10.0 // 10 tasks per second
		workerCount = 5
	)

	limiter := NewRateLimiter(bucketSize, ratePerSec)
	defer limiter.Shutdown()

	var wg sync.WaitGroup

	// Create a pool of worker goroutines
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 1; i <= taskCount; i++ {
				processTask(i, limiter)
			}
		}()
	}

	wg.Wait()
	fmt.Println("All tasks processed.")
}