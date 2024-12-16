package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Record struct {
	ID   int
	Name string
}

const maxRetries = 3

func main() {
	dsn := "user:password@tcp(127.0.0.1:3306)/dbname"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 3)

	records := []Record{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		// Add more records as needed
	}

	batchSize := 10
	var wg sync.WaitGroup

	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]

		wg.Add(1)
		go func(recordsBatch []Record) {
			defer wg.Done()
			if err := writeBatchWithRetries(db, recordsBatch); err != nil {
				log.Printf("Failed to write batch: %v\n", err)
			}
		}(batch)
	}

	wg.Wait()
}

func writeBatchWithRetries(db *sql.DB, records []Record) error {
	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if err = writeBatch(db, records); err == nil {
			return nil
		}
		log.Printf("Retry attempt %d for batch failed: %v\n", attempt+1, err)
		backoff(attempt)
	}
	return fmt.Errorf("all retry attempts failed: %w", err)
}

func writeBatch(db *sql.DB, records []Record) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Ensure rollback on error

	stmt, err := tx.Prepare("INSERT INTO users (id, name) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, record := range records {
		if _, err := stmt.Exec(record.ID, record.Name); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func backoff(attempt int) {
	time.Sleep(time.Duration(rand.Intn(1<<attempt)) * time.Second)
}