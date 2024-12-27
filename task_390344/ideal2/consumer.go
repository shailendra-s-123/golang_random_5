

package main

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

func main() {
	kafkaBroker := "localhost:9092"
	kafkaTopic := "tcp_metrics"
	groupID := "tcp_metrics_consumer_group"

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		GroupID:  groupID,
		Topic:    kafkaTopic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	defer reader.Close()
	log.Printf("Kafka consumer started for topic %s", kafkaTopic)

	for {
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		log.Printf("Received metric: %s", string(message.Value))
	}
}