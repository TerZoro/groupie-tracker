package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"groupie-tracker/internal/api"
	"groupie-tracker/internal/models"
)

var templates *template.Template

// ErrorData standardizes error template data
type ErrorData struct {
	StatusCode int
	Message    string
}

// Init initializes the handlers package
func Init() {
	var err error
	templates = template.New("").Funcs(template.FuncMap{
		"replaceUnderscores": func(s string) string {
			// Convert "playa_del_carmen-mexico" to "Playa Del Carmen, Mexico"
			parts := strings.Split(strings.ReplaceAll(s, "_", " "), "-")
			for i, part := range parts {
				if part != "" {
					parts[i] = strings.ToUpper(string(part[0])) + part[1:]
				}
			}
			return strings.Join(parts, ", ")
		},
		"sortDates": func(dates []string) []string {
			// Create a copy to avoid modifying the original
			sorted := make([]string, len(dates))
			copy(sorted, dates)
			// Sort dates chronologically
			sort.Slice(sorted, func(i, j int) bool {
				ti, err1 := time.Parse("2006-01-02", sorted[i])
				tj, err2 := time.Parse("2006-01-02", sorted[j])
				if err1 != nil || err2 != nil {
					return sorted[i] < sorted[j] // Fallback to string comparison
				}
				return ti.Before(tj)
			})
			return sorted
		},
	})
	templates, err = templates.ParseGlob("internal/templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
	log.Println("Templates loaded successfully")
}

// SetupRoutes configures all routes
func SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Static file serving
	mux.HandleFunc("/static/", serveStatic)

	// Routes
	mux.HandleFunc("/", IndexHandler)
	mux.HandleFunc("/artist/", ArtistHandler)
	mux.HandleFunc("/api/search", SearchHandler)
	mux.HandleFunc("/api/refresh-cache", RefreshCacheHandler)

	return mux
}

// serveStatic handles static file serving
func serveStatic(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Clean(filepath.Join("static", strings.TrimPrefix(r.URL.Path, "/static/")))
	if !strings.HasPrefix(filePath, "static/") {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusForbidden,
			Message:    "Invalid file path",
		})
		return
	}
	if isDirectory(filePath) {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusForbidden,
			Message:    "Access to directories is forbidden",
		})
		return
	}
	http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP(w, r)
}

// isDirectory checks if the path is a directory
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// IndexHandler handles the home page
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusMethodNotAllowed,
			Message:    "Method not allowed",
		})
		return
	}
	if r.URL.Path != "/" {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusNotFound,
			Message:    "Page not found",
		})
		return
	}
	artists, err := api.FetchArtists()
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("Failed to fetch artists: %v", err),
		})
		return
	}
	if t := templates.Lookup("index.html"); t == nil {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusInternalServerError,
			Message:    "Template not found",
		})
		return
	}
	if err := templates.ExecuteTemplate(w, "index.html", artists); err != nil {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusInternalServerError,
			Message:    "Template rendering error",
		})
		log.Printf("Template error: %v", err)
	}
}

// ArtistHandler handles artist detail pages
func ArtistHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusMethodNotAllowed,
			Message:    "Method not allowed",
		})
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/artist/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid artist ID",
		})
		return
	}
	artist, err := api.FetchArtistByID(id)
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("Failed to fetch artist: %v", err),
		})
		return
	}
	if artist.ID == 0 {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusNotFound,
			Message:    "Artist not found",
		})
		return
	}
	// Fetch related data, fail fast on error
	location, err := api.FetchLocations(artist.Locations)
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("Failed to fetch locations: %v", err),
		})
		return
	}
	date, err := api.FetchDates(artist.ConcertDates)
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("Failed to fetch dates: %v", err),
		})
		return
	}
	relation, err := api.FetchRelations(artist.Relations)
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("Failed to fetch relations: %v", err),
		})
		return
	}
	data := struct {
		Artist   models.Artist
		Location models.Location
		Date     models.Date
		Relation models.Relation
	}{
		Artist:   artist,
		Location: location,
		Date:     date,
		Relation: relation,
	}
	if t := templates.Lookup("artist.html"); t == nil {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusInternalServerError,
			Message:    "Template not found",
		})
		return
	}
	if err := templates.ExecuteTemplate(w, "artist.html", data); err != nil {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusInternalServerError,
			Message:    "Template rendering error",
		})
		log.Printf("Template error: %v", err)
	}
}

// SearchHandler handles search requests
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query().Get("q")
	if query == "" || len(query) > 100 {
		http.Error(w, "Query parameter 'q' is required and must be under 100 characters", http.StatusBadRequest)
		return
	}
	artists, err := api.FetchArtists()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch artists: %v", err), http.StatusInternalServerError)
		return
	}
	var results []models.Artist
	query = strings.ToLower(query)
	for _, artist := range artists {
		if strings.Contains(strings.ToLower(artist.Name), query) ||
			strings.Contains(strings.ToLower(strings.Join(artist.Members, " ")), query) ||
			strings.Contains(strings.ToLower(artist.FirstAlbum), query) {
			results = append(results, artist)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("Failed to encode search results: %v", err)
	}
}

// RefreshCacheHandler handles cache refresh requests
func RefreshCacheHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := api.RefreshCache(); err != nil {
		http.Error(w, "Failed to refresh cache", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Cache refreshed"))
}
