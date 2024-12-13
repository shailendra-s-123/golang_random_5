package main

import (
	"fmt"
	"log"
	"net"
)

// Client represents a connected client
type Client struct {
	conn net.Conn
	name string
}

var clients []*Client

// handleClient manages communication with a single client
func handleClient(conn net.Conn) {
	defer conn.Close()

	// Read client's name
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		log.Println("Error reading client name:", err)
		return
	}
	name := string(buf)

	// Welcome the client
	fmt.Fprintf(conn, "Welcome, %s!\n", name)

	// Add client to the list
	clients = append(clients, &Client{conn: conn, name: name})
	defer func() {
		for i, c := range clients {
			if c.conn == conn {
				clients = append(clients[:i], clients[i+1:]...)
				break
			}
		}
	}()

	log.Printf("%s connected\n", name)

	// Listen for messages from the client
	for {
		_, err = conn.Read(buf)
		if err != nil {
			log.Println("Error reading from client:", err)
			return
		}
		message := string(buf)
		broadcast(name + ": " + message)
	}
}

// broadcast sends a message to all connected clients
func broadcast(message string) {
	for _, client := range clients {
		fmt.Fprintf(client.conn, "%s\n", message)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	log.Println("Server listening on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go handleClient(conn)
	}
}