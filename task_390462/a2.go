package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketClient struct {
	url        *url.URL
	conn       *websocket.Conn
	reconnectCtx context.Context
	reconnectFunc context.CancelFunc
	state        map[string]interface{}
	cfg         Config
}

type Config struct {
	MaxReconnectAttempts int
	InitialBackoff       time.Duration
	BackoffFactor        float64
	DisconnectTimeout   time.Duration
}

func NewWebSocketClient(url *url.URL, cfg Config) *WebSocketClient {
	return &WebSocketClient{
		url:          url,
		cfg:          cfg,
		state:        make(map[string]interface{}),
		reconnectCtx: context.Background(),
	}
}

func (wsc *WebSocketClient) Connect() error {
	ctx, cancel := context.WithCancel(wsc.reconnectCtx)
	wsc.reconnectFunc = cancel
	for attempt := 1; attempt <= wsc.cfg.MaxReconnectAttempts; attempt++ {
		log.Printf("Attempt %d to connect to WebSocket server at %s\n", attempt, wsc.url)
		dialer := websocket.DefaultDialer{
			ReadTimeout:  wsc.cfg.DisconnectTimeout,
			WriteTimeout: wsc.cfg.DisconnectTimeout,
		}
		conn, _, err := dialer.DialContext(ctx, wsc.url.String(), nil)
		if err != nil {
			log.Printf("Failed to connect: %v\n", err)
			time.Sleep(wsc.backoffDuration(attempt))
			continue
		}
		wsc.conn = conn
		go wsc.handleMessages(conn)
		go wsc.handleConnection(conn)
		return nil
	}
	return fmt.Errorf("exceeded maximum reconnection attempts")
}

func (wsc *WebSocketClient) handleConnection(conn *websocket.Conn) {
	defer conn.Close()
	for {
		select {
		case <-wsc.reconnectCtx.Done():
			log.Println("Reconnection context done, closing connection")
			return
		case <-time.After(wsc.cfg.DisconnectTimeout):
			wsc.disconnect(fmt.Errorf("connection timed out"))
		}
	}
}

func (wsc *WebSocketClient) handleMessages(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			wsc.disconnect(err)
			return
		}
		log.Printf("Received message: %s\n", string(message))
		// Process the message and update state as needed
		wsc.updateState(string(message))
	}
}

func (wsc *WebSocketClient) backoffDuration(attempt int) time.Duration {
	base := time.Duration(wsc.cfg.InitialBackoff)
	multiplier := math.Pow(float64(wsc.cfg.BackoffFactor), float64(attempt-1))
	return base * time.Duration(multiplier)
}

func (wsc *WebSocketClient) disconnect(err error) {
	log.Printf("Disconnected: %v\n", err)
	if wsc.conn != nil {
		wsc.conn.Close()
	}
	wsc.Connect()
}

func (wsc *WebSocketClient) updateState(message string) {
	// Example state update based on message content
	var key, value string
	fmt.Sscanf(message, "%s=%s", &key, &value)
	wsc.state[key] = value
	log.Printf("State updated: %s=%s\n", key, value)
}

func main() {
	wsUrl, _ := url.Parse("ws://localhost:8080/ws")
	wsc := NewWebSocketClient(wsUrl, Config{
		MaxReconnectAttempts: 10,
		InitialBackoff:       time.Second,
		BackoffFactor:        2.0,
		DisconnectTimeout:   time.Minute,
	})

	if err := wsc.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}

	select {
	case <-time.After(time.Hour): // Keep running for an hour for demonstration purposes
	}

	wsc.reconnectFunc() // Disconnect by cancelling the reconnection context
}