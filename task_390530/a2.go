// legacy.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"database/sql"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

var db *sql.DB

func main() {
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/mydatabase")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/", handleRoot)
	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var data []string
	for rows.Next() {
		var item string
		if err := rows.Scan(&item); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data = append(data, item)
	}

	fmt.Fprintf(w, "Users: %v", data)
}