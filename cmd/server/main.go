package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"groupie-tracker/internal/api"
	"groupie-tracker/internal/handlers"
	"groupie-tracker/internal/middleware"
)

func main() {
	// Initialize handlers (templates)
	log.Println("Initializing handlers")
	handlers.Init()

	// Preload cache
	if err := api.RefreshCache(); err != nil {
		log.Printf("Warning: Failed to preload cache: %v", err)
	}

	// Setup routes
	mux := handlers.SetupRoutes()

	// Apply middleware
	handler := middleware.SetupMiddleware(mux)

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
