package main

import (
	"net/http"
	"runtime/pprof"
)

func main() {
	http.HandleFunc("/pprof/profile", pprof.Profile)
	http.HandleFunc("/pprof/heap", pprof.Heap)
	http.HandleFunc("/pprof/goroutine", pprof.Goroutine)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}