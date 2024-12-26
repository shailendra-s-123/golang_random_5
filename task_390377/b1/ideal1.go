

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Simulated message broker
type MessageBroker struct {
	channel chan string
}

func NewMessageBroker() *MessageBroker {
	return &MessageBroker{
		channel: make(chan string, 100),
	}
}

func (b *MessageBroker) Publish(message string) {
	b.channel <- message
	log.Printf("Message published: %s", message)
}

func (b *MessageBroker) Subscribe() <-chan string {
	return b.channel
}

// Simulated databases
var dbA map[string]string
var dbB map[string]string
var mu sync.Mutex

// Global message broker instance
var broker *MessageBroker

func init() {
	// Initialize databases
	dbA = make(map[string]string)
	dbB = make(map[string]string)
	// Initialize message broker
	broker = NewMessageBroker()
}

// Service A - Strong Consistency (synchronous update)
func serviceA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	}

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Ensure strong consistency for Service A
	mu.Lock()
	dbA[data.ID] = data.Value
	mu.Unlock()

	// Publish the event to the message broker for eventual consistency in Service B
	broker.Publish(fmt.Sprintf("%s:%s", data.ID, data.Value))

	// Respond to the client
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Data saved in Service A and event published"))
}

// Service B - Eventual Consistency (asynchronous update)
func serviceB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Read and return the data from Service B's database (which will be eventually consistent)
	mu.Lock()
	defer mu.Unlock()

	var response string
	for id, value := range dbB {
		response += fmt.Sprintf("%s: %s\n", id, value)
	}

	w.Write([]byte(response))
}

// Event Handler for Service B to update its state based on events from Service A
func handleEvents() {
	for message := range broker.Subscribe() {
		// Simulate processing delay
		time.Sleep(1 * time.Second)

		// Parse the event message
		var id, value string
		_, err := fmt.Sscanf(message, "%s:%s", &id, &value)
		if err != nil {
			log.Printf("Error processing event: %v", err)
			continue
		}

		// Update Service B's state based on the event from Service A
		mu.Lock()
		dbB[id] = value
		mu.Unlock()

		log.Printf("Service B updated: %s -> %s", id, value)
	}
}

// Health check to monitor API status
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service is running"))
}

func main() {
	// Start the event handler in a separate goroutine
	go handleEvents()

	// Define routes
	http.HandleFunc("/serviceA", serviceA) // POST to serviceA to simulate strong consistency
	http.HandleFunc("/serviceB", serviceB) // GET to serviceB to check eventual consistency
	http.HandleFunc("/health", healthCheck) // Health check endpoint

	// Start the HTTP server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

