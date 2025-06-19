package main

import (
	"log"
	"net/http"

	"groupie-tracker/internal/handlers"
)

func main() {
	// load templates once
	handlers.InitTemplates("templates/*.html")

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Home)
	mux.HandleFunc("/artist/", handlers.Artist)
	mux.HandleFunc("/search", handlers.Search)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("▶️ Starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
