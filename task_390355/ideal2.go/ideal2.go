package main

import (
 "log"
 "net/http"
 "time"

 "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
 CheckOrigin: func(r *http.Request) bool {
 	return true
 },
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
 conn, err := upgrader.Upgrade(w, r, nil)
 if err != nil {
 	log.Println("Upgrade error:", err)
 	return
 }
 defer conn.Close()

 pingTicker := time.NewTicker(5 * time.Second)
 defer pingTicker.Stop()

 for {
 	select {
 	case <-pingTicker.C:
 		if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
 			log.Println("Ping error:", err)
 			return
 		}
 	default:
 		messageType, message, err := conn.ReadMessage()
 		if err != nil {
 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
 				log.Println("Read error:", err)
 			}
 			return
 		}
 		log.Printf("Received: %s\n", message)

 		if err = conn.WriteMessage(messageType, message); err != nil {
 			log.Println("Write error:", err)
 			return
 		}
 	}
 }
}

func main() {
 http.HandleFunc("/ws", wsHandler)
 log.Println("WebSocket server started on :8080/ws")
 log.Fatal(http.ListenAndServe(":8080", nil))
}
