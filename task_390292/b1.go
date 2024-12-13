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

func (f *Fetcher) Fetch(ctx context.Context, out chan<- Data) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.Url, nil)
	if err != nil {
		log.Printf("fetch: %v", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("fetch: %v", err)
		return
	}
	defer resp.Body.Close()

	// Simulate reading response, here we just send a dummy data
	out <- Data{Raw: "some raw data"}

}

// Processor defines a stage in the processing pipeline.
type Processor struct {
}

func (p *Processor) Process(in <-chan Data, out chan<- Data) {
	for data := range in {
		// Simulate processing
		time.Sleep(time.Millisecond * 100) // simulate processing delay
		data.Processed = fmt.Sprintf("processed: %s", data.Raw)
		out <- data
	}
}

// Writer defines a stage for sending data to a database.
type Writer struct {
}

func (w *Writer) Write(in <-chan Data) {
	for data := range in {
		if len(data.Errors) > 0 {
			log.Printf("error processing data: %v", data.Errors)
			continue
		}
		// Simulate writing to a database
		fmt.Printf("writing to database: %s\n", data.Processed)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fetcher := &Fetcher{Url: "http://example.com/api/data"}
	processor := &Processor{}
	writer := &Writer{}

	dataChan := make(chan Data)
	processedChan := make(chan Data)

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(dataChan)
				return
			default:
				fetcher.Fetch(ctx, dataChan)
			}
		}
	}()

	go processor.Process(dataChan, processedChan)

	go writer.Write(processedChan)

	// Let the pipeline run for some time, simulating data streaming
	time.Sleep(time.Second * 2)
}
