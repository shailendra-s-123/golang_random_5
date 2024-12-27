package main

import (
    "context"
    "log"
    "net"
    "sync"
    "time"

    "github.com/segmentio/kafka-go"
)

// Define a structure for the metrics
type Metrics struct {
    PacketLoss float64
    RTT        time.Duration
    Throughput float64
}

// A TCP metrics ingestion service
type TCPMetricsService struct {
    kafkaWriter *kafka.Writer
    conn        net.Conn
    sendInterval time.Duration
}

// Function to calculate metrics
func (service *TCPMetricsService) calculateMetrics() Metrics {
    // You will implement your calculations here
    // Initialize with dummy values for illustration
    return Metrics{
        PacketLoss: 0.02,     // Dummy value: 2% packet loss
        RTT:        30 * time.Millisecond, // Dummy value: average RTT
        Throughput: 1500.0,   // Dummy value: 1500 bytes/sec (placeholder)
    }
}

// Publish metrics to Kafka
func (service *TCPMetricsService) publishMetrics(metrics Metrics) {
    msg := kafka.Message{
        Key:   []byte("metrics"),
        Value: []byte(metricsToString(metrics)), // Convert metrics to string format
    }

    err := service.kafkaWriter.WriteMessages(context.Background(), msg)
    if err != nil {
        log.Printf("failed to publish metrics: %v", err)
    }
}

// Convert metrics struct to a string representation
func metricsToString(metrics Metrics) string {
    return fmt.Sprintf("PacketLoss: %.2f, RTT: %v, Throughput: %.2f", metrics.PacketLoss, metrics.RTT, metrics.Throughput)
}

// Start the TCP metrics ingestion service
func (service *TCPMetricsService) Start(address string) {
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

func (service *TCPMetricsService) handleConnection(conn net.Conn) {
    defer conn.Close()
    log.Printf("new connection from %s", conn.RemoteAddr())

    ticker := time.NewTicker(service.sendInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            metrics := service.calculateMetrics()
            service.publishMetrics(metrics)
        }
    }
}

func main() {
    kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
        Brokers: []string{"localhost:9092"},
        Topic:   "metrics_topic",
        Balancer: &kafka.LeastBytes{},
    })
    defer kafkaWriter.Close()

    tcpMetricsService := &TCPMetricsService{
        kafkaWriter:  kafkaWriter,
        sendInterval: 5 * time.Second, // Send every 5 seconds
    }

    tcpMetricsService.Start(":8080")
}