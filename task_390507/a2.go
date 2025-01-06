package main

import (
	"container/heap"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Request struct {
	ID       int
	Priority int
}

type PriorityQueue []*Request

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority > pq[j].Priority // Higher priority first
}
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Request)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

type RateLimiter struct {
	c         chan struct{}
	done      chan struct{}
	interval  time.Duration
	capacity  int
	queue     PriorityQueue
	mutex     sync.Mutex
	metrics   Metrics
}

type Metrics struct {
	TotalRequests int
	ProcessedRequests int
	Latency []time.Duration
	Throughput []time.Duration
}

func NewRateLimiter(capacity int, interval time.Duration) *RateLimiter {
	c := make(chan struct{}, capacity)
	rl := &RateLimiter{
		c:         c,
		done:      make(chan struct{}),
		interval:  interval,
		capacity:  capacity,
		queue:     make(PriorityQueue, 0),
	}

	go rl.leak()
	return rl
}

func (rl *RateLimiter) leak() {
	ticker := time.NewTicker(rl.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mutex.Lock()
			if len(rl.queue) > 0 {
				select {
				case rl.c <- struct{}{}: // Refill the bucket
					request := heap.Pop(&rl.queue).(*Request)
					go rl.processRequest(request)
				default:
					// Bucket is full; do nothing
				}
			}
			rl.mutex.Unlock()
		case <-rl.done:
			return
		}
	}
}

func (rl *RateLimiter) Allow(req *Request) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	rl.metrics.TotalRequests++
	heap.Push(&rl.queue, req) // Add request to the priority queue
}

func (rl *RateLimiter) processRequest(req *Request) {
	start := time.Now()
	// Simulate processing a task
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	latency := time.Since(start)

	// Update metrics
	rl.mutex.Lock()
	rl.metrics.ProcessedRequests++
	rl.metrics.Latency = append(rl.metrics.Latency, latency)
	throughput := time.Duration(len(rl.metrics.Latency)) * rl.interval
	rl.metrics.Throughput = append(rl.metrics.Throughput, throughput)
	rl.mutex.Unlock()

	fmt.Printf("Processed request %d with latency %v\n", req.ID, latency)
}

func (rl *RateLimiter) GetMetrics() Metrics {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	return rl.metrics
}

func (rl *RateLimiter) Shutdown() {
	close(rl.done)
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
		go func(id int) {
			defer wg.Done()
			priority := rand.Intn(10) // Random priority for each request
			req := &Request{ID: id, Priority: priority}
			rateLimiter.Allow(req)
		}(i)
	}

	wg.Wait()
	rateLimiter.Shutdown()

	// Display metrics
	metrics := rateLimiter.GetMetrics()
	fmt.Printf("Total Requests: %d\n", metrics.TotalRequests)
	fmt.Printf("Processed Requests: %d\n", metrics.ProcessedRequests)
	fmt.Printf("Latencies: %v\n", metrics.Latency)
	fmt.Printf("Throughputs: %v\n", metrics.Throughput)
}