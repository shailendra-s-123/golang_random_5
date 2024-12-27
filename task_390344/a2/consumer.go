// kafka_consumer.go

package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type Metrics struct {
	ActiveConnections int     `json:"active_connections"`
	Throughput        float64 `json:"throughput"` // Bytes per second
	Latency           float64 `json:"latency"`    // Milliseconds
}

func main() {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"}, // Kafka broker address
		Topic:   "metrics_topic",             // The Kafka topic to consume from
		GroupID: "metrics_consumer_group",    // Consumer group ID
	})
	defer kafkaReader.Close()

	for {
		m, err := kafkaReader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error while reading message: %v", err)
			continue
		}

		var metrics Metrics
		if err := json.Unmarshal(m.Value, &metrics); err != nil {
			log.Printf("Error unmarshalling metrics: %v", err)
			continue
		}

		log.Printf("Metrics received: Active Connections: %d, Throughput: %.2f bytes/s, Latency: %.2f ms", 
			metrics.ActiveConnections, metrics.Throughput, metrics.Latency)
	}
}