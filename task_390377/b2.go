package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Simulated order data
var (
	ordersServiceAData = make(map[string]Order)
	ordersServiceBData = make(map[string]Order)
	mu                  sync.Mutex // Mutex for strong consistency
)

// Order struct
type Order struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Updated time.Time `json:"updated"`
}

// Simulated Message Broker
type MessageBroker struct {
	queues map[string]chan Order
}

func NewMessageBroker() *MessageBroker {
	return &MessageBroker{
		queues: make(map[string]chan Order),
	}
}

func (mb *MessageBroker) CreateQueue(queueName string) {
	mb.queues[queueName] = make(chan Order, 100)
}

func (mb *MessageBroker) Publish(queueName string, order Order) {
	if queue, exists := mb.queues[queueName]; exists {
		queue <- order
	} else {
		log.Printf("Queue %s does not exist\n", queueName)
	}
}

func (mb *MessageBroker) Subscribe(queueName string) <-chan Order {
	if queue, exists := mb.queues[queueName]; exists {
		return queue
	}
	log.Printf("Queue %s does not exist\n", queueName)
	return nil
}

// Eventual Consistency handler
func eventualConsistencyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Service A updates its data
	ordersServiceAData[order.ID] = order
	fmt.Fprintf(w, "Order %s updated to '%s' in Service A\n", order.ID, order.Status)

	// Publish update to Service B queue
	mb.Publish("ServiceBQueue", order)
	fmt.Fprintf(w, "Published update to ServiceBQueue\n")
}

// Strong Consistency handler
func strongConsistencyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Begin distributed transaction
	fmt.Fprintf(w, "Transaction: Starting to update order %s to '%s'\n", order.ID, order.Status)
	mu.Lock()
	defer mu.Unlock()

	// Update both services atomically
	ordersServiceAData[order.ID] = order
	ordersServiceBData[order.ID] = order
	fmt.Fprintf(w, "Transaction: Updated order %s to '%s' in both services\n", order.ID, order.Status)

	fmt.Fprintf(w, "Transaction: Committed\n")
}

// Healthcheck handler
func healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "API is healthy\n")
}

// Main function
func main() {
	// Initialize message broker and queues
	mb := NewMessageBroker()
	mb.CreateQueue("ServiceBQueue")

	// Start Service B consumer in the background
	go func() {
		for order := range mb.Subscribe("ServiceBQueue") {
			ordersServiceBData[order.ID] = order
			fmt.Printf("Service B: Synced order %s to '%s'\n", order.ID, order.Status)
		}
	}()

	http.HandleFunc("/api/orders/eventual", eventualConsistencyHandler)
	http.HandleFunc("/api/orders/strong", strongConsistencyHandler)
	http.HandleFunc("/healthcheck", healthcheckHandler)

	fmt.Println("Order API listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}