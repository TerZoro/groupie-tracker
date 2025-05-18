package cache

import (
	"groupie/internal/models"
	"sync"
	"time"
)

// Cache stores API data with thread-safe access
type Cache struct {
	mu           sync.RWMutex
	artists      []models.Artist
	locations    models.LocationIndex
	dates        models.DateIndex
	relations    models.RelationIndex
	lastUpdated  time.Time
	artistsMap   map[int]models.Artist
	locationsMap map[int]models.Location
	datesMap     map[int]models.Date
	relationsMap map[int]models.Relation
}

// Global cache instance
var globalCache = &Cache{
	artistsMap:   make(map[int]models.Artist),
	locationsMap: make(map[int]models.Location),
	datesMap:     make(map[int]models.Date),
	relationsMap: make(map[int]models.Relation),
}

// SetArtists updates the cached artists
func SetArtists(artists []models.Artist) {
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()
	globalCache.artists = artists
	globalCache.artistsMap = make(map[int]models.Artist)
	for _, artist := range artists {
		globalCache.artistsMap[artist.ID] = artist
	}
	globalCache.lastUpdated = time.Now()
}

// GetArtists returns the cached artists
func GetArtists() ([]models.Artist, time.Time) {
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
	return globalCache.artists, globalCache.lastUpdated
}

// GetArtistByID returns a single artist by ID
func GetArtistByID(id int) (models.Artist, bool, time.Time) {
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
	artist, exists := globalCache.artistsMap[id]
	return artist, exists, globalCache.lastUpdated
}

// SetLocations updates the cached locations
func SetLocations(locations models.LocationIndex) {
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()
	globalCache.locations = locations
	globalCache.locationsMap = make(map[int]models.Location)
	for _, loc := range locations.Index {
		globalCache.locationsMap[loc.ID] = loc
	}
	globalCache.lastUpdated = time.Now()
}

// GetLocations returns the cached locations
func GetLocations() (models.LocationIndex, time.Time) {
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
	return globalCache.locations, globalCache.lastUpdated
}

// GetLocationByID returns a single location by artist ID
func GetLocationByID(id int) (models.Location, bool, time.Time) {
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
	loc, exists := globalCache.locationsMap[id]
	return loc, exists, globalCache.lastUpdated
}

// UpdateLocation updates the cache for a single artist's location
func UpdateLocation(id int, location models.Location) {
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()
	globalCache.locationsMap[id] = location
	// Update locations.Index to keep it consistent
	found := false
	for i, loc := range globalCache.locations.Index {
		if loc.ID == id {
			globalCache.locations.Index[i] = location
			found = true
			break
		}
	}
	if !found && location.ID != 0 {
		globalCache.locations.Index = append(globalCache.locations.Index, location)
	}
	globalCache.lastUpdated = time.Now()
}

// SetDates updates the cached dates
func SetDates(dates models.DateIndex) {
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()
	globalCache.dates = dates
	globalCache.datesMap = make(map[int]models.Date)
	for _, date := range dates.Index {
		globalCache.datesMap[date.ID] = date
	}
	globalCache.lastUpdated = time.Now()
}

// GetDates returns the cached dates
func GetDates() (models.DateIndex, time.Time) {
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
	return globalCache.dates, globalCache.lastUpdated
}

// GetDateByID returns a single date by artist ID
func GetDateByID(id int) (models.Date, bool, time.Time) {
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
	date, exists := globalCache.datesMap[id]
	return date, exists, globalCache.lastUpdated
}

// UpdateDate updates the cache for a single artist's date
func UpdateDate(id int, date models.Date) {
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()
	globalCache.datesMap[id] = date
	// Update dates.Index to keep it consistent
	found := false
	for i, d := range globalCache.dates.Index {
		if d.ID == id {
			globalCache.dates.Index[i] = date
			found = true
			break
		}
	}
	if !found && date.ID != 0 {
		globalCache.dates.Index = append(globalCache.dates.Index, date)
	}
	globalCache.lastUpdated = time.Now()
}

// SetRelations updates the cached relations
func SetRelations(relations models.RelationIndex) {
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()
	globalCache.relations = relations
	globalCache.relationsMap = make(map[int]models.Relation)
	for _, rel := range relations.Index {
		globalCache.relationsMap[rel.ID] = rel
	}
	globalCache.lastUpdated = time.Now()
}

// GetRelations returns the cached relations
func GetRelations() (models.RelationIndex, time.Time) {
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
	return globalCache.relations, globalCache.lastUpdated
}

// GetRelationByID returns a single relation by artist ID
func GetRelationByID(id int) (models.Relation, bool, time.Time) {
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
	rel, exists := globalCache.relationsMap[id]
	return rel, exists, globalCache.lastUpdated
}

// UpdateRelation updates the cache for a single artist's relation
func UpdateRelation(id int, relation models.Relation) {
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()
	globalCache.relationsMap[id] = relation
	// Update relations.Index to keep it consistent
	found := false
	for i, r := range globalCache.relations.Index {
		if r.ID == id {
			globalCache.relations.Index[i] = relation
			found = true
			break
		}
	}
	if !found && relation.ID != 0 {
		globalCache.relations.Index = append(globalCache.relations.Index, relation)
	}
	globalCache.lastUpdated = time.Now()
}

// Clear clears the entire cache
func Clear() {
	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()
	globalCache.artists = nil
	globalCache.locations = models.LocationIndex{}
	globalCache.dates = models.DateIndex{}
	globalCache.relations = models.RelationIndex{}
	globalCache.artistsMap = make(map[int]models.Artist)
	globalCache.locationsMap = make(map[int]models.Location)
	globalCache.datesMap = make(map[int]models.Date)
	globalCache.relationsMap = make(map[int]models.Relation)
	globalCache.lastUpdated = time.Time{}
}
