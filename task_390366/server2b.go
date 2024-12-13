// server.go
package main

import (
    "bufio"
    "fmt"
    "log"
    "net"
    "sync"
    "time"
)

const (
    messageBatchSize = 5 // Number of messages to batch before broadcasting
    batchInterval     = 50 * time.Millisecond // Interval to send batched messages
)

type Client struct {
    conn    net.Conn
    addr    string
    reader  *bufio.Reader
    writer  *bufio.Writer
    lastMsg time.Time
}

var (
    clients     map[string]*Client
    clientsMu   sync.RWMutex
    messageChan = make(chan string)
    batchWg     sync.WaitGroup
)

func startBatcher() {
    batchWg.Add(1)
    defer batchWg.Done()

    ticker := time.NewTicker(batchInterval)
    for {
        select {
        case msg := <-messageChan:
            clientsMu.RLock()
            defer clientsMu.RUnlock()
            for _, client := range clients {
                client.lastMsg = time.Now()
                if _, err := client.writer.WriteString(msg + "\n"); err != nil {
                    log.Printf("Error writing to client %s: %v", client.addr, err)
                    clientsMu.Lock()
                    delete(clients, client.addr)
                    clientsMu.Unlock()
                    client.conn.Close()
                }
            }
        case <-ticker.C:
            clientsMu.RLock()
            for addr, client := range clients {
                if time.Since(client.lastMsg) > 30*time.Second {
                    log.Printf("Disconnecting idle client %s", addr)
                    clientsMu.Lock()
                    delete(clients, addr)
                    clientsMu.Unlock()
                    client.conn.Close()
                }
            }
            clientsMu.RUnlock()
        case <-time.After(time.Minute):
            log.Println("Server status:", len(clients), "connected clients")
        }
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()

    reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)
    defer writer.Flush()

    addr := conn.RemoteAddr().String()
    clientsMu.Lock()
    clients[addr] = &Client{conn, addr, reader, writer, time.Now()}
    clientsMu.Unlock()

    log.Printf("New client connected: %s\n", addr)

    for {
        message, err := reader.ReadString('\n')
        if err != nil {
            log.Printf("Error reading from client %s: %v", addr, err)
            break
        }

        message = addr + ": " + message
        messageChan <- message
    }

    clientsMu.Lock()
    delete(clients, addr)
    clientsMu.Unlock()
    log.Printf("Client disconnected: %s\n", addr)
}

func main() {
    log.Println("Chat server started on :8080")

    listeners := make([]net.Listener, 5) // Simple load balancing via multiple listeners
    for i := 0; i < 5; i++ {
        listener, err := net.Listen("tcp", fmt.Sprintf(":%d", 8080+i))
        if err != nil {
            log.Fatal(err)
        }
        listeners[i] = listener
    }

    go startBatcher()

    for _, listener := range listeners {
        defer listener.Close()

        for {
            conn, err := listener.Accept()
            if err != nil {
                log.Printf("Error accepting connection: %v", err)
                continue
            }
            go handleConnection(conn)
        }
    }
}