package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"time"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second) // Simulate work
	w.Write([]byte("Hello, World!"))
}

func main() {
	// Start a CPU profile
	cpuF, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatalf("Could not create CPU profile: %v", err)
	}
	defer cpuF.Close()
	if err := pprof.StartCPUProfile(cpuF); err != nil {
		log.Fatalf("Could not start CPU profile: %v", err)
	}
	defer pprof.StopCPUProfile()

	// Start a memory profile
	memF, err := os.Create("mem.prof")
	if err != nil {
		log.Fatalf("Could not create memory profile: %v", err)
	}
	defer memF.Close()
	runtime.GC() // Ensure garbage collection is run before profiling
	time.Sleep(time.Second) // Wait for GC to finish
	if err := pprof.WriteHeapProfile(memF); err != nil {
		log.Fatalf("Could not write memory profile: %v", err)
	}

	// Start a trace profile
	traceF, err := os.Create("trace.out")
	if err != nil {
		log.Fatalf("Could not create trace profile: %v", err)
	}
	defer traceF.Close()
	if err := pprof.StartTrace(&pprof.TraceConfig{Output: traceF}); err != nil {
		log.Fatalf("Could not start trace profile: %v", err)
	}
	defer pprof.StopTrace()

	// Set up a simple HTTP server
	http.HandleFunc("/hello", helloHandler)
	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}