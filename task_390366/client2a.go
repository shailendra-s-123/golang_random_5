// client.go
package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // Connect to the chat server
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        fmt.Printf("Error connecting to server: %v\n", err)
        return
    }
    defer conn.Close()
    fmt.Println("Connected to chat server at localhost:8080")

    // Handle messages received from the server
    go receiveMessages(conn)

    // Capture signals to cleanly exit when the user wants to quit
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    // Read messages from stdin and send to the server
    inputScanner := bufio.NewScanner(os.Stdin)
    for {
        select {
        case <-quit:
            fmt.Println("Exiting chat client.")
            return
        default:
            if inputScanner.Scan() {
                message := inputScanner.Text()
                if message == "/quit" { // User can type /quit to exit
                    return
                }
                _, err := conn.Write([]byte(message + "\n")) // Send message with newline
                if err != nil {
                    fmt.Printf("Error sending message: %v\n", err)
                    return
                }
            }
        }
    }
}

// receiveMessages listens for incoming messages from the server
func receiveMessages(conn net.Conn) {
    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        fmt.Print(scanner.Text()) // Print messages from the server
    }
    
    // Check for scanner errors
    if err := scanner.Err(); err != nil {
        fmt.Printf("Error reading from server: %v\n", err)
    }
}