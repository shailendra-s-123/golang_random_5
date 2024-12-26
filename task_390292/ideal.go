
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Data represents a single unit of data in the pipeline.
type Data struct {
	Raw         string
	Processed   string
	Errors      []error
}

// Fetcher defines a method for fetching data from an API.
type Fetcher struct {
	Url string
}

func (f *Fetcher) Fetch(ctx context.Context, out chan<- Data, errCh chan<- error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.Url, nil)
	if err != nil {
		errCh <- fmt.Errorf("fetch: %v", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		errCh <- fmt.Errorf("fetch: %v", err)
		return
	}
	defer resp.Body.Close()

	// Simulate reading response, here we just send a dummy data
	out <- Data{Raw: "some raw data"}
}

// Processor defines a stage in the processing pipeline.
type Processor struct {
}

func (p *Processor) Process(in <-chan Data, out chan<- Data, errCh chan<- error) {
	for data := range in {
		// Simulate processing
		time.Sleep(time.Millisecond * 100) // simulate processing delay
		data.Processed = fmt.Sprintf("processed: %s", data.Raw)
		if data.Processed == "" {
			errCh <- fmt.Errorf("processing error for data: %v", data)
			continue
		}
		out <- data
	}
}

// Writer defines a stage for sending data to a database.
type Writer struct {
}

func (w *Writer) Write(in <-chan Data, errCh chan<- error) {
	for data := range in {
		if len(data.Errors) > 0 {
			errCh <- fmt.Errorf("error processing data: %v", data.Errors)
			continue
		}
		// Simulate writing to a database
		fmt.Printf("writing to database: %s\n", data.Processed)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Error channel to propagate errors through the pipeline
	errCh := make(chan error)

	// Create pipeline stages
	fetcher := &Fetcher{Url: "http://example.com/api/data"}
	processor := &Processor{}
	writer := &Writer{}

	// Create channels for each stage
	dataChan := make(chan Data)
	processedChan := make(chan Data)

	// Start goroutines for each pipeline stage
	go func() {
		fetcher.Fetch(ctx, dataChan, errCh)
	}()

	go processor.Process(dataChan, processedChan, errCh)

	go writer.Write(processedChan, errCh)

	// Wait for errors or completion
	go func() {
		for err := range errCh {
			log.Printf("Error in pipeline: %v", err)
		}
	}()

	// Let the pipeline run for some time, simulating data streaming
	time.Sleep(time.Second * 2)

	// Cleanly close channels
	close(dataChan)
	close(processedChan)
	close(errCh)

	log.Println("Pipeline finished.")
}
