package main

import (
	"log"
	"net/http"
	"os"

	"groupie-tracker/internal/handlers"
)

func main() {
	// Initialize handlers (templates)
	log.Println("Initializing handlers")
	handlers.Init()

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/artist/", handlers.ArtistHandler)
	mux.HandleFunc("/search", handlers.SearchHandler)
	mux.HandleFunc("/static/", handlers.StaticHandler)

	// Get port from environment, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if port[0] != ':' {
		port = ":" + port
	}

	// Start server
	log.Printf("Server starting on %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
