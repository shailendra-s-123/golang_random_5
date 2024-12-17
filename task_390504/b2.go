package main

import (
	"fmt"
	"myapp/internal/domain/service"
	"myapp/internal/infrastructure/repository"
)

func main() {
	repo := repository.NewRepository()
	svc := service.NewService(repo)
	result := svc.Execute()
	fmt.Println(result)
}