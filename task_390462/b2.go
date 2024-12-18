
package main  
import (  
        "context"
        "errors"
        "fmt"
        "log"
        "net/url"
        "sync"
        "time"

        "github.com/gorilla/websocket"
        "github.com/kelseyhightower/envconfig"
)

// WebSocketConfig holds configuration for WebSocket client
type WebSocketConfig struct {
        URL                string        `envconfig:"WS_URL" default:"ws://localhost:8080/ws"`
        MaxReconnectionAttempts int         `envconfig:"MAX_RECONNECTION_ATTEMPTS" default:"5"`
        ReconnectionBackoff    time.Duration `envconfig:"RECONNECTION_BACKOFF" default:"1s"`
        ReadTimeout            time.Duration `envconfig:"READ_TIMEOUT" default:"10s"`
        WriteTimeout           time.Duration `envconfig:"WRITE_TIMEOUT" default:"10s"`
        PingPeriod             time.Duration `envconfig:"PING_PERIOD" default:"30s"`
}

// WebSocketClient represents a WebSocket client  
type WebSocketClient struct {  
        config WebSocketConfig
        conn   *websocket.Conn
        connMu sync.Mutex
        connChan chan *websocket.Conn
        disconnectChan chan error
        reconnectCtx context.Context
        reconnectFunc context.CancelFunc
        isConnected bool
}  

// NewWebSocketClient creates a new WebSocket client  
func NewWebSocketClient() (*WebSocketClient, error) {  
        var config WebSocketConfig
        if err := envconfig.Process("ws", &config); err != nil {
                return nil, err
        }
        wsc := &WebSocketClient{
                config: config,
                connChan: make(chan *websocket.Conn),
                disconnectChan: make(chan error),
        }
        return wsc, nil
}  

// Connect establishes a WebSocket connection and starts handling reconnections  
func (wsc *WebSocketClient) Connect(ctx context.Context) {  
        wsc.connMu.Lock()
        defer wsc.connMu.Unlock()

        if wsc.isConnected {
                log.Println("Already connected")
                return
        }

        wsc.reconnectCtx, wsc.reconnectFunc = context.WithCancel(ctx)
        go wsc.reconnect(wsc.reconnectCtx)
        log.Println("Connecting to WebSocket server")
}  

func (wsc *WebSocketClient) reconnect(ctx context.Context) {  
        attempt := 0
        for {  
                select {  
                case <-ctx.Done():  
                        log.Println("Reconnection context done, stopping")  
                        return  
                default:  
                        attempt++
                        if err := wsc.connectOnce(ctx); err != nil {  
                                log.Printf("Reconnect attempt %d failed: %v\n", attempt, err)  
                                if attempt >= wsc.config.MaxReconnectionAttempts {  
                                        log.Println("Exceeded maximum reconnection attempts")  
                                        return  
                                }  
                                backoff := time.Duration(attempt) * wsc.config.ReconnectionBackoff  
                                log.Printf("Reconnecting in %s\n", backoff)  
                                time.Sleep(backoff)  
                        } else {  
                                log.Println("Reconnected successfully")  
                                attempt = 0  
                                return  
                        }  
                }  
        }  
}  

func (wsc *WebSocketClient) connectOnce(ctx context.Context) error {  
        // Connect to the WebSocket server  
        dialer := websocket.Dialer{  
                ReadTimeout:  wsc.config.ReadTimeout,  
                WriteTimeout: wsc.config.WriteTimeout,