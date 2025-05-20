package handlers

import (
	"encoding/json"
	"fmt"
	"groupie/internal/api"
	"groupie/internal/models"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var templates *template.Template

// ErrorData standardizes error template data
type ErrorData struct {
	StatusCode int
	Message    string
}

func SetTemplates(tmpl *template.Template) {
	templates = tmpl
}

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
	templates.ExecuteTemplate(w, "index.html", artists)
}

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
	location, err := api.FetchLocations(artist.LocationsURL)
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("Failed to fetch locations: %v", err),
		})
		return
	}
	date, err := api.FetchDates(artist.DatesURL)
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("Failed to fetch dates: %v", err),
		})
		return
	}
	relation, err := api.FetchRelations(artist.RelationsURL)
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

func AllConcertsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusMethodNotAllowed,
			Message:    "Method not allowed",
		})
		return
	}
	relationIndex, err := api.FetchAllRelations()
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("Failed to fetch concerts: %v", err),
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
	type ConcertView struct {
		ArtistName string
		Location   string
		Dates      []string
	}
	var concerts []ConcertView
	artistMap := make(map[int]string)
	for _, artist := range artists {
		artistMap[artist.ID] = artist.Name
	}
	for _, rel := range relationIndex.Index {
		artistName, exists := artistMap[rel.ID]
		if !exists {
			continue
		}
		for loc, dates := range rel.DatesLocations {
			concerts = append(concerts, ConcertView{
				ArtistName: artistName,
				Location:   loc,
				Dates:      dates,
			})
		}
	}
	if t := templates.Lookup("concerts.html"); t == nil {
		templates.ExecuteTemplate(w, "error.html", ErrorData{
			StatusCode: http.StatusInternalServerError,
			Message:    "Template not found",
		})
		return
	}
	templates.ExecuteTemplate(w, "concerts.html", concerts)
}
