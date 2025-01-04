package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn     *websocket.Conn
	lastMessage []byte
}

func (c *Client) handleConnection() {
	defer c.conn.Close()

	// Attempt to read messages in a loop.
	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			// Attempt to reconnect on failure.
			c.reconnect()
			return
		}
		c.lastMessage = message
		log.Printf("Received: %s\n", message)

		if err = c.conn.WriteMessage(messageType, message); err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

func (c *Client) reconnect() {
	for {
		log.Println("Attempting to reconnect...")
		conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
		if err != nil {
			log.Println("Reconnect failed:", err)
			time.Sleep(3 * time.Second) // Retry after a short delay.
			continue
		}
		c.conn = conn
		log.Println("Reconnected successfully.")
		go c.handleConnection()
		return
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := &Client{
		conn: conn,
	}

	// Handle the connection
	go client.handleConnection()
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("WebSocket server started on :8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
}