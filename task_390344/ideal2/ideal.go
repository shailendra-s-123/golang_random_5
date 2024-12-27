//main.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
)

type Metric struct {
	ConnectionID string  `json:"connection_id"`
	Timestamp    string  `json:"timestamp"`
	BytesRead    int64   `json:"bytes_read"`
	Latency      float64 `json:"latency_ms"`
}

var (
	activeConnections int64
	metrics           []Metric // Stores metrics for display in the web server
)

func main() {
	// Define your server and Kafka details
	address := "0.0.0.0:9000"
	kafkaBroker := "localhost:9092"
	kafkaTopic := "tcp_metrics"
	webPort := "8080" // Web server will run on port 8080

	// Start the HTTP web server for metrics
	go startWebServer(webPort)

	// Start the TCP server for metrics ingestion
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error starting TCP server: %v", err)
	}
	defer listener.Close()

	log.Printf("TCP server started at %s", address)

	// Accept connections and handle them
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		atomic.AddInt64(&activeConnections, 1)
		go handleConnection(conn, kafkaBroker, kafkaTopic)
	}
}

// Handle each incoming connection and track metrics
func handleConnection(conn net.Conn, kafkaBroker, kafkaTopic string) {
	defer conn.Close()
	defer atomic.AddInt64(&activeConnections, -1)

	connectionID := fmt.Sprintf("%d", time.Now().UnixNano())
	startTime := time.Now()

	reader := bufio.NewReader(conn)
	totalBytesRead := int64(0)

	// Track data read from the connection
	for {
		data := make([]byte, 1024)
		bytesRead, err := reader.Read(data)
		if err != nil {
			log.Printf("Connection %s closed: %v", connectionID, err)
			break
		}
		totalBytesRead += int64(bytesRead)

		latency := time.Since(startTime).Seconds() * 1000
		metric := Metric{
			ConnectionID: connectionID,
			Timestamp:    time.Now().Format(time.RFC3339),
			BytesRead:    totalBytesRead,
			Latency:      latency,
		}

		// Send metrics to Kafka and store locally for web access
		sendToKafka(kafkaBroker, kafkaTopic, metric)
		metrics = append(metrics, metric)
	}
}

// Send metrics to Kafka for later processing
func sendToKafka(kafkaBroker, topic string, metric Metric) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaBroker},
		Topic:   topic,
	})
	defer writer.Close()

	data, _ := json.Marshal(metric)
	err := writer.WriteMessages(nil, kafka.Message{
		Value: data,
	})
	if err != nil {
		log.Printf("Failed to write metric to Kafka: %v", err)
	}
}

// Start a simple HTTP server that serves the metrics page
func startWebServer(port string) {
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// Serve a simple metrics page
		w.Header().Set("Content-Type", "application/json")
		metricsJSON, _ := json.Marshal(metrics)
		w.Write(metricsJSON)
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		// Serve basic server status
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(fmt.Sprintf("Active Connections: %d\n", atomic.LoadInt64(&activeConnections))))
	})

	log.Printf("Web server started at http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}