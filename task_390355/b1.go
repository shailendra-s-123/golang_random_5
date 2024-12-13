package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/handler"
	"github.com/graph-gophers/graphql-go/introspection"
	"github.com/graph-gophers/graphql-go/playground"
	"github.com/joho/godotenv"
	jwt "github.com/dgrijalva/jwt-go"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // For simplicity, allow all origins in this example
		},
	}
	secretKey = []byte("your_secret_key_here") // Replace with a secure key
)

type Query struct {
}

type Mutation struct {
}

type Subscription struct {
}

var schema = graphql.NewSchema(graphql.SchemaConfig{
	Query:    graphql.NewObject(graphql.ObjectConfig{Name: "Query", Type: graphql.EmptyInterface()}),
	Mutation: graphql.NewObject(graphql.ObjectConfig{Name: "Mutation", Type: graphql.EmptyInterface()}),
	Subscription: graphql.NewObject(graphql.ObjectConfig{
		Name: "Subscription",
		Type: graphql.EmptyInterface(),
	}),
	Fields: graphql.Fields{
		"__schema": &graphql.Field{
			Type: introspection.Schema.Type(),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return introspection.Schema(p.Context), nil
			},
		},
	},
})

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Origin")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		var requestHandler handler.Handler
		requestHandler = handler.New(&handler.Config{
			Schema: &schema,
			Playground: true,
			GraphiQL:  true,
			Context: func(r *http.Request) context.Context {
				ctx := r.Context()
				token, err := getTokenFromHeader(r)
				if err != nil {
					log.Println("Error getting token:", err)
					return ctx
				}

				claims, err := verifyToken(token)
				if err != nil {
					log.Println("Error verifying token:", err)
					return ctx
				}

				return context.WithValue(ctx, "user", claims)
			},
		})

		requestHandler.ServeHTTP(w, r)
	})

	http.HandleFunc("/websocket", handleWebSocket)

	log.Println("Server starting on port 4000")
	err = http.ListenAndServeTLS(":4000", "server.crt", "server.key", nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Origin")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	token, err := getTokenFromHeader(r)
	if err != nil {
		log.Println("Error getting token for WebSocket:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Unauthorized"))
		return
	}

	claims, err := verifyToken(token)
	if err != nil {
		log.Println("Error verifying token for WebSocket:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Unauthorized"))
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading WebSocket message:", err)
			break
		}

		var query struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(message, &query); err != nil {
			log.Println("Error unmarshaling WebSocket message:", err)
			continue
		}

		result := executeQuery(ctx, query.Query)
		response, err := json.Marshal(result)
		if err != nil {
			log.Println("Error marshaling response:", err)
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, response)
		if err != nil {
			log.Println("Error writing WebSocket response:", err)
			break
		}
	}
}

func executeQuery(ctx context.Context, query string) interface{} {
	params := graphql.Params{
		Schema:        &schema,
		RequestString: query,
		Context:       ctx,
	}
	return graphql.Do(params)
}

func getTokenFromHeader(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", fmt.Errorf("token not found in header")
	}
	return strings.TrimPrefix(token, "Bearer "), nil
}

func verifyToken(tokenString string) (*jwt.StandardClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHS256); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}