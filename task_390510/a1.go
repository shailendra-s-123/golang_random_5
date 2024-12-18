package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/davecgh/go-jsonstream"
)

func main() {
	file, err := os.Open("large.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	parser := jsonstream.NewParser(reader)

	for event := range parser.Stream() {
		switch event.Type {
		case jsonstream.EventStartObject:
			fmt.Println("Starting an object")
		case jsonstream.EventEndObject:
			fmt.Println("Ending an object")
		case jsonstream.EventStartArray:
			fmt.Println("Starting an array")
		case jsonstream.EventEndArray:
			fmt.Println("Ending an array")
		case jsonstream.EventString:
			str, err := event.String()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("String:", str)
		case jsonstream.EventNumber:
			num, err := event.Number()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Number:", num)
		case jsonstream.EventNull:
			fmt.Println("Null value")
		case jsonstream.EventBoolean:
			boolValue, err := event.Boolean()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Boolean:", boolValue)
		default:
			fmt.Println("Unknown event type")
		}
	}

	if err := parser.Err(); err != nil {
		log.Fatal(err)
	}
}