package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Simulated message broker interface
type MessageBroker interface {
	Publish(topic string, message []byte) error
}

// In-memory message broker for demonstration purposes
type inMemoryBroker struct {
	mu sync.Mutex
	channels map[string][]*chan []byte
}

func NewInMemoryBroker() *inMemoryBroker {
	return &inMemoryBroker{channels: make(map[string][]*chan []byte)}
}

func (b *inMemoryBroker) Publish(topic string, message []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.channels[topic] == nil {
		b.channels[topic] = make([]*chan []byte, 0)
	}

	newChan := make(chan []byte)
	b.channels[topic] = append(b.channels[topic], &newChan)

	go func() {
		defer close(*newChan)
		*newChan <- message
	}()

	return nil
}

func (b *inMemoryBroker) Subscribe(topic string, c chan []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.channels[topic] == nil {
		b.channels[topic] = make([]*chan []byte, 0)
	}

	b.channels[topic] = append(b.channels[topic], &c)
}

// Order struct for demonstration
type Order struct {
	ID     int    `json:"id"`
	Amount float64 `json:"amount"`
}

// Inventory struct for demonstration
type Inventory struct {
	Stock int `json:"stock"`
}

var inventory Inventory
var orderServiceChan chan []byte

func main() {
	broker := NewInMemoryBroker()

	// Simulate initial inventory
	inventory.Stock = 100

	// Set up order service to listen for events
	go orderService()

	// Set up inventory service
	http.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(inventory)
		} else if r.Method == http.MethodPost {
			var order Order
			if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			// Simulate strong consistency with a distributed transaction
			if !processOrderStrongConsistency(order) {
				http.Error(w, "Insufficient stock", http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(order)
		}
	})

	fmt.Println("Inventory service listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func orderService() {
	orderTopic := "order-created"
	orderServiceChan = make(chan []byte)

	broker := NewInMemoryBroker()
	broker.Subscribe(orderTopic, orderServiceChan)

	for message := range orderServiceChan {
		var order Order
		if err := json.Unmarshal(message, &order); err != nil {
			log.Println("Error processing order message:", err)
			continue
		}

		log.Printf("Order received: %+v\n", order)

		// Simulate eventual consistency with a message broker
		if !processOrderEventualConsistency(order) {
			log.Println("Failed to update inventory due to eventual consistency delay")
		} else {
			log.Println("Inventory updated successfully")
		}
	}
}

func processOrderStrongConsistency(order Order) bool {
	// Simulate locking and checking stock
	log.Println("Processing order with strong consistency...")
	inventory.Stock -= 1
	time.Sleep(100 * time.Millisecond) // Simulate transaction time

	if inventory.Stock < 0 {
		inventory.Stock += 1 // Rollback
		return false
	}

	return true
}

func processOrderEventualConsistency(order Order) bool {
	log.Println("Processing order with eventual consistency...")
	inventory.Stock -= 1
	time.Sleep(500 * time.Millisecond) // Simulate delay in updating

	return true
}