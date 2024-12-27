package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

var (
	// Store valid username-password pairs (hardcoded for simplicity)
	validUsers = map[string]string{
		"alice": "password123",
		"bob":   "securepass",
	}
	// For holding connected clients
	clients = make(map[net.Conn]string)
	mu      sync.Mutex
)

func main() {
	address := "0.0.0.0:9000"
	webPort := "8080" // Web server will run on port 8080

	// Start the HTTP web server for metrics or chat
	go startWebServer(webPort)

	// Start the TCP server for chat connection
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	log.Printf("Server started on TCP %s", address)

	// Accept new client connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleClient(conn)
	}
}

// Handle client authentication and communication
func handleClient(conn net.Conn) {
	defer conn.Close()

	// Authenticate the client
	if !authenticateClient(conn) {
		conn.Write([]byte("Authentication failed. Closing connection.\n"))
		return
	}

	// Add the client to the active clients list
	mu.Lock()
	clients[conn] = "Authenticated"
	mu.Unlock()

	// Send a welcome message
	conn.Write([]byte("Welcome to the chat! Type 'exit' to leave.\n"))

	// Start the chat
	handleChat(conn)
}

func authenticateClient(conn net.Conn) bool {
	conn.Write([]byte("Enter username: "))
	username, _ := bufio.NewReader(conn).ReadString('\n')
	username = strings.TrimSpace(username)

	conn.Write([]byte("Enter password: "))
	password, _ := bufio.NewReader(conn).ReadString('\n')
	password = strings.TrimSpace(password)

	// Debugging: Log the entered credentials
	log.Printf("Received username: %s", username)
	log.Printf("Received password: %s", password)

	// Check if the username and password are correct
	if correctPassword, exists := validUsers[username]; exists {
		if correctPassword == password {
			conn.Write([]byte("Authentication successful!\n"))
			return true
		}
	}

	conn.Write([]byte("Invalid username or password.\n"))
	return false
}

// Handle the chat messages from a client
func handleChat(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)

		if message == "exit" {
			conn.Write([]byte("Goodbye!\n"))
			break
		}

		// Broadcast the message to all connected clients
		mu.Lock()
		for client := range clients {
			if client != conn {
				client.Write([]byte(message + "\n"))
			}
		}
		mu.Unlock()
	}
}

// Start a simple HTTP server that serves the metrics or chat page
func startWebServer(port string) {
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// Serve a simple metrics page
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"message\":\"This could be your metrics page!\"}"))
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		// Serve basic server status
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(fmt.Sprintf("Active Connections: %d\n", len(clients))))
	})

	// Print the full URL that can be accessed via the browser
	log.Printf("Web server started at http://localhost:%s", port)

	// Start the HTTP server
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}