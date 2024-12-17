package main

import (
    "fmt"
    "myapp/repository"
    "myapp/service"
)

func main() {
    repo := repository.NewRepository()
    svc := service.NewService(repo)

    result := svc.DoSomething()
    fmt.Println(result)
}