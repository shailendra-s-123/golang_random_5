package main

import (
    "fmt"
    "net/http"
    "runtime"
    "runtime/pprof"
    "time"
)

func main() {
    // Start CPU profiling
    f, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal(err)
    }
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // Simulate a large-scale application
    go simulateWorkers()

    // Start HTTP server for profiling and monitoring
    http.HandleFunc("/debug/pprof/", pprof.Index)
    http.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
    http.HandleFunc("/debug/pprof/profile", pprof.Profile)
    http.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
    http.HandleFunc("/debug/pprof/trace", pprof.Trace)
    fmt.Println("HTTP server listening on port 6060")
    log.Fatal(http.ListenAndServe(":6060", nil))
}

func simulateWorkers() {
    for {
        numWorkers := runtime.NumCPU()
        for i := 0; i < numWorkers; i++ {
            go doWork()
        }
        time.Sleep(1 * time.Second)
    }
}

func doWork() {
    // Simulate some work that takes time
    time.Sleep(100 * time.Millisecond)
}