// server.go
package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"
)

type Client struct {
    conn     net.Conn
    sendChan chan []byte // Channel for sending messages to this client
}

type ChatServer struct {
    clients     map[net.Conn]*Client // All connected clients
    clientsMu   sync.Mutex            // Mutex for clients map
    broadcastCh chan []byte           // Channel for broadcasting messages
    wg          sync.WaitGroup         // WaitGroup for graceful shutdown
}

// NewChatServer initializes a new chat server
func NewChatServer() *ChatServer {
    return &ChatServer{
        clients:     make(map[net.Conn]*Client),
        broadcastCh: make(chan []byte, 100), // Buffered channel for efficient broadcasting
    }
}

// Run starts the chat server
func (cs *ChatServer) Run(port string) {
    listener, err := net.Listen("tcp", port)
    if err != nil {
        fmt.Printf("Error starting server: %v\n", err)
        return
    }
    defer listener.Close()
    fmt.Println("Chat server started on", port)

    // Listen for incoming connections
    go cs.handleBroadcast()

    // Graceful shutdown
    go cs.gracefulShutdown()

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("Error accepting connection: %v\n", err)
            continue
        }
        cs.wg.Add(1)
        go cs.handleConnection(conn)
    }
}

// handleConnection manages a single client connection
func (cs *ChatServer) handleConnection(conn net.Conn) {
    defer func() {
        conn.Close()
        cs.wg.Done()
        cs.removeClient(conn)
    }()

    client := &Client{
        conn:     conn,
        sendChan: make(chan []byte, 10), // Buffered channel for sending messages
    }

    cs.clientsMu.Lock()
    cs.clients[conn] = client
    cs.clientsMu.Unlock()

    fmt.Printf("New client connected: %s\n", conn.RemoteAddr().String())

    go cs.handleMessages(client) // Start message handling for this client

    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        message := scanner.Bytes()
        cs.broadcastCh <- message // Send message to broadcast channel
    }

    // Check for scanner errors
    if err := scanner.Err(); err != nil {
        fmt.Printf("Scanner error for client %s: %v\n", conn.RemoteAddr(), err)
    }
}

// handleMessages sends messages from the client's send channel to the client
func (cs *ChatServer) handleMessages(client *Client) {
    for message := range client.sendChan {
        _, err := client.conn.Write(message)
        if err != nil {
            fmt.Printf("Error sending message to client %s: %v\n", client.conn.RemoteAddr(), err)
            break
        }
    }
}

// removeClient removes a client from the server
func (cs *ChatServer) removeClient(conn net.Conn) {
    cs.clientsMu.Lock()
    defer cs.clientsMu.Unlock()
    delete(cs.clients, conn)
    fmt.Printf("Client disconnected: %s\n", conn.RemoteAddr().String())
}

// handleBroadcast listens for messages to broadcast to clients
func (cs *ChatServer) handleBroadcast() {
    for message := range cs.broadcastCh {
        cs.clientsMu.Lock()
        for _, client := range cs.clients {
            select {
            case client.sendChan <- message: // Send message to client's send channel
            default:
                fmt.Println("Client's send channel is full; skipping...")
            }
        }
        cs.clientsMu.Unlock()
    }
}

// gracefulShutdown waits for shutdown signals and closes channels
func (cs *ChatServer) gracefulShutdown() {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    <-quit // Wait for shutdown signal
    fmt.Println("Shutting down server gracefully...")
    close(cs.broadcastCh) // Close the broadcast channel to stop broadcasting
    cs.wg.Wait()          // Wait for all clients to finish processing
    fmt.Println("Server shutdown complete.")
}

func main() {
    server := NewChatServer()
    server.Run(":8080")
}