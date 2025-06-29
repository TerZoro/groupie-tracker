package api

import (
	"encoding/json"
	"groupie-tracker/internal/models"
	"net/http"
	"strings"
	"sync"
	"time"
)

const baseURL = "https://groupietrackers.herokuapp.com/api"

// Cache structure for artists data
type ArtistCache struct {
	artists    []models.Artist
	lastUpdate time.Time
	mutex      sync.RWMutex
}

// Cache structure for locations data
type LocationCache struct {
	locations  map[string][]int // location name -> artist IDs
	lastUpdate time.Time
	mutex      sync.RWMutex
}

var (
	artistCache   = &ArtistCache{}
	locationCache = &LocationCache{
		locations: make(map[string][]int),
	}
	cacheTTL = 5 * time.Minute // Cache for 5 minutes
)

// FetchArtists gets all artists from the API with caching
func FetchArtists() ([]models.Artist, error) {
	// Check cache first
	artistCache.mutex.RLock()
	if !artistCache.lastUpdate.IsZero() && time.Since(artistCache.lastUpdate) < cacheTTL {
		artists := make([]models.Artist, len(artistCache.artists))
		copy(artists, artistCache.artists)
		artistCache.mutex.RUnlock()
		return artists, nil
	}
	artistCache.mutex.RUnlock()

	// Fetch fresh data
	artists, err := fetchArtistsFromAPI()
	if err != nil {
		return nil, err
	}

	// Update cache
	artistCache.mutex.Lock()
	artistCache.artists = make([]models.Artist, len(artists))
	copy(artistCache.artists, artists)
	artistCache.lastUpdate = time.Now()
	artistCache.mutex.Unlock()

	return artists, nil
}

// fetchArtistsFromAPI gets all artists from the API without caching
func fetchArtistsFromAPI() ([]models.Artist, error) {
	resp, err := http.Get(baseURL + "/artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var artists []models.Artist
	err = json.NewDecoder(resp.Body).Decode(&artists)
	return artists, err
}

// FetchAllLocations gets all locations data and builds a fast search index
func FetchAllLocations() (map[string][]int, error) {
	// Check cache first
	locationCache.mutex.RLock()
	if !locationCache.lastUpdate.IsZero() && time.Since(locationCache.lastUpdate) < cacheTTL {
		// Return a copy of the cached data
		result := make(map[string][]int)
		for k, v := range locationCache.locations {
			result[k] = make([]int, len(v))
			copy(result[k], v)
		}
		locationCache.mutex.RUnlock()
		return result, nil
	}
	locationCache.mutex.RUnlock()

	// Fetch fresh data
	locations, err := fetchAllLocationsFromAPI()
	if err != nil {
		return nil, err
	}

	// Update cache
	locationCache.mutex.Lock()
	locationCache.locations = locations
	locationCache.lastUpdate = time.Now()
	locationCache.mutex.Unlock()

	return locations, nil
}

// fetchAllLocationsFromAPI fetches all locations data from the API
func fetchAllLocationsFromAPI() (map[string][]int, error) {
	// Fetch all relations data
	resp, err := http.Get(baseURL + "/relation")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var relationIndex models.RelationIndex
	err = json.NewDecoder(resp.Body).Decode(&relationIndex)
	if err != nil {
		return nil, err
	}

	// Build location index
	locationIndex := make(map[string][]int)

	for _, relation := range relationIndex.Index {
		artistID := relation.ID
		for location := range relation.DatesLocations {
			cleanedLocation := models.CleanLocationName(location)
			locationIndex[cleanedLocation] = append(locationIndex[cleanedLocation], artistID)
		}
	}

	return locationIndex, nil
}

// SearchLocations performs fast location search using the cached index
func SearchLocations(query string) ([]int, error) {
	locations, err := FetchAllLocations()
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	var matchingArtistIDs []int
	seen := make(map[int]bool)

	for location, artistIDs := range locations {
		if strings.Contains(strings.ToLower(location), queryLower) {
			for _, artistID := range artistIDs {
				if !seen[artistID] {
					matchingArtistIDs = append(matchingArtistIDs, artistID)
					seen[artistID] = true
				}
			}
		}
	}

	return matchingArtistIDs, nil
}

// ClearCache clears the artist cache
func ClearCache() {
	artistCache.mutex.Lock()
	artistCache.artists = nil
	artistCache.lastUpdate = time.Time{}
	artistCache.mutex.Unlock()

	locationCache.mutex.Lock()
	locationCache.locations = make(map[string][]int)
	locationCache.lastUpdate = time.Time{}
	locationCache.mutex.Unlock()
}

// GetCacheStatus returns cache information
func GetCacheStatus() (bool, time.Time) {
	artistCache.mutex.RLock()
	defer artistCache.mutex.RUnlock()
	return !artistCache.lastUpdate.IsZero(), artistCache.lastUpdate
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

// GetLocationSuggestions returns location names that match the query
func GetLocationSuggestions(query string, limit int) ([]string, error) {
	locations, err := FetchAllLocations()
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	var suggestions []string
	seen := make(map[string]bool)

	for location := range locations {
		if strings.Contains(strings.ToLower(location), queryLower) {
			if !seen[location] {
				suggestions = append(suggestions, location)
				seen[location] = true
				if len(suggestions) >= limit {
					break
				}
			}
		}
	}

	return suggestions, nil
}
