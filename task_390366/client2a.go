// client.go

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	// Authentication
	var username, password string
	reader := bufio.NewReader(os.Stdin)

	// Get username
	fmt.Print("Enter your username: ")
	username, _ = reader.ReadString('\n')
	username = username[:len(username)-1] // Remove newline

	// Get password
	fmt.Print("Enter your password: ")
	password, _ = reader.ReadString('\n')
	password = password[:len(password)-1] // Remove newline

	// Send credentials
	conn.Write([]byte(username + "\n"))
	conn.Write([]byte(password + "\n"))

	go func() {
		reader := bufio.NewReader(conn)
		for {
			msg, _ := reader.ReadString('\n')
			fmt.Print(msg)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		_, err := conn.Write([]byte(text + "\n"))
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}