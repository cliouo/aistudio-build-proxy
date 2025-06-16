package main

import (
	"log"
	"os"

	"wsproxy/internal/proxy"
)

func main() {
	addr := ":5345"
	if v := os.Getenv("LISTEN_ADDR"); v != "" {
		addr = v
	}
	if err := proxy.Start(addr); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
