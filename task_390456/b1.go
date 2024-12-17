// Example usage in the public module
package main

import (
	"fmt"
	"github.com/example/privateModule"
)

func main() {
	fmt.Println("Using publicModule with privateModule v1.2.3")
	privateModule.Hello()
}