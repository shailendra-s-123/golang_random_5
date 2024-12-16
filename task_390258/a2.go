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
type BatchItem struct {
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
	batch := make([]BatchItem, 1000) // Example batch of 1000 items

	// Simulate populating the batch
	for i := range batch {
		batch[i] = BatchItem{ID: i + 1, Name: fmt.Sprintf("Item %d", i + 1)}
	}

	// Write batch items concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()
			for {
				selectedItem := <-batchChannel
				if selectedItem == nil {
					return
				}

				err := writeItem(db, selectedItem)
				if err != nil {
					log.Printf("Worker %d: Failed to write item %d: %v. Retrying...\n", workerID, selectedItem.ID, err)
					time.Sleep(time.Second) // Simulate backoff time
					continue
				}
				log.Printf("Worker %d: Successfully wrote item %d\n", workerID, selectedItem.ID)
			}
		}(i)
	}

	// Close the batch channel once all items are processed
	go func() {
		wg.Wait()
		close(batchChannel)
	}()

	// Distribute batch items to workers
	for _, item := range batch {
		batchChannel <- &item
	}
}

func writeItem(db *sql.DB, item *BatchItem) error {
	// Simulating a write operation
	stmt, err := db.Prepare("INSERT INTO table_name (id, name) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(item.ID, item.Name)
	return err
}

// Channel to distribute batch items to worker Goroutines
var batchChannel = make(chan *BatchItem, 100)