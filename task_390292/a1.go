package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Base Pipeline Stage Interface
type PipelineStage interface {
	Process(ctx context.Context, in chan interface{}, out chan interface{}, wg *sync.WaitGroup, errch chan error)
}

// API Reader Stage
type APIReader struct {
	APIUrl string
}

func (ar *APIReader) Process(ctx context.Context, in chan interface{}, out chan interface{}, wg *sync.WaitGroup, errch chan error) {
	defer wg.Done()
	defer close(out)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
			resp, err := http.Get(ar.APIUrl)
			if err != nil {
				errch <- err
				return
			}
			defer resp.Body.Close()

			var data []map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				errch <- err
				return
			}

			for _, d := range data {
				out <- d
			}
		}
	}
}

// Data Processor Stage
type DataProcessor struct{}

func (dp *DataProcessor) Process(ctx context.Context, in chan interface{}, out chan interface{}, wg *sync.WaitGroup, errch chan error) {
	defer wg.Done()
	defer close(out)

	for item := range in {
		selectedData, err := processItem(item)
		if err != nil {
			errch <- err
			continue
		}
		out <- selectedData
	}
}

// Mock Database Writer Stage
type DatabaseWriter struct{}

func (dw *DatabaseWriter) Process(ctx context.Context, in chan interface{}, out chan interface{}, wg *sync.WaitGroup, errch chan error) {
	defer wg.Done()

	for item := range in {
		if err := insertData(item); err != nil {
			errch <- err
		}
	}
}

// Mock data processing function
func processItem(item interface{}) (interface{}, error) {
	itemMap, ok := item.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected item type %T", item)
	}

	selectedData := map[string]interface{}{
		"id":      itemMap["id"],
		"value":   float64(itemMap["value"].(int64)) * 1.1,
		"timestamp": time.Now(),
	}

	return selectedData, nil
}

// Mock database insertion function
func insertData(item interface{}) error {
	itemMap, ok := item.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected item type %T", item)
	}

	// Simulate database insertion logic
	sqlStmt := "INSERT INTO table_name (id, value, timestamp) VALUES (?, ?, ?)"
	// db := setupDatabaseConnection()
	// defer db.Close()
	// preparedStmt, err := db.Prepare(sqlStmt)
	// if err != nil {
	//     return err
	// }
	// defer preparedStmt.Close()
	// _, err = preparedStmt.Exec(itemMap["id"], itemMap["value"], itemMap["timestamp"])
	// return err

	log.Printf("Inserted item: %v\n", itemMap)
	return nil
}

func main() {
	apiUrl := "https://api.example.com/data"

	ctx := context.Background()
	wg := &sync.WaitGroup{}
	errch := make(chan error)

	reader := &APIReader{APIUrl: apiUrl}
	processor := &DataProcessor{}
	writer := &DatabaseWriter{}

	// Initialize all stages
	inChannelReader := make(chan interface{})
	inChannelProcessor := make(chan interface{})
	inChannelWriter := make(chan interface{})

	// Interconnect stages
	wg.Add(1)
	go reader.Process(ctx, inChannelReader, inChannelProcessor, wg, errch)
	wg.Add(1)
	go processor.Process(ctx, inChannelProcessor, inChannelWriter, wg, errch)
	wg.Add(1)
	go writer.Process(ctx, inChannelWriter, nil, wg, errch)

	// Process and wait for completion
	for range errch {
		log.Println("Error in pipeline:", err)
	}

	wg.Wait()
	log.Println("Pipeline finished successfully.")
}