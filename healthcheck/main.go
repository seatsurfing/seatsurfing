package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	// Get port from environment variable or use default
	port := getPortFromEnv()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make GET request to health check endpoint
	url := fmt.Sprintf("http://localhost:%s/healthcheck", port)
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Check if status code is 200
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Health check passed")
		os.Exit(0)
	} else {
		fmt.Printf("Health check failed: HTTP status %d\n", resp.StatusCode)
		os.Exit(1)
	}
}

func getPortFromEnv() string {
	// Read PUBLIC_LISTEN_ADDR environment variable
	listenAddr := os.Getenv("PUBLIC_LISTEN_ADDR")
	if listenAddr == "" {
		return "8080"
	}

	// Extract port from address format like "0.0.0.0:8080"
	parts := strings.Split(listenAddr, ":")
	if len(parts) != 2 {
		// Invalid format, return default port
		return "8080"
	}

	port := parts[1]
	if port == "" {
		// Empty port, return default
		return "8080"
	}

	return port
}
