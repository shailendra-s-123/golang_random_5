package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

const (
	cpuProfileDuration = time.Second * 10
	memProfileDuration = time.Second * 10
	traceProfileDuration = time.Second * 10
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Millisecond * 500) // Simulate work
	w.Write([]byte("Hello, World!"))
}

func main() {
	// Start CPU profile and limit it to a specific duration
	cpuF, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatalf("Could not create CPU profile: %v", err)
	}
	defer cpuF.Close()
	if err := pprof.StartCPUProfile(cpuF); err != nil {
		log.Fatalf("Could not start CPU profile: %v", err)
	}
	defer pprof.StopCPUProfile()
	time.AfterFunc(cpuProfileDuration, func() {
		pprof.StopCPUProfile()
	})

	// Start memory profile and limit it to a specific duration
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

	// Start trace profile and limit it to a specific duration
	traceF, err := os.Create("trace.out")
	if err != nil {
		log.Fatalf("Could not create trace profile: %v", err)
	}
	defer traceF.Close()
	if err := pprof.StartTrace(&pprof.TraceConfig{Output: traceF}); err != nil {
		log.Fatalf("Could not start trace profile: %v", err)
	}
	defer pprof.StopTrace()
	time.AfterFunc(traceProfileDuration, func() {
		pprof.StopTrace()
	})

	// Start monitoring system metrics
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			cpuPercent, err := cpu.Percent(time.Second, false)
			if err != nil {
				log.Printf("Error reading CPU percent: %v", err)
				continue
			}
			memInfo, err := mem.VirtualMemory()
			if err != nil {
				log.Printf("Error reading memory info: %v", err)
				continue
			}
			fmt.Printf("CPU Usage: %f%%\n", cpuPercent[0])
			fmt.Printf("Memory Usage: %d/%d bytes\n", memInfo.Used, memInfo.Total)
			time.Sleep(time.Second)
		}
	}()

	// Set up a simple HTTP server
	http.HandleFunc("/hello", helloHandler)
	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
	wg.Wait()
}