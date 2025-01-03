package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Order represents an order structure
type Order struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// Simulated Databases for two microservices: Order Service and Notification Service
var (
	orderData        = make(map[string]Order)
	notificationData = make(map[string]string)
	mu               sync.Mutex // Mutex for strong consistency
)

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

// Eventual Consistency Handler
func handleEventualConsistency(order Order, mb *MessageBroker) {
	// Publish order to Notification Service asynchronously
	mb.Publish("NotificationQueue", order)
	fmt.Printf("Order %s published to Notification Service for eventual consistency.\n", order.ID)
}

// Strong Consistency Handler
func handleStrongConsistency(order Order) {
	mu.Lock() // Start critical section
	defer mu.Unlock()

	// Update order data in a consistent manner
	orderData[order.ID] = order
	notificationData[order.ID] = order.Status // Simulate notification
	fmt.Printf("Order %s updated in both Order and Notification Services with strong consistency.\n", order.ID)
}

// CreateOrder creates a new order
func createOrder(w http.ResponseWriter, r *http.Request, mb *MessageBroker) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Use strong consistency for this example
	handleStrongConsistency(order)

	// Use eventual consistency to notify other services
	handleEventualConsistency(order, mb)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// Main function
func main() {
	// Initialize message broker and queues
	mb := NewMessageBroker()
	mb.CreateQueue("NotificationQueue")

	// Simulate notification service listening to the message queue
	go func() {
		for order := range mb.Subscribe("NotificationQueue") {
			time.Sleep(1 * time.Second) // Simulate processing delay
			fmt.Printf("Notification Service: Order %s with status '%s' has been processed.\n", order.ID, order.Status)
		}
	}()

	// Set up HTTP routes
	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			createOrder(w, r, mb)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start the HTTP server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}