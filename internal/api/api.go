package api

import (
	"encoding/json"
	"groupie-tracker/internal/models"
	"net/http"
)

const baseURL = "https://groupietrackers.herokuapp.com/api"

// FetchArtists gets all artists from the API
func FetchArtists() ([]models.Artist, error) {
	resp, err := http.Get(baseURL + "/artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var artists []models.Artist
	err = json.NewDecoder(resp.Body).Decode(&artists)
	return artists, err
}

// FetchLocation gets location data for an artist
func FetchLocation(url string) (models.Location, error) {
	var location models.Location
	resp, err := http.Get(url)
	if err != nil {
		return location, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&location)
	return location, err
}

// FetchRelation gets relation data for an artist
func FetchRelation(url string) (models.Relation, error) {
	var relation models.Relation
	resp, err := http.Get(url)
	if err != nil {
		return relation, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&relation)
	return relation, err
}
