package main

import (
    "net/http"

    "pkg/api"
)

func main() {
    http.HandleFunc("/user", api.GetUserHandler)
    http.ListenAndServe(":8080", nil)
}