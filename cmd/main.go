package main

import (
	"log"
	"net/http"
	"os"

	"groupie-tracker/internal/handlers"
)

func checkRequiredDirs() {
	requiredDirs := []string{"static", "internal/templates"}

	for _, dir := range requiredDirs {
		info, err := os.Stat(dir)
		if os.IsNotExist(err) || !info.IsDir() {
			log.Fatalf("ERROR: Required folder '%s' is missing or not a directory.\n", dir)
		}
		checkFolderNotEmpty(dir)
	}
}

func checkFolderNotEmpty(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("ERROR: Failed to open '%s': %v", path, err)
	}
	defer f.Close()

	files, err := f.Readdirnames(1)
	if err != nil || len(files) == 0 {
		log.Fatalf("ERROR: '%s' folder is empty or unreadable", path)
	}
}

func main() {
	// security measures
	checkRequiredDirs()

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
