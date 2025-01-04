func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Failed to upgrade connection: %v", err)
        return
    }
    defer conn.Close()

    // Heartbeat mechanism
    go func() {
        for {
            time.Sleep(30 * time.Second) // Send ping every 30 seconds
            if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
                log.Printf("SetWriteDeadline error: %v", err)
                return
            }
            if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
                log.Printf("Ping error: %v", err)
                return
            }
        }
    }()

    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Read error: %v", err)
            break
        }
        log.Printf("Received message: %s", msg)
        // Handle the message and send the response...
    }
}