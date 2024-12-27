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