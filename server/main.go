package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"groupie/internal/api"
	"groupie/internal/handlers"
)

var tpl *template.Template

func init() {
	var err error
	tpl, err = template.ParseGlob("internal/templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Preload cache
	if err := api.RefreshCache(); err != nil {
		log.Printf("Warning: Failed to preload cache: %v", err)
	}
}

// ErrorData standardizes error template data
type ErrorData struct {
	StatusCode int
	Message    string
}

// renderError renders the error.html template
func renderError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	data := ErrorData{
		StatusCode: statusCode,
		Message:    message,
	}
	if err := tpl.ExecuteTemplate(w, "error.html", data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
	}
}

// isDirectory checks if the path is a directory
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// loggingMiddleware logs request details
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// recoverMiddleware catches panics
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				renderError(w, http.StatusInternalServerError, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	// Static file serving
	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Clean(filepath.Join("static", strings.TrimPrefix(r.URL.Path, "/static/")))
		if !strings.HasPrefix(filePath, "static/") {
			renderError(w, http.StatusForbidden, "Invalid file path")
			return
		}
		if isDirectory(filePath) {
			renderError(w, http.StatusForbidden, "Access to directories is forbidden")
			return
		}
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP(w, r)
	})

	// Routes
	mux.HandleFunc("/", handlers.IndexHandler)
	mux.HandleFunc("/artist/", handlers.ArtistHandler)
	mux.HandleFunc("/api/search", handlers.SearchHandler)
	mux.HandleFunc("/api/refresh-cache", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			renderError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		if err := api.RefreshCache(); err != nil {
			renderError(w, http.StatusInternalServerError, "Failed to refresh cache")
			return
		}
		w.Write([]byte("Cache refreshed"))
	})

	// Pass templates to handlers
	handlers.SetTemplates(tpl)

	// Apply middleware
	handler := recoverMiddleware(loggingMiddleware(mux))

	// Get port from environment
	port := os.Getenv("PORT")
	if port != "" {
		if _, err := strconv.Atoi(port); err != nil {
			log.Fatalf("Invalid PORT value: %v", err)
		}
		if !strings.HasPrefix(port, ":") {
			port = ":" + port
		}
	} else {
		port = ":8080"
	}

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    port,
		Handler: handler,
	}
	go func() {
		log.Printf("Server starting on %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
}
