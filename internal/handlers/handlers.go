package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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

	// Don't populate location data here - it's too slow for the main page
	// Location data will be fetched on-demand for search

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

	// Clean location names in relation data
	cleanedRelation := cleanRelationData(relation)

	// Prepare data for template
	data := struct {
		Artist   models.Artist
		Location models.Location
		Relation models.Relation
	}{
		Artist:   artist,
		Location: location,
		Relation: cleanedRelation,
	}

	err = templates.ExecuteTemplate(w, "artist.html", data)
	if err != nil {
		renderError(w, "Server Error", "Error loading artist page. Please try again later.", 500)
		log.Println("Template error:", err)
	}
}

// populateLocationData fetches and populates location data for all artists
func populateLocationData(artists []models.Artist) []models.Artist {
	for i := range artists {
		relation, err := api.FetchRelation(artists[i].Relations)
		if err != nil {
			log.Printf("Error fetching relation for artist %d: %v", artists[i].ID, err)
			continue
		}

		// Extract location names from relation data
		var locations []string
		for location := range relation.DatesLocations {
			locations = append(locations, models.CleanLocationName(location))
		}
		artists[i].LocationList = locations
	}
	return artists
}

// cleanRelationData cleans location names in relation data
func cleanRelationData(relation models.Relation) models.Relation {
	cleaned := models.Relation{
		ID:             relation.ID,
		DatesLocations: make(map[string][]string),
	}

	for location, dates := range relation.DatesLocations {
		cleanedLocation := models.CleanLocationName(location)
		cleaned.DatesLocations[cleanedLocation] = dates
	}

	return cleaned
}

// renderError renders the error template with proper styling
func renderError(w http.ResponseWriter, title, message string, statusCode int) {
	// Set headers before writing status
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
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
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
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

	// Only populate location data for search results, not all artists
	// This is much faster than populating for all artists
	searchQuery := strings.ToLower(query)
	var results []models.Artist

	// First pass: search without location data
	for _, artist := range artists {
		if strings.Contains(strings.ToLower(artist.Name), searchQuery) ||
			strings.Contains(strings.ToLower(artist.FirstAlbum), searchQuery) ||
			strings.Contains(strconv.Itoa(artist.CreationDate), searchQuery) ||
			containsAny(artist.Members, searchQuery) {
			results = append(results, artist)
		}
	}

	// If no results found, try with location data (slower but more comprehensive)
	if len(results) == 0 {
		// Populate location data only for search
		artistsWithLocations := populateLocationData(artists)
		for _, artist := range artistsWithLocations {
			if strings.Contains(artist.GetSearchableText(), searchQuery) {
				results = append(results, artist)
			}
		}
	}

	err = templates.ExecuteTemplate(w, "index.html", results)
	if err != nil {
		renderError(w, "Server Error", "Error loading search results. Please try again later.", 500)
		log.Println("Template error:", err)
	}
}

// containsAny checks if any member contains the search query
func containsAny(members []string, query string) bool {
	for _, member := range members {
		if strings.Contains(strings.ToLower(member), query) {
			return true
		}
	}
	return false
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

	// Don't populate location data here - it's too slow
	// Frontend will handle location search separately

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300") // Cache for 5 minutes

	// Convert to JSON and send
	jsonData, err := json.Marshal(artists)
	if err != nil {
		http.Error(w, "Failed to encode artists data", 500)
		log.Println("JSON encoding error:", err)
		return
	}

	w.Write(jsonData)
}

// APICacheStatusHandler returns cache status information
func APICacheStatusHandler(w http.ResponseWriter, r *http.Request) {
	isCached, lastUpdate := api.GetCacheStatus()

	status := struct {
		IsCached   bool      `json:"isCached"`
		LastUpdate time.Time `json:"lastUpdate"`
	}{
		IsCached:   isCached,
		LastUpdate: lastUpdate,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// APILocationSearchHandler handles location-based search
func APILocationSearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", 400)
		return
	}

	// Use the optimized location search
	matchingArtistIDs, err := api.SearchLocations(query)
	if err != nil {
		http.Error(w, "Failed to search locations", 500)
		log.Println("Error searching locations:", err)
		return
	}

	if len(matchingArtistIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]models.Artist{})
		return
	}

	// Get the matching artists
	allArtists, err := api.FetchArtists()
	if err != nil {
		http.Error(w, "Failed to load artists", 500)
		log.Println("Error fetching artists:", err)
		return
	}

	// Create a map for fast lookup
	artistMap := make(map[int]models.Artist)
	for _, artist := range allArtists {
		artistMap[artist.ID] = artist
	}

	// Build results
	var results []models.Artist
	for _, artistID := range matchingArtistIDs {
		if artist, exists := artistMap[artistID]; exists {
			results = append(results, artist)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// APILocationSuggestionsHandler returns location suggestions for search
func APILocationSuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", 400)
		return
	}

	limit := 5 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	suggestions, err := api.GetLocationSuggestions(query, limit)
	if err != nil {
		http.Error(w, "Failed to get location suggestions", 500)
		log.Println("Error getting location suggestions:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

// APIClearCacheHandler clears the cache
func APIClearCacheHandler(w http.ResponseWriter, r *http.Request) {
	api.ClearCache()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Cache cleared successfully"})
}
