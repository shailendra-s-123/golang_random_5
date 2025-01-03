package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Simulated message broker for eventual consistency
var messageBroker = make(chan string)

// Simulated database for Service A and B
var dbA = make(map[string]string)
var dbB = make(map[string]string)

// Mutex for synchronizing access to databases
var muA sync.Mutex
var muB sync.Mutex

// Service A: Strong Consistency
func serviceA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data map[string]string
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Start a distributed transaction
	muA.Lock()
	defer muA.Unlock()

	dbA[data["id"]] = data["value"]

	// Strong consistency: immediately update Service B
	muB.Lock()
	defer muB.Unlock()
	dbB[data["id"]] = data["value"]

	fmt.Fprintf(w, "Service A: Data saved with strong consistency, id: %s", data["id"])
}

// Service B: Eventual Consistency using message broker
func serviceB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data map[string]string
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Simulated eventual consistency by sending a message to the broker
	message := fmt.Sprintf("Eventual consistency message for id: %s, value: %s", data["id"], data["value"])
	messageBroker <- message

	fmt.Fprintf(w, "Service B: Message sent to broker for eventual consistency, id: %s", data["id"])
}

// Simulate the message broker processing messages
func messageBrokerProcess() {
	for message := range messageBroker {
		time.Sleep(2 * time.Second) // Simulate delay in processing
		// Here we would update Service B's database
		parts := strings.Split(message, ", ")
		id := strings.Split(parts[0], ": ")[1]
		value := strings.Split(parts[1], ": ")[1]

		muB.Lock()
		dbB[id] = value
		muB.Unlock()

		log.Printf("Message processed: %s", message)
	}
}

func main() {
	// Start message broker processing in a separate goroutine
	go messageBrokerProcess()

	http.HandleFunc("/serviceA", serviceA)
	http.HandleFunc("/serviceB", serviceB)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}