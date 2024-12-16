package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // Driver for MySQL database
)

func main() {
	// Establish a connection pool
	db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/database")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Prepare a statement for efficient querying
	stmt, err := db.Prepare("SELECT id, name FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Define the number of Goroutines
	numGoroutines := 10
	waitGroup := sync.WaitGroup{}

	// Function to be executed by each Goroutine
	fetchData := func(wg *sync.WaitGroup) {
		defer wg.Done()

		for {
			// Get a connection from the pool
			conn, err := db.Conn(context.Background())
			if err != nil {
				log.Printf("Error getting connection: %v\n", err)
				return
			}
			defer conn.Close()

			// Execute the query
			rows, err := stmt.QueryContext(context.Background(), conn)
			if err != nil {
				log.Printf("Error querying: %v\n", err)
				continue
			}
			defer rows.Close()

			// Fetch and print the data
			for rows.Next() {
				var id int
				var name string
				err := rows.Scan(&id, &name)
				if err != nil {
					log.Printf("Error scanning: %v\n", err)
					continue
				}
				fmt.Printf("ID: %d, Name: %s\n", id, name)
			}

			// Pause to simulate work
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Start Goroutines
	for i := 0; i < numGoroutines; i++ {
		waitGroup.Add(1)
		go fetchData(&waitGroup)
	}

	// Wait for all Goroutines to finish
	waitGroup.Wait()
}