// package api

// import (
// 	"encoding/json"
// 	"groupie/internal/models"
// 	"io"
// 	"net/http"
// )

// const baseURL = "https://groupietrackers.herokuapp.com/api"

// // FetchAllLocations fetches all locations from the global /locations endpoint
// func FetchAllLocations() (models.LocationIndex, error) {
// 	resp, err := http.Get(baseURL + "/locations")
// 	if err != nil {
// 		return models.LocationIndex{}, err
// 	}
// 	defer resp.Body.Close()
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return models.LocationIndex{}, err
// 	}
// 	var locationIndex models.LocationIndex
// 	if err := json.Unmarshal(body, &locationIndex); err != nil {
// 		return models.LocationIndex{}, err
// 	}
// 	return locationIndex, nil
// }

// // FetchAllDates fetches all dates from the global /dates endpoint
// func FetchAllDates() (models.DateIndex, error) {
// 	resp, err := http.Get(baseURL + "/dates")
// 	if err != nil {
// 		return models.DateIndex{}, err
// 	}
// 	defer resp.Body.Close()
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return models.DateIndex{}, err
// 	}
// 	var dateIndex models.DateIndex
// 	if err := json.Unmarshal(body, &dateIndex); err != nil {
// 		return models.DateIndex{}, err
// 	}
// 	return dateIndex, nil
// }

// // FetchAllRelations fetches all relations from the global /relation endpoint
// func FetchAllRelations() (models.RelationIndex, error) {
// 	resp, err := http.Get(baseURL + "/relation")
// 	if err != nil {
// 		return models.RelationIndex{}, err
// 	}
// 	defer resp.Body.Close()
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return models.RelationIndex{}, err
// 	}
// 	var relationIndex models.RelationIndex
// 	if err := json.Unmarshal(body, &relationIndex); err != nil {
// 		return models.RelationIndex{}, err
// 	}
// 	return relationIndex, nil
// }

// // FetchLocations fetches a single artist's locations from their LocationsURL
// func FetchLocations(url string) (models.Location, error) {
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return models.Location{}, err
// 	}
// 	defer resp.Body.Close()
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return models.Location{}, err
// 	}
// 	var location models.Location
// 	if err := json.Unmarshal(body, &location); err != nil {
// 		return models.Location{}, err
// 	}
// 	return location, nil
// }

// // Similar functions for FetchDates(url) and FetchRelations(url)

package api

import (
	"encoding/json"
	"groupie/internal/cache"
	"groupie/internal/models"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://groupietrackers.herokuapp.com/api"

// cacheTTL defines how long cached data is considered fresh (e.g., 1 hour)
const cacheTTL = 1 * time.Hour

// FetchArtists fetches artists from cache or API
func FetchArtists() ([]models.Artist, error) {
	// Check cache
	cachedArtists, lastUpdated := cache.GetArtists()
	if len(cachedArtists) > 0 && time.Since(lastUpdated) < cacheTTL {
		return cachedArtists, nil
	}

	// Fetch from API
	resp, err := http.Get(baseURL + "/artists")
	if err != nil {
		log.Printf("Failed to fetch artists: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read artists response: %v", err)
		return nil, err
	}
	var fetchedArtists []models.Artist
	if err := json.Unmarshal(body, &fetchedArtists); err != nil {
		log.Printf("Failed to unmarshal artists: %v", err)
		return nil, err
	}

	// Update cache
	cache.SetArtists(fetchedArtists)
	return fetchedArtists, nil
}

// FetchArtistByID fetches a single artist by ID
func FetchArtistByID(id int) (models.Artist, error) {
	// Check cache
	artist, exists, lastUpdated := cache.GetArtistByID(id)
	if exists && time.Since(lastUpdated) < cacheTTL {
		return artist, nil
	}

	// Fetch all artists
	allArtists, err := FetchArtists()
	if err != nil {
		return models.Artist{}, err
	}
	for _, a := range allArtists {
		if a.ID == id {
			return a, nil
		}
	}
	return models.Artist{}, nil // Not found
}

// FetchAllLocations fetches all locations from cache or API
func FetchAllLocations() (models.LocationIndex, error) {
	// Check cache
	locations, lastUpdated := cache.GetLocations()
	if len(locations.Index) > 0 && time.Since(lastUpdated) < cacheTTL {
		return locations, nil
	}

	// Fetch from API
	resp, err := http.Get(baseURL + "/locations")
	if err != nil {
		log.Printf("Failed to fetch locations: %v", err)
		return models.LocationIndex{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read locations response: %v", err)
		return models.LocationIndex{}, err
	}
	var locationIndex models.LocationIndex
	if err := json.Unmarshal(body, &locationIndex); err != nil {
		log.Printf("Failed to unmarshal locations: %v", err)
		return models.LocationIndex{}, err
	}

	// Update cache
	cache.SetLocations(locationIndex)
	return locationIndex, nil
}

// FetchLocations fetches a single artist's locations
func FetchLocations(url string) (models.Location, error) {
	// Extract artist ID
	idStr := url[strings.LastIndex(url, "/")+1:]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid artist ID in URL %s: %v", url, err)
		return models.Location{}, err
	}

	// Check cache
	loc, exists, lastUpdated := cache.GetLocationByID(id)
	if exists && time.Since(lastUpdated) < cacheTTL {
		return loc, nil
	}

	// Fetch from API
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch locations from %s: %v", url, err)
		return models.Location{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read locations response: %v", err)
		return models.Location{}, err
	}
	var location models.Location
	if err := json.Unmarshal(body, &location); err != nil {
		// Try unmarshaling as LocationIndex
		var locIndex models.LocationIndex
		if err := json.Unmarshal(body, &locIndex); err != nil {
			log.Printf("Failed to unmarshal locations: %v", err)
			return models.Location{}, err
		}
		if len(locIndex.Index) > 0 {
			location = locIndex.Index[0]
		} else {
			return models.Location{}, nil // Not found
		}
	}

	// Update cache for this artist
	cache.UpdateLocation(id, location)
	return location, nil
}

// FetchAllDates fetches all dates from cache or API
func FetchAllDates() (models.DateIndex, error) {
	// Check cache
	dates, lastUpdated := cache.GetDates()
	if len(dates.Index) > 0 && time.Since(lastUpdated) < cacheTTL {
		return dates, nil
	}

	// Fetch from API
	resp, err := http.Get(baseURL + "/dates")
	if err != nil {
		log.Printf("Failed to fetch dates: %v", err)
		return models.DateIndex{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read dates response: %v", err)
		return models.DateIndex{}, err
	}
	var dateIndex models.DateIndex
	if err := json.Unmarshal(body, &dateIndex); err != nil {
		log.Printf("Failed to unmarshal dates: %v", err)
		return models.DateIndex{}, err
	}

	// Update cache
	cache.SetDates(dateIndex)
	return dateIndex, nil
}

// FetchDates fetches a single artist's dates
func FetchDates(url string) (models.Date, error) {
	// Extract artist ID
	idStr := url[strings.LastIndex(url, "/")+1:]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid artist ID in URL %s: %v", url, err)
		return models.Date{}, err
	}

	// Check cache
	fetchedDate, exists, lastUpdated := cache.GetDateByID(id)
	if exists && time.Since(lastUpdated) < cacheTTL {
		return fetchedDate, nil
	}

	// Fetch from API
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch dates from %s: %v", url, err)
		return models.Date{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read dates response: %v", err)
		return models.Date{}, err
	}
	var date models.Date
	if err := json.Unmarshal(body, &date); err != nil {
		// Try unmarshaling as DateIndex
		var dateIndex models.DateIndex
		if err := json.Unmarshal(body, &dateIndex); err != nil {
			log.Printf("Failed to unmarshal dates: %v", err)
			return models.Date{}, err
		}
		if len(dateIndex.Index) > 0 {
			date = dateIndex.Index[0]
		} else {
			return models.Date{}, nil
		}
	}

	// Update cache for this artist
	cache.UpdateDate(id, date)
	return date, nil
}

// FetchAllRelations fetches all relations from cache or API
func FetchAllRelations() (models.RelationIndex, error) {
	// Check cache
	relations, lastUpdated := cache.GetRelations()
	if len(relations.Index) > 0 && time.Since(lastUpdated) < cacheTTL {
		return relations, nil
	}

	// Fetch from API
	resp, err := http.Get(baseURL + "/relation")
	if err != nil {
		log.Printf("Failed to fetch relations: %v", err)
		return models.RelationIndex{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read relations response: %v", err)
		return models.RelationIndex{}, err
	}
	var relationIndex models.RelationIndex
	if err := json.Unmarshal(body, &relationIndex); err != nil {
		log.Printf("Failed to unmarshal relations: %v", err)
		return models.RelationIndex{}, err
	}

	// Update cache
	cache.SetRelations(relationIndex)
	return relationIndex, nil
}

// FetchRelations fetches a single artist's relations
func FetchRelations(url string) (models.Relation, error) {
	// Extract artist ID
	idStr := url[strings.LastIndex(url, "/")+1:]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid artist ID in URL %s: %v", url, err)
		return models.Relation{}, err
	}

	// Check cache
	rel, exists, lastUpdated := cache.GetRelationByID(id)
	if exists && time.Since(lastUpdated) < cacheTTL {
		return rel, nil
	}

	// Fetch from API
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch relations from %s: %v", url, err)
		return models.Relation{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read relations response: %v", err)
		return models.Relation{}, err
	}
	var relation models.Relation
	if err := json.Unmarshal(body, &relation); err != nil {
		// Try unmarshaling as RelationIndex
		var relIndex models.RelationIndex
		if err := json.Unmarshal(body, &relIndex); err != nil {
			log.Printf("Failed to unmarshal relations: %v", err)
			return models.Relation{}, err
		}
		if len(relIndex.Index) > 0 {
			relation = relIndex.Index[0]
		} else {
			return models.Relation{}, nil
		}
	}

	// Update cache for this artist
	cache.UpdateRelation(id, relation)
	return relation, nil
}

// RefreshCache forces a refresh of all cached data
func RefreshCache() error {
	cache.Clear()
	artists, err := FetchArtists()
	if err != nil {
		log.Printf("Failed to refresh artists cache: %v", err)
		return err
	}
	locations, err := FetchAllLocations()
	if err != nil {
		log.Printf("Failed to refresh locations cache: %v", err)
		return err
	}
	dates, err := FetchAllDates()
	if err != nil {
		log.Printf("Failed to refresh dates cache: %v", err)
		return err
	}
	relations, err := FetchAllRelations()
	if err != nil {
		log.Printf("Failed to refresh relations cache: %v", err)
		return err
	}
	log.Printf("Cache refreshed: %d artists, %d locations, %d dates, %d relations",
		len(artists), len(locations.Index), len(dates.Index), len(relations.Index))
	return nil
}
