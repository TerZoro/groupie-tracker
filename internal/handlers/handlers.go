package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"groupie-tracker/internal/api"
	"groupie-tracker/internal/models"
)

var templates *template.Template

// Init loads the HTML templates
func Init() {
	var err error
	templates, err = template.ParseGlob("internal/templates/*.html")
	if err != nil {
		log.Fatal("Failed to load templates:", err)
	}
	log.Println("Templates loaded successfully")
}

// HomeHandler shows the main page with all artists
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		renderError(w, "Page Not Found", "The page you're looking for doesn't exist.", 404)
		return
	}

	artists, err := api.FetchArtists()
	if err != nil {
		renderError(w, "Server Error", "Failed to load artists. Please try again later.", 500)
		log.Println("Error fetching artists:", err)
		return
	}

	err = templates.ExecuteTemplate(w, "index.html", artists)
	if err != nil {
		renderError(w, "Server Error", "Error loading page. Please try again later.", 500)
		log.Println("Template error:", err)
	}
}

// ArtistHandler shows details for a specific artist
func ArtistHandler(w http.ResponseWriter, r *http.Request) {
	// Get artist ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/artist/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		renderError(w, "Invalid Artist ID", "The artist ID you provided is not valid. Please try again.", 400)
		return
	}

	// Get all artists and find the one we need
	artists, err := api.FetchArtists()
	if err != nil {
		renderError(w, "Server Error", "Failed to load artists. Please try again later.", 500)
		return
	}

	var artist models.Artist
	found := false
	for _, a := range artists {
		if a.ID == id {
			artist = a
			found = true
			break
		}
	}

	if !found {
		renderError(w, "Artist Not Found", "The artist you're looking for doesn't exist. Please check the URL and try again.", 404)
		return
	}

	// Get additional data
	location, err := api.FetchLocation(artist.Locations)
	if err != nil {
		log.Println("Error fetching location:", err)
	}

	relation, err := api.FetchRelation(artist.Relations)
	if err != nil {
		log.Println("Error fetching relation:", err)
	}

	// Prepare data for template
	data := struct {
		Artist   models.Artist
		Location models.Location
		Relation models.Relation
	}{
		Artist:   artist,
		Location: location,
		Relation: relation,
	}

	err = templates.ExecuteTemplate(w, "artist.html", data)
	if err != nil {
		renderError(w, "Server Error", "Error loading artist page. Please try again later.", 500)
		log.Println("Template error:", err)
	}
}

// renderError renders the error template with proper styling
func renderError(w http.ResponseWriter, title, message string, statusCode int) {
	w.WriteHeader(statusCode)

	data := struct {
		Title   string
		Message string
	}{
		Title:   title,
		Message: message,
	}

	err := templates.ExecuteTemplate(w, "error.html", data)
	if err != nil {
		// Fallback to plain text if template fails
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(title + ": " + message))
	}
}

// SearchHandler handles search requests
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Redirect(w, r, "/", 302)
		return
	}

	artists, err := api.FetchArtists()
	if err != nil {
		renderError(w, "Server Error", "Failed to load artists. Please try again later.", 500)
		return
	}

	// Simple search
	var results []models.Artist
	searchQuery := strings.ToLower(query)

	for _, artist := range artists {
		// Search in artist name
		if strings.Contains(strings.ToLower(artist.Name), searchQuery) {
			results = append(results, artist)
			continue
		}

		// Search in members
		for _, member := range artist.Members {
			if strings.Contains(strings.ToLower(member), searchQuery) {
				results = append(results, artist)
				break
			}
		}

		// Search in creation date
		if strings.Contains(strconv.Itoa(artist.CreationDate), searchQuery) {
			results = append(results, artist)
			continue
		}

		// Search in first album
		if strings.Contains(strings.ToLower(artist.FirstAlbum), searchQuery) {
			results = append(results, artist)
			continue
		}
	}

	err = templates.ExecuteTemplate(w, "index.html", results)
	if err != nil {
		renderError(w, "Server Error", "Error loading search results. Please try again later.", 500)
		log.Println("Template error:", err)
	}
}

// StaticHandler serves static files
func StaticHandler(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))).ServeHTTP(w, r)
}

// APIArtistsHandler serves artists data as JSON for frontend
func APIArtistsHandler(w http.ResponseWriter, r *http.Request) {
	artists, err := api.FetchArtists()
	if err != nil {
		http.Error(w, "Failed to load artists", 500)
		log.Println("Error fetching artists:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300") // Cache for 5 minutes

	// Convert to JSON and send
	jsonData, err := json.Marshal(artists)
	if err != nil {
		http.Error(w, "Error encoding data", 500)
		return
	}

	w.Write(jsonData)
}
