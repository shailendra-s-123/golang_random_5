package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Example struct for database row
type User struct {
	ID   int
	Name string
}

func main() {
	// Set up the database connection
	dsn := "user:password@tcp(127.0.0.1:3306)/dbname"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Set up connection pool limits
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 3)

	var wg sync.WaitGroup
	users := make(chan User, 10) // Buffered channel to handle users

	// Fetch User data concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()
			// Simulating a query
			rows, err := db.Query("SELECT id, name FROM users LIMIT 10")
			if err != nil {
				log.Printf("Worker %d: Error querying database: %v\n", workerID, err)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var user User
				if err := rows.Scan(&user.ID, &user.Name); err != nil {
					log.Printf("Worker %d: Error scanning row: %v\n", workerID, err)
					return
				}
				users <- user // Send user to channel
			}
		}(i)
	}

	// Close the users channel once all Goroutines are done
	go func() {
		wg.Wait()
		close(users)
	}()

	// Handle results
	for user := range users {
		fmt.Printf("Received user: ID=%d, Name=%s\n", user.ID, user.Name)
	}
}