// server.go

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

type Server struct {
	clients     map[string]net.Conn
	clientMutex sync.Mutex
	broadcastCh chan string
	privateCh   chan privateMessage
}

type privateMessage struct {
	recipient string
	message   string
}

func NewServer() *Server {
	return &Server{
		clients:     make(map[string]net.Conn),
		broadcastCh: make(chan string),
		privateCh:   make(chan privateMessage),
	}
}

func (s *Server) Start(address string) {
	go s.startTCPServer(address)
	go s.startHTTPServer()

	// Concurrently handle broadcasts and private messages
	go s.handleBroadcasts()
	go s.handlePrivateMessages()

	select {}
}

func (s *Server) startTCPServer(address string) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error starting TCP server: %v", err)
	}
	defer ln.Close()

	log.Printf("TCP server listening on %s", address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handleClient(conn)
	}
}

func (s *Server) startHTTPServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<h1>Welcome to the Chat Server!</h1><p>Connect to the server via <strong>localhost:8080</strong> for chat.</p>"))
	})

	log.Println("HTTP server started on http://localhost:8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("Enter your username: "))
	var username string
	fmt.Fscan(conn, &username)

	s.clientMutex.Lock()
	s.clients[username] = conn
	s.clientMutex.Unlock()

	log.Printf("New client connected: %s", username)

	s.broadcastCh <- fmt.Sprintf("%s has joined the chat!", username)

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("Error reading from client %s: %v", username, err)
			break
		}
		message := string(buffer[:n])

		if strings.HasPrefix(message, "/private") {
			parts := strings.SplitN(message, " ", 3)
			if len(parts) < 3 {
				conn.Write([]byte("Usage: /private <username> <message>\n"))
				continue
			}

			recipient := parts[1]
			privateMsg := parts[2]
			s.privateCh <- privateMessage{recipient, privateMsg}
		} else {
			s.broadcastCh <- fmt.Sprintf("%s: %s", username, message)
		}
	}

	s.clientMutex.Lock()
	delete(s.clients, username)
	s.clientMutex.Unlock()

	s.broadcastCh <- fmt.Sprintf("%s has left the chat!", username)
}

func (s *Server) handleBroadcasts() {
	for {
		msg := <-s.broadcastCh
		s.clientMutex.Lock()
		for _, conn := range s.clients {
			conn.Write([]byte(msg + "\n"))
		}
		s.clientMutex.Unlock()
	}
}

func (s *Server) handlePrivateMessages() {
	for {
		msg := <-s.privateCh
		s.clientMutex.Lock()
		conn, exists := s.clients[msg.recipient]
		s.clientMutex.Unlock()

		if !exists {
			log.Printf("Private message failed: user %s not found", msg.recipient)
			continue
		}

		conn.Write([]byte("Private message: " + msg.message + "\n"))
	}
}

func main() {
	server := NewServer()
	server.Start(":8080")
}