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

// LocationDatePair represents a location and its dates for sorting
type LocationDatePair struct {
	Location string
	Dates    []string
}

// Init initializes the handlers package
func Init() {
	var err error
	templates = template.New("").Funcs(template.FuncMap{
		"replaceUnderscores": func(s string) string {
			parts := strings.Split(strings.ReplaceAll(s, "_", " "), "-")
			for i, part := range parts {
				if part != "" {
					parts[i] = strings.ToUpper(string(part[0])) + part[1:]
				}
			}
			return strings.Join(parts, ", ")
		},
		"sortDates": func(dates []string) []string {
			sorted := make([]string, len(dates))
			copy(sorted, dates)
			sort.Slice(sorted, func(i, j int) bool {
				ti, err1 := time.Parse("2006-01-02", sorted[i])
				tj, err2 := time.Parse("2006-01-02", sorted[j])
				if err1 != nil || err2 != nil {
					log.Printf("sortDates: invalid date %s or %s", sorted[i], sorted[j])
					return sorted[i] > sorted[j]
				}
				return ti.After(tj)
			})
			return sorted
		},
		"sortLocationsByDate": func(datesLocations map[string][]string) []LocationDatePair {
			pairs := make([]LocationDatePair, 0, len(datesLocations))
			for location, dates := range datesLocations {
				if len(dates) == 0 {
					log.Printf("sortLocationsByDate: skipping empty dates for %s", location)
					continue
				}
				pairs = append(pairs, LocationDatePair{
					Location: location,
					Dates:    dates,
				})
			}
			log.Printf("sortLocationsByDate: processing %d locations", len(pairs))
			sort.SliceStable(pairs, func(i, j int) bool {
				maxDateI := getMostRecentDate(pairs[i].Dates)
				maxDateJ := getMostRecentDate(pairs[j].Dates)
				ti, err1 := time.Parse("2006-01-02", maxDateI)
				tj, err2 := time.Parse("2006-01-02", maxDateJ)
				if err1 != nil || maxDateI == "" {
					if err2 != nil || maxDateJ == "" {
						log.Printf("sortLocationsByDate: both dates invalid %s, %s", maxDateI, maxDateJ)
						return pairs[i].Location > pairs[j].Location
					}
					log.Printf("sortLocationsByDate: invalid date %s for %s", maxDateI, pairs[i].Location)
					return false
				}
				if err2 != nil || maxDateJ == "" {
					log.Printf("sortLocationsByDate: invalid date %s for %s", maxDateJ, pairs[j].Location)
					return true
				}
				log.Printf("sortLocationsByDate: comparing %s (%s) vs %s (%s)", pairs[i].Location, maxDateI, pairs[j].Location, maxDateJ)
				return ti.After(tj)
			})
			log.Printf("sortLocationsByDate: sorted %d locations", len(pairs))
			return pairs
		},
	})
	templates, err = templates.ParseGlob("internal/templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
	log.Println("Templates loaded successfully")
}

// getMostRecentDate returns the most recent date from a slice of dates
func getMostRecentDate(dates []string) string {
	var maxDate string
	var maxTime time.Time
	for _, d := range dates {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			log.Printf("getMostRecentDate: invalid date %s", d)
			continue
		}
		if maxDate == "" || t.After(maxTime) {
			maxDate = d
			maxTime = t
		}
	}
	return maxDate
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
	log.Printf("Artist %d: %d locations in DatesLocations", artist.ID, len(relation.DatesLocations))
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