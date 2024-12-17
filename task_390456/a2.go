package main

import (
    "fmt"
    "github.com/example/privateModule"
)

func main() {
    // Use a function from the private module
    result := privateModule.SayHello("World")
    fmt.Println(result)
}