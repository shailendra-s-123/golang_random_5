// server.go
package main

import (
    "bufio"
    "fmt"
    "net"
    "sync"
)

var (
    clients   = make(map[net.Conn]bool) // Map of connected clients
    clientsMu sync.Mutex                 // Mutex to protect access to clients map
)

// Broadcast message to all connected clients
func broadcastMessage(message string, sender net.Conn) {
    clientsMu.Lock()
    defer clientsMu.Unlock()
    for client := range clients {
        if client != sender { // Don't send the message back to the sender
            _, err := client.Write([]byte(message))
            if err != nil {
                fmt.Printf("Error sending message to client: %v\n", err)
                client.Close()
                delete(clients, client) // Remove disconnected client
            }
        }
    }
}

// Handle client connection
func handleConnection(conn net.Conn) {
    defer conn.Close()
    clientsMu.Lock()
    clients[conn] = true                       // Add new client
    clientsMu.Unlock()

    fmt.Printf("New client connected: %s\n", conn.RemoteAddr().String())

    scanner := bufio.NewScanner(conn)
    for scanner.Scan() { // Read messages from the client
        message := fmt.Sprintf("%s: %s\n", conn.RemoteAddr().String(), scanner.Text())
        fmt.Print(message) // Print to server console
        broadcastMessage(message, conn) // Broadcast received message
    }

    clientsMu.Lock()
    delete(clients, conn) // Remove the client when they disconnect
    clientsMu.Unlock()
    fmt.Printf("Client disconnected: %s\n", conn.RemoteAddr().String())
}

func main() {
    listener, err := net.Listen("tcp", ":8080") // Listen on port 8080
    if err != nil {
        fmt.Printf("Error starting server: %v\n", err)
        return
    }
    defer listener.Close()
    fmt.Println("Chat server started on :8080")

    for {
        conn, err := listener.Accept() // Accept new client connections
        if err != nil {
            fmt.Printf("Error accepting connection: %v\n", err)
            continue
        }
        go handleConnection(conn) // Handle client in a goroutine
    }
}