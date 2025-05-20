package api

import (
	"encoding/json"
	"fmt"
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

// extractArtistID extracts the artist ID from a URL
func extractArtistID(url string) (int, error) {
	idStr := url[strings.LastIndex(url, "/")+1:]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid artist ID in URL %s: %w", url, err)
	}
	return id, nil
}

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
		return nil, fmt.Errorf("fetching artists: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read artists response: %v", err)
		return nil, fmt.Errorf("reading artists response: %w", err)
	}
	var fetchedArtists []models.Artist
	if err := json.Unmarshal(body, &fetchedArtists); err != nil {
		log.Printf("Failed to unmarshal artists: %v", err)
		return nil, fmt.Errorf("unmarshaling artists: %w", err)
	}

	// Validate artists
	for _, artist := range fetchedArtists {
		if err := artist.Validate(); err != nil {
			log.Printf("Invalid artist data: %v", err)
			continue
		}
	}

	// Update cache
	cache.SetArtists(fetchedArtists)
	return fetchedArtists, nil
}

// FetchArtistByID fetches a single artist by ID
func FetchArtistByID(id int) (models.Artist, error) {
	if id <= 0 {
		return models.Artist{}, fmt.Errorf("invalid artist ID: %d", id)
	}

	// Check cache
	artist, exists, lastUpdated := cache.GetArtistByID(id)
	if exists && time.Since(lastUpdated) < cacheTTL {
		return artist, nil
	}

	// Try fetching directly from API
	resp, err := http.Get(fmt.Sprintf("%s/artists/%d", baseURL, id))
	if err != nil {
		log.Printf("Failed to fetch artist %d: %v", id, err)
		return models.Artist{}, fmt.Errorf("fetching artist %d: %w", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return models.Artist{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return models.Artist{}, fmt.Errorf("unexpected status code for artist %d: %d", id, resp.StatusCode)
	}
	var fetchedArtist models.Artist
	if err := json.NewDecoder(resp.Body).Decode(&fetchedArtist); err != nil {
		log.Printf("Failed to unmarshal artist %d: %v", id, err)
		return models.Artist{}, fmt.Errorf("unmarshaling artist %d: %w", id, err)
	}
	if err := fetchedArtist.Validate(); err != nil {
		log.Printf("Invalid artist data for ID %d: %v", id, err)
		return models.Artist{}, nil
	}

	// Update cache
	cache.UpdateArtist(id, fetchedArtist)
	return fetchedArtist, nil
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
		return models.LocationIndex{}, fmt.Errorf("fetching locations: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return models.LocationIndex{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read locations response: %v", err)
		return models.LocationIndex{}, fmt.Errorf("reading locations response: %w", err)
	}
	var locationIndex models.LocationIndex
	if err := json.Unmarshal(body, &locationIndex); err != nil {
		log.Printf("Failed to unmarshal locations: %v", err)
		return models.LocationIndex{}, fmt.Errorf("unmarshaling locations: %w", err)
	}

	// Update cache
	cache.SetLocations(locationIndex)
	return locationIndex, nil
}

// FetchLocations fetches a single artist's locations
func FetchLocations(url string) (models.Location, error) {
	id, err := extractArtistID(url)
	if err != nil {
		log.Printf("Invalid artist ID: %v", err)
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
		return models.Location{}, fmt.Errorf("fetching locations from %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return models.Location{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return models.Location{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read locations response: %v", err)
		return models.Location{}, fmt.Errorf("reading locations response: %w", err)
	}
	var location models.Location
	if err := json.Unmarshal(body, &location); err != nil {
		var locIndex models.LocationIndex
		if err := json.Unmarshal(body, &locIndex); err != nil {
			log.Printf("Failed to unmarshal locations: %v", err)
			return models.Location{}, fmt.Errorf("unmarshaling locations: %w", err)
		}
		for _, loc := range locIndex.Index {
			if loc.ID == id {
				location = loc
				break
			}
		}
		if location.ID == 0 {
			return models.Location{}, nil // Not found
		}
	}

	// Update cache
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
		return models.DateIndex{}, fmt.Errorf("fetching dates: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return models.DateIndex{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read dates response: %v", err)
		return models.DateIndex{}, fmt.Errorf("reading dates response: %w", err)
	}
	var dateIndex models.DateIndex
	if err := json.Unmarshal(body, &dateIndex); err != nil {
		log.Printf("Failed to unmarshal dates: %v", err)
		return models.DateIndex{}, fmt.Errorf("unmarshaling dates: %w", err)
	}

	// Update cache
	cache.SetDates(dateIndex)
	return dateIndex, nil
}

// FetchDates fetches a single artist's dates
func FetchDates(url string) (models.Date, error) {
	id, err := extractArtistID(url)
	if err != nil {
		log.Printf("Invalid artist ID: %v", err)
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
		return models.Date{}, fmt.Errorf("fetching dates from %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return models.Date{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return models.Date{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read dates response: %v", err)
		return models.Date{}, fmt.Errorf("reading dates response: %w", err)
	}
	var date models.Date
	if err := json.Unmarshal(body, &date); err != nil {
		var dateIndex models.DateIndex
		if err := json.Unmarshal(body, &dateIndex); err != nil {
			log.Printf("Failed to unmarshal dates: %v", err)
			return models.Date{}, fmt.Errorf("unmarshaling dates: %w", err)
		}
		for _, d := range dateIndex.Index {
			if d.ID == id {
				date = d
				break
			}
		}
		if date.ID == 0 {
			return models.Date{}, nil
		}
	}

	// Update cache
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
		return models.RelationIndex{}, fmt.Errorf("fetching relations: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return models.RelationIndex{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read relations response: %v", err)
		return models.RelationIndex{}, fmt.Errorf("reading relations response: %w", err)
	}
	var relationIndex models.RelationIndex
	if err := json.Unmarshal(body, &relationIndex); err != nil {
		log.Printf("Failed to unmarshal relations: %v", err)
		return models.RelationIndex{}, fmt.Errorf("unmarshaling relations: %w", err)
	}

	// Update cache
	cache.SetRelations(relationIndex)
	return relationIndex, nil
}

// FetchRelations fetches a single artist's relations
func FetchRelations(url string) (models.Relation, error) {
	id, err := extractArtistID(url)
	if err != nil {
		log.Printf("Invalid artist ID: %v", err)
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
		return models.Relation{}, fmt.Errorf("fetching relations from %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return models.Relation{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return models.Relation{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read relations response: %v", err)
		return models.Relation{}, fmt.Errorf("reading relations response: %w", err)
	}
	var relation models.Relation
	if err := json.Unmarshal(body, &relation); err != nil {
		var relIndex models.RelationIndex
		if err := json.Unmarshal(body, &relIndex); err != nil {
			log.Printf("Failed to unmarshal relations: %v", err)
			return models.Relation{}, fmt.Errorf("unmarshaling relations: %w", err)
		}
		for _, r := range relIndex.Index {
			if r.ID == id {
				relation = r
				break
			}
		}
		if relation.ID == 0 {
			return models.Relation{}, nil
		}
	}

	// Update cache
	cache.UpdateRelation(id, relation)
	return relation, nil
}

// RefreshCache forces a refresh of all cached data
func RefreshCache() error {
	cache.Clear()
	artists, err := FetchArtists()
	if err != nil {
		log.Printf("Failed to refresh artists cache: %v", err)
		return fmt.Errorf("refreshing artists cache: %w", err)
	}
	locations, err := FetchAllLocations()
	if err != nil {
		log.Printf("Failed to refresh locations cache: %v", err)
		return fmt.Errorf("refreshing locations cache: %w", err)
	}
	dates, err := FetchAllDates()
	if err != nil {
		log.Printf("Failed to refresh dates cache: %v", err)
		return fmt.Errorf("refreshing dates cache: %w", err)
	}
	relations, err := FetchAllRelations()
	if err != nil {
		log.Printf("Failed to refresh relations cache: %v", err)
		return fmt.Errorf("refreshing relations cache: %w", err)
	}
	log.Printf("Cache refreshed: %d artists, %d locations, %d dates, %d relations",
		len(artists), len(locations.Index), len(dates.Index), len(relations.Index))
	return nil
}
