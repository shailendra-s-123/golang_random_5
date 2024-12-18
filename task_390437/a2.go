package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

func delayedHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate a heavy computation or delay
	time.Sleep(1 * time.Second)
	w.Write([]byte("Hello, World!"))
}

func main() {
	// Start a CPU profile with a sampling rate
	cpuF, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatalf("Could not create CPU profile: %v", err)
	}
	defer cpuF.Close()
	if err := pprof.StartCPUProfile(cpuF); err != nil {
		log.Fatalf("Could not start CPU profile: %v", err)
	}
	defer pprof.StopCPUProfile()

	// Set up an HTTP server
	http.HandleFunc("/delay", delayedHandler)
	fmt.Println("Starting server on :8080")

	// Monitor memory usage periodically
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for _ = range ticker.C {
		var gcStats runtime.GCStats
		runtime.GCStats(&gcStats)
		log.Printf("Total alloc: %d bytes, Total system alloc: %d bytes, Live bytes: %d bytes",
			gcStats.TotalAlloc, gcStats.Sys, gcStats.Mallocs-gcStats.Frees)
	}

	log.Fatal(http.ListenAndServe(":8080", nil))
}