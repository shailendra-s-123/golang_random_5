package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run client.go <server-address>")
		os.Exit(1)
	}

	serverAddr := os.Args[1]

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Send client's name to the server
	name := fmt.Sprintf("%s:%d", os.Args[0], os.Getpid())
	_, err = conn.Write([]byte(name))
	if err != nil {
		log.Println("Error sending name:", err)
		return
	}

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	fmt.Println("Connected to server. Type 'quit' to exit.")

	for {
		// Read input from user
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			break
		}

		// Send message to server
		if _, err = writer.WriteString(message); err != nil {
			log.Println("Error sending message:", err)
			break
		}
		if err = writer.Flush(); err != nil {
			log.Println("Error flushing writer:", err)
			break
		}

		// Check if user wants to quit
		if message == "quit\n" {
			break
		}

		// Read messages from server
		serverMessage, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading server message:", err)
			break
		}
		fmt.Println(serverMessage)
	}
}