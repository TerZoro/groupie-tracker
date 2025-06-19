package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"groupie-tracker/internal/api"
	"groupie-tracker/internal/models"
)

var tpl *template.Template

// InitTemplates loads your HTML only once.
func InitTemplates(pattern string) {
	var err error
	tpl, err = template.ParseGlob(pattern)
	if err != nil {
		log.Fatalf("failed to parse templates %q: %v", pattern, err)
	}
}

// render is your single place to exec a template or error out.
func render(w http.ResponseWriter, name string, data interface{}) {
	if err := tpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "template error", 500)
		log.Printf("template %s exec error: %v", name, err)
	}
}

// Home shows all artists.
func Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	artists, err := api.FetchArtists()
	if err != nil {
		http.Error(w, "unable to fetch artists", 500)
		log.Println("FetchArtists:", err)
		return
	}
	render(w, "index.html", artists)
}

// Artist shows one artist's detail.
func Artist(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/artist/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid artist ID", 400)
		return
	}
	artists, err := api.FetchArtists()
	if err != nil {
		http.Error(w, "unable to fetch artists", 500)
		return
	}
	var a models.Artist
	for _, art := range artists {
		if art.ID == id {
			a = art
			break
		}
	}
	if a.ID == 0 {
		http.NotFound(w, r)
		return
	}
	loc, _ := api.FetchLocation(a.Locations)
	rel, _ := api.FetchRelation(a.Relations)
	render(w, "artist.html", struct {
		models.Artist
		models.Location
		models.Relation
	}{a, loc, rel})
}

// Search filters artists by query "q".
func Search(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	if q == "" {
		http.Redirect(w, r, "/", 302)
		return
	}
	artists, err := api.FetchArtists()
	if err != nil {
		http.Error(w, "unable to fetch artists", 500)
		return
	}
	var out []models.Artist
	for _, art := range artists {
		if strings.Contains(strings.ToLower(art.Name), q) ||
			strings.Contains(strings.ToLower(art.FirstAlbum), q) {
			out = append(out, art)
		}
	}
	render(w, "index.html", out)
}
