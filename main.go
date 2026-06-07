package main

import (
	"fmt"
)

// main is the entry point for our RedisLite server.
func main() {
	fmt.Println("RedisLite starting up...")
	
	server := NewServer(":6380")
	if err := server.Start(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
