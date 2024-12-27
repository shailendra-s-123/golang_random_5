package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	serverAddress := "localhost:9000"
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	// Receive and display the server's response
	go receiveMessages(conn)

	// Handle user input for sending messages
	sendMessages(conn)
}

// Receive and print messages from the server
func receiveMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading from server: %v", err)
		}
		fmt.Print("Server: " + message)
	}
}

// Send messages to the server
func sendMessages(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("You: ")
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message) // Remove any surrounding spaces
		_, err := conn.Write([]byte(message + "\n"))
		if err != nil {
			log.Fatalf("Error sending message to server: %v", err)
		}
		if message == "exit" {
			break
		}
	}
}