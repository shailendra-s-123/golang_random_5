package main

import (
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":   []string{"localhost:9092"},
		"group.id":            "metrics_consumer",
		"auto.offset.reset":   "earliest",
		"enable.auto.commit":  "true",
		"auto.commit.interval.ms": "1000",
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer c.Close()

	err = c.SubscribeTopics([]string{"metrics_topic"}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}

	for {
		msg, err := c.ReadMessage(-1)
		if err == kafka.ErrAllTopicsConsumed {
			log.Printf("All topics consumed")
			time.Sleep(1 * time.Second)
			continue
		} else if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		var metrics Metrics
		if err := json.Unmarshal(msg.Value, &metrics); err != nil {
			log.Printf("Error unmarshalling JSON: %v", err)
			continue
		}

		log.Printf("Received metrics: %+v", metrics)
	}
}