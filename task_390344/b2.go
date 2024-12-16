package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Define a structure for the metrics
type Metrics struct {
	PacketLoss float64
	RTT        time.Duration
	Throughput float64
}

// Helper function to create a Kafka producer
func createKafkaProducer(brokers []string, topic string) (*kafka.Producer, error) {
	p := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"delivery.report.on.error": true,
		"delivery.report.on.success": true,
	})

	go func() {
		for err := range p.Events() {
			if err != nil {
				log.Printf("Kafka delivery error: %v\n", err)
			}
		}
	}()

	if err := p.Produce(&kafka.Message{Topic: topic, Key: []byte("metrics")}); err != nil {
		return nil, err
	}

	return p, nil
}

// Function to calculate metrics
func calculateMetrics(sent, received int, startTime time.Time) Metrics {
	rtt := time.Since(startTime)
	packetLoss := float64(sent-received) / float64(sent)
	throughput := float64(received) * 8.0 / rtt.Seconds() // in bits per second
	return Metrics{
		PacketLoss: packetLoss,
		RTT:        rtt,
		Throughput: throughput,
	}
}

// Start the TCP metrics ingestion service
func StartTCPMetricsService(address string, brokers []string, topic string) {
	log.Printf("TCP server listening on %s", address)

	producer, err := createKafkaProducer(brokers, topic)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Could not start TCP listener: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Could not accept connection: %v", err)
			continue
		}

		go handleConnection(conn, producer)
	}
}

func handleConnection(conn net.Conn, producer *kafka.Producer) {
	defer conn.Close()
	log.Printf("New connection from %s", conn.RemoteAddr())

	sendInterval := 5 * time.Second
	ticker := time.NewTicker(sendInterval)
	defer ticker.Stop()

	sent := 0
	received := 0
	startTime := time.Now()

	for {
		select {
		case <-ticker.C:
			metrics := calculateMetrics(sent, received, startTime)
			publishMetrics(producer, metrics)
			startTime = time.Now() // Reset start time for next interval
			sent = 0
			received = 0

		default:
			// Simulate sending a packet
			_, err := conn.Write([]byte("ping"))
			if err != nil {
				log.Printf("Failed to send packet: %v", err)
				return
			}
			sent++

			// Simulate receiving a packet
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil || n == 0 {
				log.Printf("Failed to read packet: %v", err)
				return
			}
			received++
		}
	}
}

func publishMetrics(producer *kafka.Producer, metrics Metrics) {
	msg := kafka.Message{
		Topic: "metrics_topic",
		Value: []byte(fmt.Sprintf("PacketLoss: %.2f, RTT: %v, Throughput: %.2f", metrics.PacketLoss, metrics.RTT, metrics.Throughput)),
	}

	err := producer.Produce(&msg)
	if err != nil {
		log.Printf("Failed to publish metrics: %v", err)
	}
}

func main() {
	brokers := []string{"localhost:9092"}
	StartTCPMetricsService(":8080", brokers, "metrics_topic")
}