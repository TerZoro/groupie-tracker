package api

import (
	"encoding/json"
	"groupie/internal/models"
	"io"
	"net/http"
)

const baseURL = "https://groupietrackers.herokuapp.com/api"

// FetchAllLocations fetches all locations from the global /locations endpoint
func FetchAllLocations() (models.LocationIndex, error) {
	resp, err := http.Get(baseURL + "/locations")
	if err != nil {
		return models.LocationIndex{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.LocationIndex{}, err
	}
	var locationIndex models.LocationIndex
	if err := json.Unmarshal(body, &locationIndex); err != nil {
		return models.LocationIndex{}, err
	}
	return locationIndex, nil
}

// FetchAllDates fetches all dates from the global /dates endpoint
func FetchAllDates() (models.DateIndex, error) {
	resp, err := http.Get(baseURL + "/dates")
	if err != nil {
		return models.DateIndex{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.DateIndex{}, err
	}
	var dateIndex models.DateIndex
	if err := json.Unmarshal(body, &dateIndex); err != nil {
		return models.DateIndex{}, err
	}
	return dateIndex, nil
}

// FetchAllRelations fetches all relations from the global /relation endpoint
func FetchAllRelations() (models.RelationIndex, error) {
	resp, err := http.Get(baseURL + "/relation")
	if err != nil {
		return models.RelationIndex{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.RelationIndex{}, err
	}
	var relationIndex models.RelationIndex
	if err := json.Unmarshal(body, &relationIndex); err != nil {
		return models.RelationIndex{}, err
	}
	return relationIndex, nil
}

// FetchLocations fetches a single artist's locations from their LocationsURL
func FetchLocations(url string) (models.Location, error) {
	resp, err := http.Get(url)
	if err != nil {
		return models.Location{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Location{}, err
	}
	var location models.Location
	if err := json.Unmarshal(body, &location); err != nil {
		return models.Location{}, err
	}
	return location, nil
}

// Similar functions for FetchDates(url) and FetchRelations(url)
