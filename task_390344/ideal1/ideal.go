


   

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// Define the structure for metrics
type Metrics struct {
	PacketLoss float64       // Percentage of lost packets
	RTT        time.Duration // Round-trip time (in milliseconds)
	Throughput float64       // Throughput in bytes per second
}

// TCP Metrics Service to handle connections and publish metrics
type TCPMetricsService struct {
	kafkaWriter  *kafka.Writer
	sendInterval time.Duration
	mu           sync.Mutex // Protects the metrics state
	currentMetrics Metrics  // Stores the current metrics
}

// Function to simulate metric calculation
func (service *TCPMetricsService) calculateMetrics() Metrics {
	// In a real scenario, implement logic to calculate packet loss, RTT, and throughput.
	// For this example, we will use dummy values.
	return Metrics{
		PacketLoss: 0.01,                  // 1% packet loss (for demonstration)
		RTT:        45 * time.Millisecond, // Average RTT of 45ms
		Throughput: 1000.0,                // 1000 bytes per second throughput
	}
}

// Publish metrics to Kafka
func (service *TCPMetricsService) publishMetrics(metrics Metrics) {
	msg := kafka.Message{
		Key:   []byte("metrics"),
		Value: []byte(fmt.Sprintf("PacketLoss: %.2f, RTT: %v, Throughput: %.2f", metrics.PacketLoss, metrics.RTT, metrics.Throughput)),
	}

	// Attempt to send metrics to Kafka
	if err := service.kafkaWriter.WriteMessages(context.Background(), msg); err != nil {
		log.Printf("failed to publish metrics to Kafka: %v", err)
	} else {
		log.Printf("Published metrics: %v", metrics)
	}
}

// Start the TCP server and accept connections
func (service *TCPMetricsService) StartTCPServer(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("could not start TCP listener: %v", err)
	}
	defer listener.Close()

	log.Printf("TCP server listening on %s", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("could not accept connection: %v", err)
			continue
		}

		go service.handleConnection(conn)
	}
}

// Handle incoming connection and send metrics periodically
func (service *TCPMetricsService) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("new connection from %s", conn.RemoteAddr())

	// Send metrics every interval
	ticker := time.NewTicker(service.sendInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := service.calculateMetrics()
			service.mu.Lock()
			service.currentMetrics = metrics
			service.mu.Unlock()
			service.publishMetrics(metrics)
		}
	}
}

// Serve metrics via HTTP API
func (service *TCPMetricsService) serveMetrics(w http.ResponseWriter, r *http.Request) {
	// Get the current metrics and send them as JSON
	service.mu.Lock()
	defer service.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(service.currentMetrics); err != nil {
		http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
	}
}

func main() {
	// Set up Kafka writer
	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:9092"}, // Update to your Kafka broker address
		Topic:    "metrics_topic",            // The Kafka topic to publish metrics to
		Balancer: &kafka.LeastBytes{},        // Simple load balancing strategy
	})
	defer kafkaWriter.Close()

	// Initialize TCP metrics service
	tcpMetricsService := &TCPMetricsService{
		kafkaWriter:  kafkaWriter,
		sendInterval: 5 * time.Second, // Send metrics every 5 seconds
	}

	// Start the TCP server on port 8080
	go tcpMetricsService.StartTCPServer(":8080")

	// HTTP server to serve metrics at /metrics
	http.HandleFunc("/metrics", tcpMetricsService.serveMetrics)
	go func() {
		log.Println("HTTP server listening on :8081")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Block main goroutine to keep both servers running
	select {}
}
