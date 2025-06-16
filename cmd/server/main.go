package main

import (
	"log"

	"wsproxy/internal/proxy"
)

func main() {
	if err := proxy.Start(":5345"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
