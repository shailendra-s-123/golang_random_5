package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ugorji/go/codec"
)

func main() {
	file, err := os.Open("large.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	h := codec.NewH(codec.NewHandle())
	d := h.NewDecoder(reader)

	var data interface{}
	for {
		err := d.Decode(&data)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		switch v := data.(type) {
		case string:
			fmt.Println("String:", v)
		case float64:
			fmt.Println("Number:", v)
		case bool:
			fmt.Println("Boolean:", v)
		case nil:
			fmt.Println("Null")
		default:
			fmt.Printf("Unknown: %#v\n", v)
		}
	}
}