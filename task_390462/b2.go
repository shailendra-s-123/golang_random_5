package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn         *websocket.Conn
	mutex        sync.Mutex
	lastMessage  []byte
	messageQueue chan []byte
	ctx          context.Context
	cancel       context.CancelFunc
}

func newClient() *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		messageQueue: make(chan []byte, 100),
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (c *Client) handleConnection() {
	defer c.conn.Close()

	for {
		select {
		case msg := <-c.messageQueue:
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Println("Write error:", err)
				// If write fails, attempt to reconnect.
				c.reconnect()
				return
			}
		case <-c.ctx.Done():
			log.Println("Disconnecting due to context cancellation")
			return
		}
	}
}

func (c *Client) reconnect() {
	for {
		log.Println("Attempting to reconnect...")
		conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
		if err != nil {
			log.Println("Reconnect failed:", err)
			time.Sleep(time.Second * 2) // Retry after a short delay.
			continue
		}

		// Clear old connection and set new one.
		c.mutex.Lock()
		c.conn = conn
		c.mutex.Unlock()

		log.Println("Reconnected successfully.")
		go c.handleConnection()
		
		// Send all the queued messages during reconnect.
		for msg := range c.messageQueue {
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Println("Write error during reconnect:", err)
				break
			}
		}
		return
	}
}

func (c *Client) SendMessage(msg []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn == nil {
		// If connection is nil, queue the message.
		c.messageQueue <- msg
		return
	}

	// If connection is established, send the message directly.
	if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Println("Write error:", err)
		// If write fails, attempt to reconnect.
		c.reconnect()
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := newClient()
	client.conn = conn
	go client.handleConnection()
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)