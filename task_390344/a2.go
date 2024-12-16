package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "sync"
    "time"

    "github.com/segmentio/kafka-go"
)

type Metrics struct {
    PacketLoss float64
    RTT        time.Duration
    Throughput float64
}

type TCPMetricsService struct {
    kafkaWriter   *kafka.Writer
    sendInterval  time.Duration
    connectionMap sync.Map // Map to hold connections and their metrics
}

func (service *TCPMetricsService) calculateMetrics(connectionID string, totalPackets, lostPackets int, totalRTT time.Duration) Metrics {
    packetLoss := float64(lostPackets) / float64(totalPackets) * 100 // percentage
    avgRTT := totalRTT / time.Duration(totalPackets)
    throughput := float64(totalPackets*1500) / time.Since(time.Now()).Seconds() // assuming MTU of 1500 bytes

    return Metrics{
        PacketLoss: packetLoss,
        RTT:        avgRTT,
        Throughput: throughput,
    }
}

func (service *TCPMetricsService) publishMetrics(metrics Metrics) {
    msg := kafka.Message{
        Key:   []byte("metrics"),
        Value: []byte(fmt.Sprintf("PacketLoss: %.2f%%, RTT: %v, Throughput: %.2f", metrics.PacketLoss, metrics.RTT, metrics.Throughput)),
    }

    if err := service.kafkaWriter.WriteMessages(context.Background(), msg); err != nil {
        log.Printf("failed to publish metrics: %v", err)
    }
}

func (service *TCPMetricsService) handleConnection(conn net.Conn) {
    defer conn.Close()
    log.Printf("New connection from %s", conn.RemoteAddr())

    var totalPackets, lostPackets int
    var totalRTT time.Duration

    defer func() {
        metrics := service.calculateMetrics(conn.RemoteAddr().String(), totalPackets, lostPackets, totalRTT)
        service.publishMetrics(metrics)
    }()

    buffer := make([]byte, 1500) // Buffer for packets

    for {
        startTime := time.Now()
        n, err := conn.Read(buffer)
        if err != nil {
            log.Printf("Connection error: %v", err)
            return
        }
        totalPackets++

        // Simulate processing and packet loss
        if n < 0 { // Simulate packet loss
            lostPackets++
            continue
        }

        // Calculate RTT
        rtt := time.Since(startTime)
        totalRTT += rtt

        // Here, you could also send a response back to the client if needed
        // conn.Write([]byte("ACK"))
    }
}

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

        go service.handleConnection(conn) // Handle each connection concurrently
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
        sendInterval: 5 * time.Second,
    }

    tcpMetricsService.Start(":8080")
}