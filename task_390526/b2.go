package main  
import (  
    "net/http"
    "runtime/pprof"
)
func main() {
    http.HandleFunc("/debug/pprof/", pprof.Index)
    http.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
    http.HandleFunc("/debug/pprof/profile", pprof.Profile)
    http.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
    http.HandleFunc("/debug/pprof/trace", pprof.Trace)
    http.ListenAndServe(":8080", nil)
}