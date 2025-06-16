package main

import (
	"flag"
	"log"
	"os"
	"time"

	"wsproxy/internal/browser"
)

func main() {
	wsFlag := flag.String("ws-endpoint", "", "Camoufox WS endpoint")
	urlFlag := flag.String("url", "", "browser url")
	bootstrap := flag.Bool("bootstrap", false, "bootstrap login mode")
	flag.Parse()

	ws := *wsFlag
	if ws == "" {
		ws = os.Getenv("CAMOUFOX_WS")
	}
	storage := os.Getenv("STORAGE_STATE")
	if storage == "" {
		storage = "./data/google.json"
	}
	url := *urlFlag
	if url == "" {
		url = os.Getenv("BROWSER_URL")
	}
	reconnect := time.Second * 5
	if v := os.Getenv("RECONNECT_INTERVAL"); v != "" {
		if dur, err := time.ParseDuration(v); err == nil {
			reconnect = dur
		}
	}

	for {
		driver, err := browser.New(ws)
		if err != nil {
			log.Printf("connect camoufox: %v", err)
			time.Sleep(reconnect)
			continue
		}
		err = driver.KeepAlive(url, storage, *bootstrap)
		driver.Close()
		if err != nil {
			log.Printf("keepalive failed: %v", err)
			time.Sleep(reconnect)
			continue
		}
	}
}
