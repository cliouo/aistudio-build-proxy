package main

import (
	"flag"
	"log"
	"os"
	"time"

	"wsproxy/internal/browser"
)

func main() {
	wsEndpoint := getenv("CAMOUFOX_WS", "ws://localhost:38835")
	storage := getenv("STORAGE_STATE", "./data/google.json")
	url := getenv("FRONTEND_URL", "")
	token := getenv("AUTH_TOKEN", "")
	intervalStr := getenv("RECONNECT_INTERVAL", "5s")
	bootstrap := flag.Bool("bootstrap", false, "bootstrap login")
	flag.Parse()

	if url == "" {
		log.Fatal("FRONTEND_URL required")
	}
	if token == "" {
		log.Fatal("AUTH_TOKEN required")
	}

	if *bootstrap {
		log.Println("bootstrap mode - manual login expected")
	}

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		interval = 5 * time.Second
	}

	for {
		drv, err := browser.New(wsEndpoint)
		if err != nil {
			log.Printf("driver new error: %v", err)
			time.Sleep(interval)
			continue
		}

		err = drv.KeepAlive(url+"?auth_token="+token, storage)
		drv.Close()
		if err != nil {
			log.Printf("keepalive error: %v", err)
			time.Sleep(interval)
			continue
		}
	}
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
