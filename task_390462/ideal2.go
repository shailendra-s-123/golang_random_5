package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Adjust for production.
	},
}

type Client struct {
	conn         *websocket.Conn
	mutex        sync.Mutex
	messageQueue chan []byte
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

func newClient(conn *websocket.Conn) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	client := &Client{
		conn:         conn,
		messageQueue: make(chan []byte, 100),
		ctx:          ctx,
		cancel:       cancel,
	}
	client.wg.Add(1)
	return client
}

func (c *Client) handleConnection() {
	defer func() {
		c.cleanup()
		c.wg.Done()
	}()

	for {
		select {
		case msg := <-c.messageQueue:
			c.mutex.Lock()
			err := c.conn.WriteMessage(websocket.TextMessage, msg)
			c.mutex.Unlock()
			if err != nil {
				log.Println("Write error:", err)
				c.reconnect()
				return
			}
		case <-c.ctx.Done():
			log.Println("Client context cancelled")
			return
		}
	}
}

func (c *Client) reconnect() {
	for {
		select {
		case <-c.ctx.Done():
			log.Println("Reconnection aborted: context cancelled")
			return
		default:
			log.Println("Attempting to reconnect...")
			conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
			if err != nil {
				log.Println("Reconnect failed:", err)
				time.Sleep(2 * time.Second)
				continue
			}

			c.mutex.Lock()
			c.conn = conn
			c.mutex.Unlock()

			log.Println("Reconnected successfully.")
			go c.handleConnection()
			return
		}
	}
}

func (c *Client) SendMessage(msg []byte) {
	select {
	case c.messageQueue <- msg:
	default:
		log.Println("Message queue full, dropping message")
	}
}

func (c *Client) cleanup() {
	c.cancel()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.conn != nil {
		c.conn.Close()
	}
	close(c.messageQueue)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := newClient(conn)
	log.Println("New client connected")
	go client.handleConnection()

	defer func() {
		client.cleanup()
		log.Println("Client disconnected")
	}()
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("WebSocket server started on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}