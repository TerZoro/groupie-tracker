// package main

// import (
// 	"groupie/internal/handlers"
// 	"html/template"
// 	"log"
// 	"net/http"
// 	"os"
// 	"path/filepath"
// 	"strings"
// )

// var tpl *template.Template

// // Initialize templates in init() for fail-fast behavior
// func init() {
// 	var err error
// 	tpl, err = template.ParseGlob("internal/templates/*.html")
// 	if err != nil {
// 		log.Fatalf("Failed to parse templates: %v", err)
// 	}
// }

// // renderError renders the error.html template with status code and message
// func renderError(w http.ResponseWriter, statusCode int, message string) {
// 	w.WriteHeader(statusCode)
// 	data := struct {
// 		StatusCode int
// 		Message    string
// 	}{
// 		StatusCode: statusCode,
// 		Message:    message,
// 	}
// 	if err := tpl.ExecuteTemplate(w, "error.html", data); err != nil {
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		log.Printf("Template error: %v", err)
// 	}
// }

// // isDirectory checks if the path is a directory
// func isDirectory(path string) bool {
// 	info, err := os.Stat(path)
// 	return err == nil && info.IsDir()
// }

// // loggingMiddleware logs request details
// func loggingMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
// 		next.ServeHTTP(w, r)
// 	})
// }

// // recoverMiddleware catches panics and renders an error page
// func recoverMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		defer func() {
// 			if err := recover(); err != nil {
// 				log.Printf("Panic recovered: %v", err)
// 				renderError(w, http.StatusInternalServerError, "Internal server error")
// 			}
// 		}()
// 		next.ServeHTTP(w, r)
// 	})
// }

// func main() {
// 	// Set up ServeMux for explicit routing
// 	mux := http.NewServeMux()

// 	// Static file serving with directory access check
// 	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
// 		filePath := filepath.Join("static", strings.TrimPrefix(r.URL.Path, "/static/"))
// 		if isDirectory(filePath) {
// 			renderError(w, http.StatusForbidden, "Access to directories is forbidden")
// 			return
// 		}
// 		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP(w, r)
// 	})

// 	// Define routes
// 	mux.HandleFunc("/", handlers.IndexHandler)
// 	mux.HandleFunc("/artist/", handlers.ArtistHandler)
// 	mux.HandleFunc("/api/search", handlers.SearchHandler)
// 	mux.HandleFunc("/concerts", handlers.AllConcertsHandler)

// 	// Pass templates to handlers
// 	handlers.SetTemplates(tpl)

// 	// Wrap mux with middleware
// 	handler := recoverMiddleware(loggingMiddleware(mux))

// 	// Get port from environment or default to :8080
// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = ":8080"
// 	} else if !strings.HasPrefix(port, ":") {
// 		port = ":" + port
// 	}

// 	// Start server
// 	log.Printf("Server starting on %s", port)
// 	if err := http.ListenAndServe(port, handler); err != nil {
// 		log.Fatalf("Server failed: %v", err)
// 	}
// }

package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

// renderError renders the error.html template
func renderError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	data := struct {
		StatusCode int
		Message    string
	}{
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
		filePath := filepath.Join("static", strings.TrimPrefix(r.URL.Path, "/static/"))
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
	mux.HandleFunc("/concerts", handlers.AllConcertsHandler)
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
	if port == "" {
		port = ":8080"
	} else if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	// Start server
	log.Printf("Server starting on %s", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
