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
	if len(os.Args) != 2 {
		log.Fatalln("Usage: client <server_address>")
	}

	conn, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	fmt.Print("Please enter your username: ")
	username, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Please enter your password: ")
	password, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	password = strings.TrimSpace(password)

	conn.Write([]byte(username + "\n"))
	conn.Write([]byte(password + "\n"))

	go func() {
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