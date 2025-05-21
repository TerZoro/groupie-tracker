package handlers

import (
	"encoding/json"
	"fmt"
	"groupie-tracker/internal/api"
	"groupie-tracker/internal/models"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
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
	templates, err = template.ParseGlob("internal/templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Preload cache
	if err := api.RefreshCache(); err != nil {
		log.Printf("Warning: Failed to preload cache: %v", err)
	}
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
		http.Error(w, "Invalid file path", http.StatusForbidden)
		return
	}
	http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP(w, r)
}

// IndexHandler handles the home page
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	artists, err := api.FetchArtists()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Failed to fetch artists: %v", err)
		return
	}
	if err := templates.ExecuteTemplate(w, "index.html", artists); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
	templates.ExecuteTemplate(w, "artist.html", data)
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
