package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	dataDir := envOrDefault("DEMO_DATA_DIR", ".data")
	handler, err := newHandler(dataDir)
	if err != nil {
		log.Fatalf("initialize API: %v", err)
	}

	address := "127.0.0.1:" + envOrDefault("PORT", "8080")
	server := &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("demo-api listening on http://%s", address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
