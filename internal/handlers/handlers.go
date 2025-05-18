package handlers

import (
	"encoding/json"
	"groupie/internal/api"
	"groupie/internal/models"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

var templates *template.Template

func SetTemplates(tmpl *template.Template) {
	templates = tmpl
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		templates.ExecuteTemplate(w, "error.html", map[string]string{"Error": "Page not found"})
		return
	}
	artists, err := api.FetchArtists()
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", map[string]string{"Error": "Failed to fetch artists"})
		return
	}
	templates.ExecuteTemplate(w, "index.html", artists)
}

func ArtistHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/artist/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", map[string]string{"Error": "Invalid artist ID"})
		return
	}
	artist, err := api.FetchArtistByID(id)
	if err != nil || artist.ID == 0 {
		templates.ExecuteTemplate(w, "error.html", map[string]string{"Error": "Artist not found"})
		return
	}
	// Fetch related data
	location, err := api.FetchLocations(artist.LocationsURL)
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", map[string]string{"Error": "Failed to fetch locations"})
		return
	}
	date, err := api.FetchDates(artist.DatesURL)
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", map[string]string{"Error": "Failed to fetch dates"})
		return
	}
	relation, err := api.FetchRelations(artist.RelationsURL)
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", map[string]string{"Error": "Failed to fetch relations"})
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
	templates.ExecuteTemplate(w, "artist.html", data)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}
	artists, err := api.FetchArtists()
	if err != nil {
		http.Error(w, "Failed to fetch artists", http.StatusInternalServerError)
		return
	}
	var results []models.Artist
	query = strings.ToLower(query)
	for _, artist := range artists {
		if strings.Contains(strings.ToLower(artist.Name), query) {
			results = append(results, artist)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func AllConcertsHandler(w http.ResponseWriter, r *http.Request) {
	relationIndex, err := api.FetchAllRelations()
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", map[string]string{"Error": "Failed to fetch concerts"})
		return
	}
	artists, err := api.FetchArtists()
	if err != nil {
		templates.ExecuteTemplate(w, "error.html", map[string]string{"Error": "Failed to fetch artists"})
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
		for loc, dates := range rel.DatesLocations {
			concerts = append(concerts, ConcertView{
				ArtistName: artistMap[rel.ID],
				Location:   loc,
				Dates:      dates,
			})
		}
	}
	templates.ExecuteTemplate(w, "concerts.html", concerts)
}
