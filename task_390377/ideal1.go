package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// Simulated Databases for Microservices
var (
	serviceAData = make(map[string]string)
	serviceBData = make(map[string]string)
	mu           sync.Mutex // Mutex for strong consistency
)

// Simulated Message Broker
type MessageBroker struct {
	queues map[string]chan string
}

func NewMessageBroker() *MessageBroker {
	return &MessageBroker{
		queues: make(map[string]chan string),
	}
}

func (mb *MessageBroker) CreateQueue(queueName string) {
	mb.queues[queueName] = make(chan string, 100)
}

func (mb *MessageBroker) Publish(queueName, message string) {
	if queue, exists := mb.queues[queueName]; exists {
		queue <- message
	} else {
		log.Printf("Queue %s does not exist\n", queueName)
	}
}

func (mb *MessageBroker) Subscribe(queueName string) <-chan string {
	if queue, exists := mb.queues[queueName]; exists {
		return queue
	}
	log.Printf("Queue %s does not exist\n", queueName)
	return nil
}

// Eventual Consistency Example
func eventualConsistency(mb *MessageBroker) {
	fmt.Println("Eventual Consistency: Starting...")

	// Service A updates its data
	mu.Lock()
	serviceAData["order-123"] = "Processed"
	mu.Unlock()
	fmt.Println("Service A: Updated order-123 to 'Processed'")

	// Service A sends an update to Service B via message broker
	mb.Publish("ServiceBQueue", "order-123:Processed")
	fmt.Println("Service A: Published update to ServiceBQueue")

	// Service B listens for updates
	go func() {
		for msg := range mb.Subscribe("ServiceBQueue") {
			parts := splitMessage(msg)
			mu.Lock()
			serviceBData[parts[0]] = parts[1]
			mu.Unlock()
			fmt.Printf("Service B: Synced %s to '%s'\n", parts[0], parts[1])
		}
	}()

	time.Sleep(2 * time.Second) // Simulate processing delay
	fmt.Println("Eventual Consistency: Finished\n")
}

// Strong Consistency Example
func strongConsistency() {
	fmt.Println("Strong Consistency: Starting...")

	// Begin distributed transaction
	fmt.Println("Transaction: Starting")
	mu.Lock()
	defer mu.Unlock()

	// Perform updates across services atomically
	serviceAData["order-456"] = "Completed"
	serviceBData["order-456"] = "Completed"
	fmt.Println("Transaction: Updated order-456 to 'Completed' in both services")

	// Commit transaction
	fmt.Println("Transaction: Committed")
	fmt.Println("Strong Consistency: Finished\n")
}

// Utility function to split simulated messages
func splitMessage(message string) [2]string {
	parts := strings.Split(message, ":")
	return [2]string{parts[0], parts[1]}
}

// Main function
func main() {
	// Initialize message broker and queues
	mb := NewMessageBroker()
	mb.CreateQueue("ServiceBQueue")

	// Demonstrate eventual consistency
	eventualConsistency(mb)

	// Demonstrate strong consistency
	strongConsistency()

	// Final state of data
	fmt.Println("Final State of Service A Data:", serviceAData)
	fmt.Println("Final State of Service B Data:", serviceBData)
}