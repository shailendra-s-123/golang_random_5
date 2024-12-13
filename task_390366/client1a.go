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
    // Connect to chat server
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        fmt.Printf("Error connecting to server: %v\n", err)
        return
    }
    defer conn.Close()
    fmt.Println("Connected to chat server at localhost:8080")

    // Handle incoming messages from the server
    go func() {
        scanner := bufio.NewScanner(conn)
        for scanner.Scan() {
            fmt.Print(scanner.Text()) // Print received messages
        }
    }()

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
                _, err := conn.Write([]byte(message))
                if err != nil {
                    fmt.Printf("Error sending message: %v\n", err)
                    return
                }
            }
        }
    }
}