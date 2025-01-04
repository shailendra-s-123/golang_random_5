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
        case msg, err := conn.ReadMessage(); err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.ClosePong) {
                log.Println("Read error:", err)
            }
            return
        } else if msg[0] == 0x8a { // Check if the message is a Pong
            continue
        }

        // Handle data message
    }
}