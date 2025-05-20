package cache

import (
	"groupie/internal/models"
	"sync"
	"time"
)

// Cache stores API data with thread-safe access
type Cache struct {
	mu                   sync.RWMutex
	artists              []models.Artist
	locations            models.LocationIndex
	dates                models.DateIndex
	relations            models.RelationIndex
	artistsLastUpdated   time.Time
	locationsLastUpdated time.Time
	datesLastUpdated     time.Time
	relationsLastUpdated time.Time
	artistsMap           map[int]models.Artist
	locationsMap         map[int]models.Location
	datesMap             map[int]models.Date
	relationsMap         map[int]models.Relation
}

var cacheInstance *Cache
var once sync.Once

// GetCache returns the singleton cache instance
func GetCache() *Cache {
	once.Do(func() {
		cacheInstance = &Cache{
			artistsMap:   make(map[int]models.Artist),
			locationsMap: make(map[int]models.Location),
			datesMap:     make(map[int]models.Date),
			relationsMap: make(map[int]models.Relation),
		}
	})
	return cacheInstance
}

// SetArtists updates the cached artists
func SetArtists(artists []models.Artist) {
	cache := GetCache()
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.artists = artists
	cache.artistsMap = make(map[int]models.Artist)
	for _, artist := range artists {
		if artist.ID == 0 {
			continue
		}
		cache.artistsMap[artist.ID] = artist
	}
	cache.artistsLastUpdated = time.Now()
}

// GetArtists returns the cached artists
func GetArtists() ([]models.Artist, time.Time) {
	cache := GetCache()
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	return cache.artists, cache.artistsLastUpdated
}

// GetArtistByID returns a single artist by ID
func GetArtistByID(id int) (models.Artist, bool, time.Time) {
	cache := GetCache()
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	artist, exists := cache.artistsMap[id]
	return artist, exists, cache.artistsLastUpdated
}

// UpdateArtist updates the cache for a single artist
func UpdateArtist(id int, artist models.Artist) {
	if id == 0 || artist.ID == 0 {
		return
	}
	cache := GetCache()
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.artistsMap[id] = artist
	// Update artists slice to keep it consistent
	found := false
	for i, a := range cache.artists {
		if a.ID == id {
			cache.artists[i] = artist
			found = true
			break
		}
	}
	if !found && artist.ID != 0 {
		cache.artists = append(cache.artists, artist)
	}
	cache.artistsLastUpdated = time.Now()
}

// SetLocations updates the cached locations
func SetLocations(locations models.LocationIndex) {
	cache := GetCache()
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.locations = locations
	cache.locationsMap = make(map[int]models.Location)
	for _, loc := range locations.Index {
		if loc.ID == 0 {
			continue
		}
		cache.locationsMap[loc.ID] = loc
	}
	cache.locationsLastUpdated = time.Now()
}

// GetLocations returns the cached locations
func GetLocations() (models.LocationIndex, time.Time) {
	cache := GetCache()
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	return cache.locations, cache.locationsLastUpdated
}

// GetLocationByID returns a single location by artist ID
func GetLocationByID(id int) (models.Location, bool, time.Time) {
	cache := GetCache()
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	loc, exists := cache.locationsMap[id]
	return loc, exists, cache.locationsLastUpdated
}

// UpdateLocation updates the cache for a single artist's location
func UpdateLocation(id int, location models.Location) {
	if id == 0 || location.ID == 0 {
		return
	}
	cache := GetCache()
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.locationsMap[id] = location
	// Update locations.Index to keep it consistent
	found := false
	for i, loc := range cache.locations.Index {
		if loc.ID == id {
			cache.locations.Index[i] = location
			found = true
			break
		}
	}
	if !found && location.ID != 0 {
		cache.locations.Index = append(cache.locations.Index, location)
	}
	cache.locationsLastUpdated = time.Now()
}

// SetDates updates the cached dates
func SetDates(dates models.DateIndex) {
	cache := GetCache()
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.dates = dates
	cache.datesMap = make(map[int]models.Date)
	for _, date := range dates.Index {
		if date.ID == 0 {
			continue
		}
		cache.datesMap[date.ID] = date
	}
	cache.datesLastUpdated = time.Now()
}

// GetDates returns the cached dates
func GetDates() (models.DateIndex, time.Time) {
	cache := GetCache()
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	return cache.dates, cache.datesLastUpdated
}

// GetDateByID returns a single date by artist ID
func GetDateByID(id int) (models.Date, bool, time.Time) {
	cache := GetCache()
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	date, exists := cache.datesMap[id]
	return date, exists, cache.datesLastUpdated
}

// UpdateDate updates the cache for a single artist's date
func UpdateDate(id int, date models.Date) {
	if id == 0 || date.ID == 0 {
		return
	}
	cache := GetCache()
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.datesMap[id] = date
	// Update dates.Index to keep it consistent
	found := false
	for i, d := range cache.dates.Index {
		if d.ID == id {
			cache.dates.Index[i] = date
			found = true
			break
		}
	}
	if !found && date.ID != 0 {
		cache.dates.Index = append(cache.dates.Index, date)
	}
	cache.datesLastUpdated = time.Now()
}

// SetRelations updates the cached relations
func SetRelations(relations models.RelationIndex) {
	cache := GetCache()
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.relations = relations
	cache.relationsMap = make(map[int]models.Relation)
	for _, rel := range relations.Index {
		if rel.ID == 0 {
			continue
		}
		cache.relationsMap[rel.ID] = rel
	}
	cache.relationsLastUpdated = time.Now()
}

// GetRelations returns the cached relations
func GetRelations() (models.RelationIndex, time.Time) {
	cache := GetCache()
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	return cache.relations, cache.relationsLastUpdated
}

// GetRelationByID returns a single relation by artist ID
func GetRelationByID(id int) (models.Relation, bool, time.Time) {
	cache := GetCache()
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	rel, exists := cache.relationsMap[id]
	return rel, exists, cache.relationsLastUpdated
}

// UpdateRelation updates the cache for a single artist's relation
func UpdateRelation(id int, relation models.Relation) {
	if id == 0 || relation.ID == 0 {
		return
	}
	cache := GetCache()
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.relationsMap[id] = relation
	// Update relations.Index to keep it consistent
	found := false
	for i, r := range cache.relations.Index {
		if r.ID == id {
			cache.relations.Index[i] = relation
			found = true
			break
		}
	}
	if !found && relation.ID != 0 {
		cache.relations.Index = append(cache.relations.Index, relation)
	}
	cache.relationsLastUpdated = time.Now()
}

// Clear clears the entire cache
func Clear() {
	cache := GetCache()
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.artists = nil
	cache.locations = models.LocationIndex{}
	cache.dates = models.DateIndex{}
	cache.relations = models.RelationIndex{}
	cache.artistsMap = make(map[int]models.Artist)
	cache.locationsMap = make(map[int]models.Location)
	cache.datesMap = make(map[int]models.Date)
	cache.relationsMap = make(map[int]models.Relation)
	cache.artistsLastUpdated = time.Time{}
	cache.locationsLastUpdated = time.Time{}
	cache.datesLastUpdated = time.Time{}
	cache.relationsLastUpdated = time.Time{}
}
