// myapp.go
package main

import (
	"fmt"
	"log"
	"net/http"
)

var (
	db *dbHelper
)

type dbHelper struct {
	// Database fields
}

func (db *dbHelper) Connect() error {
	// Database connection logic
	return nil
}

func (db *dbHelper) Query(query string) ([]string, error) {
	// Database query logic
	return nil, nil
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	data, err := db.Query("SELECT * FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, data)
}

func main() {
	db = new(dbHelper)
	err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handleRoot)
	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}