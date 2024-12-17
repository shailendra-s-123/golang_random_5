package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	bucketSize = 5  // Number of tokens in the bucket
	refillRate  = 1 // Tokens added per second
	interval    = time.Second
)

type LeakyBucket struct {
	mu         sync.Mutex
	tokens     int
	lastRefill time.Time
}

func NewLeakyBucket() *LeakyBucket {
	return &LeakyBucket{
		tokens:     bucketSize,
		lastRefill: time.Now(),
	}
}

func (lb *LeakyBucket) Acquire() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Refill the bucket
	now := time.Now()
	delta := now.Sub(lb.lastRefill)
	tokensToAdd := int(delta.Seconds()) * refillRate
	if tokensToAdd > 0 {
		lb.tokens = min(lb.tokens+tokensToAdd, bucketSize)
		lb.lastRefill = now
	}

	// Check if there's a token available
	if lb.tokens > 0 {
		lb.tokens--
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func makeHTTPRequest(url string) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := resp.Body.Bytes()
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}

	fmt.Printf("Response from %s: %s\n", url, strings.TrimSpace(string(body)))
}

func main() {
	lb := NewLeakyBucket()
	urls := []string{
		"https://api.example.com/data1",
		"https://api.example.com/data2",
		"https://api.example.com/data3",
		"https://api.example.com/data4",
		"https://api.example.com/data5",
		"https://api.example.com/data6",
		"https://api.example.com/data7",
		"https://api.example.com/data8",
		"https://api.example.com/data9",
		"https://api.example.com/data10",
	}

	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			for {
				if lb.Acquire() {
					makeHTTPRequest(url)
					time.Sleep(interval) // Simulate request processing time
				} else {
					time.Sleep(time.Millisecond * 100) // Back off slightly
				}
			}
		}(url)
	}

	wg.Wait()
	fmt.Println("All requests completed.")
}
