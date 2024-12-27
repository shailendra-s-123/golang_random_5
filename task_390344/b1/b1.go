package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "time"

    "github.com/confluentinc/confluent-kafka-go/kafka"
)

// Metric represents a TCP metric.
type Metric struct {
    Timestamp time.Time `json:"timestamp"`
    PacketLoss float32   `json:"packet_loss"`
    RTT        float32   `json:"rtt"` // Round-Trip Time in milliseconds
    Throughput float32   `json:"throughput"` // In bytes per second
}

// ProducerConfig holds Kafka producer configuration.
var ProducerConfig *kafka.ConfigMap = &kafka.ConfigMap{
    "bootstrap.servers": "localhost:9092", // Update with your Kafka broker addresses
    "topic.metrics.ingestion": "tcp_metrics", // Change to your desired topic name
}

func main() {
    // Start Kafka producer
    producer, err := kafka.NewProducer(ProducerConfig)
    if err != nil {
        log.Fatalf("Failed to create Kafka producer: %v", err)
    }
    defer producer.Close()

    // Start TCP server
    ln, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    defer ln.Close()

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Printf("Failed to accept connection: %v", err)
            continue
        }
        defer conn.Close()

        scanner := bufio.NewScanner(conn)
        for scanner.Scan() {
            var metric Metric
            if err := json.Unmarshal([]byte(scanner.Text()), &metric); err != nil {
                log.Printf("Failed to parse metric: %v", err)
                continue
            }

            fmt.Printf("Received metric: %+v\n", metric)

            // Produce metric to Kafka
            msg := &kafka.Message{
                Value: []byte(scanner.Text()),
            }

            if err := producer.Produce(&kafka.DeliveryReport{Topic: "tcp_metrics", Partition: kafka.PartitionAny}, msg); err != nil {
                log.Printf("Failed to produce message to Kafka: %v", err)
            }
        }

        if err := scanner.Err(); err != nil {
            log.Printf("Scanner error: %v", err)
        }
    }
}