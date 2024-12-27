package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Metrics struct {
	ConnectionCount int    `json:"connection_count"`
	Throughput      float64 `json:"throughput"` // in bytes per second
	Latency         float64 `json:"latency"`    // in milliseconds
}

type TCPMetricsServer struct {
	producer *kafka.Producer
	mu       sync.Mutex
	connections int
}

func NewTCPMetricsServer(brokers []string, topic string) (*TCPMetricsServer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		return nil, err
	}
	return &TCPMetricsServer{
		producer: p,
	}, nil
}

func (s *TCPMetricsServer) Start(address string) {
	log.Printf("TCP server listening on %s", address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("could not start TCP listener: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("could not accept connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *TCPMetricsServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	s.mu.Lock()
	s.connections++
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.connections--
		s.mu.Unlock()
	}()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("error reading from connection: %v", err)
			}
			break
		}
		// Simulate throughput calculation
		time.Sleep(10 * time.Millisecond)
	}

	s.publishMetrics()
}

func (s *TCPMetricsServer) publishMetrics() {
	s.mu.Lock()
	metrics := Metrics{
		ConnectionCount: s.connections,
		Throughput:     1000.0, // Example throughput
		Latency:         50.0,   // Example latency
	}
	s.mu.Unlock()

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: "metrics_topic", Partition: kafka.PartitionAny},
		Value:          []byte(json.Marshal(metrics)),
	}

	err := s.producer.Produce(&kafka.DeliveryReport{}, msg)
	if err != nil {
		log.Printf("failed to produce message: %v", err)
	}
}

func main() {
	brokers := []string{"localhost:9092"}
	server, err := NewTCPMetricsServer(brokers, "metrics_topic")
	if err != nil {
		log.Fatalf("failed to initialize producer: %v", err)
	}

	go server.Start(":8080")

	select {} // Block main thread
}