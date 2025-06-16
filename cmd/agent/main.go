package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"time"

	"wsproxy/internal/browser"
)

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func main() {
	wsEndpoint := flag.String("ws-endpoint", getenv("CAMOUFOX_WS", "ws://localhost:38835"), "Camoufox WebSocket endpoint")
	bootstrap := flag.Bool("bootstrap", false, "run in bootstrap mode")
	browserURL := flag.String("url", getenv("BROWSER_URL", ""), "URL to keep alive")
	storageState := flag.String("storage-state", getenv("STORAGE_STATE", "./data/google.json"), "path to storageState.json")
	authToken := getenv("AUTH_TOKEN", "")
	flag.Parse()

	if *browserURL == "" {
		log.Fatalf("BROWSER_URL environment variable or --url flag is required")
	}

	for {
		log.Printf("connecting to Camoufox at %s", *wsEndpoint)
		drv, err := browser.New(*wsEndpoint)
		if err != nil {
			log.Printf("failed to connect to Camoufox: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer drv.Close()

		if *bootstrap {
			log.Println("Bootstrap mode: complete login in the browser then press Enter...")
			go func() {
				reader := bufio.NewReader(os.Stdin)
				_, _ = reader.ReadBytes('\n')
				if err := drv.SaveStorageState(*storageState); err != nil {
					log.Printf("error saving storage state: %v", err)
				} else {
					log.Printf("storage state saved to %s", *storageState)
				}
				os.Exit(0)
			}()
			if err := drv.KeepAlive(*browserURL, *storageState); err != nil {
				log.Printf("bootstrap KeepAlive error: %v", err)
			}
			select {}
		} else {
			if err := drv.KeepAlive(*browserURL, *storageState, browser.WithAuthToken(authToken)); err != nil {
				log.Printf("KeepAlive failed: %v", err)
				drv.Close()
				time.Sleep(5 * time.Second)
				continue
			}
			log.Println("session ended, reconnecting in 5s...")
			drv.Close()
			time.Sleep(5 * time.Second)
		}
	}
}
