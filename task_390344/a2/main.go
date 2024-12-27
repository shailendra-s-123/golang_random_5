// main.go

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type Metrics struct {
	ActiveConnections int       `json:"active_connections"`
	Throughput        float64   `json:"throughput"` // Bytes per second
	Latency           float64   `json:"latency"`    // Milliseconds
}

type TCPMetricsService struct {
	kafkaWriter          *kafka.Writer
	activeConnections    int
	mutex                sync.Mutex
	dataReceived         int64
	startTime            time.Time
}

func (service *TCPMetricsService) handleConnection(conn net.Conn) {
	defer conn.Close()
	defer service.decrementConnections()

	service.incrementConnections()
	defer service.recordStartTime()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break // Connection closed
		}
		service.dataReceived += int64(n)
	}
}

func (service *TCPMetricsService) incrementConnections() {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.activeConnections++
}

func (service *TCPMetricsService) decrementConnections() {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.activeConnections--
}

func (service *TCPMetricsService) recordStartTime() {
	service.startTime = time.Now()
}

func (service *TCPMetricsService) calculateMetrics() Metrics {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	elapsed := time.Since(service.startTime).Seconds()
	throughput := float64(service.dataReceived) / elapsed

	return Metrics{
		ActiveConnections: service.activeConnections,
		Throughput:        throughput,
		Latency:           elapsed * 1000, // Convert to milliseconds
	}
}

func (service *TCPMetricsService) publishMetrics() {
	metrics := service.calculateMetrics()

	message, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("failed to marshal metrics: %v", err)
		return
	}

	msg := kafka.Message{
		Key:   []byte("metrics"),
		Value: message,
	}

	if err := service.kafkaWriter.WriteMessages(context.Background(), msg); err != nil {
		log.Printf("failed to publish metrics to Kafka: %v", err)
	}
}

func (service *TCPMetricsService) StartTCPServer(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("could not start TCP listener: %v", err)
	}
	defer listener.Close()

	log.Printf("TCP server listening on %s", address)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			service.publishMetrics()
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("could not accept connection: %v", err)
			continue
		}
		go service.handleConnection(conn)
	}
}

func main() {
	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:9092"}, // Update to your Kafka broker address
		Topic:    "metrics_topic",            // The Kafka topic to publish metrics to
		Balancer: &kafka.LeastBytes{},        // Simple load balancing strategy
	})
	defer kafkaWriter.Close()

	tcpMetricsService := &TCPMetricsService{
		kafkaWriter:   kafkaWriter,
		startTime:     time.Now(),
	}

	go tcpMetricsService.StartTCPServer(":8080")

	select {} // Keep the main goroutine alive
}