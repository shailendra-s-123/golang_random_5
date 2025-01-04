package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketClient represents a WebSocket client
type WebSocketClient struct {
	url           *url.URL
	conn          *websocket.Conn
	connChan      chan *websocket.Conn
	disconnectChan chan error
	reconnectCtx   context.Context
	reconnectFunc context.CancelFunc
}

func NewWebSocketClient(url *url.URL) *WebSocketClient {
	return &WebSocketClient{
		url:           url,
		connChan:      make(chan *websocket.Conn),
		disconnectChan: make(chan error),
	}
}

func (wsc *WebSocketClient) Connect(ctx context.Context) error {
	// Connect to the WebSocket server
	log.Printf("Connecting to WebSocket server at %s\n", wsc.url)
	dialer := websocket.DefaultDialer{
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	conn, _, err := dialer.DialContext(ctx, wsc.url.String(), nil)
	if err != nil {
		log.Printf("Failed to connect: %v\n", err)
		return err
	}
	wsc.conn = conn
	wsc.connChan <- conn
	go wsc.handleConnection(conn)
	go wsc.handleMessages(conn)
	return nil
}

func (wsc *WebSocketClient) handleConnection(conn *websocket.Conn) {
	defer conn.Close()
	for {
		select {
		case <-wsc.reconnectCtx.Done():
			log.Println("Reconnection context done, closing connection")
			return
		case err := <-wsc.disconnectChan:
			log.Printf("Disconnected: %v\n", err)
			wsc.reconnect(wsc.reconnectCtx)
		}
	}
}

func (wsc *WebSocketClient) handleMessages(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			wsc.disconnectChan <- err
			return
		}
		log.Printf("Received message: %s\n", string(message))
	}
}

func (wsc *WebSocketClient) reconnect(ctx context.Context) {
	log.Println("Attempting to reconnect")
	backoff := time.Millisecond * 100
	for {
		select {
		case <-ctx.Done():
			log.Println("Reconnection context done, stopping")
			return
		default:
			log.Printf("Reconnecting in %s\n", backoff)
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
			if err := wsc.Connect(ctx); err != nil {
				log.Printf("Reconnect failed: %v\n", err)
			} else {
				log.Println("Reconnected successfully")
				return
			}
		}
	}
}

func (wsc *WebSocketClient) Disconnect() {
	close(wsc.disconnectChan)
	wsc.reconnectFunc()
	log.Println("Disconnected")
}

func main() {
	wsUrl, _ := url.Parse("ws://localhost:8080/ws")
	wsc := NewWebSocketClient(wsUrl)

	ctx, cancel := context.WithCancel(context.Background())
	wsc.reconnectCtx, wsc.reconnectFunc = context.WithCancel(ctx)

	defer cancel()

	if err := wsc.Connect(wsc.reconnectCtx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}

	// Simulate a disconnect after 5 seconds
	time.Sleep(time.Second * 5)
	wsc.Disconnect()

	// Wait for the connection to be closed
	select {
	case <-time.After(time.Second * 10):
		log.Println("Connection closed")
	}
}