package main

import (
	"fmt"
)

func main() {
	fmt.Println("hello creature ...")
	server := NewApiServer(":8080")
	server.Run()
}
