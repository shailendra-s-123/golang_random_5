package main

import (
    "fmt"
    "myapp/authentication"
    "myapp/user"
)

func main() {
    u := user.NewUser("JohnDoe", "password123")
    if authentication.Authenticate(u) {
        fmt.Println("Authenticated successfully!")
    } else {
        fmt.Println("Authentication failed!")
    }
}