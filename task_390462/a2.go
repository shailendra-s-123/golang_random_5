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
	conn      *websocket.Conn
	state     map[string]interface{}
	lastMessage []byte
	reconnect int
}

func (c *Client) handleConnection() {
	defer c.conn.Close()

	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			// Attempt to reconnect on failure.
			c.reconnect++
			c.reconnectWithBackoff()
			return
		}

		c.lastMessage = message
		log.Printf("Received: %s\n", message)

		// Handle state updates here.
		// For simplicity, let's log the received state.
		if len(message) > 0 {
			c.updateState(string(message))
		}

		// For demonstration, echo back the message.
		if err = c.conn.WriteMessage(messageType, message); err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

func (c *Client) reconnectWithBackoff() {
	time.Sleep(c.backoff())
	log.Println("Attempting to reconnect...")
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Println("Reconnect failed:", err)
		return
	}
	c.conn = conn
	log.Println("Reconnected successfully.")
	go c.handleConnection()
}

func (c *Client) backoff() time.Duration {
	if c.reconnect == 0 {
		return 100 * time.Millisecond // Fast initial reconnect.
	}
	base := 250 * time.Millisecond
	factor := 1.5
	max := 2 * time.Second
	return time.Duration(min(int64(base*time.Millisecond*float64(math.Pow(factor, float64(c.reconnect)))), max.Nanoseconds()))
}

func (c *Client) updateState(data string) {
	if c.state == nil {
		c.state = make(map[string]interface{})
	}
	// For simplicity, assume data is JSON-encoded key-value pairs.
	// Replace this with a proper JSON parsing/unmarshalling logic.
	for _, pair := range strings.Split(data, ";") {
		key, value := strings.Split(pair, "=")
		c.state[key] = value
	}
	log.Printf("State updated: %+v\n", c.state)
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