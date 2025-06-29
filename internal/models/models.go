package models

import (
	"fmt"
	"strings"
)

// Artist represents a musical artist or band
type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"concertDates"`
	Relations    string   `json:"relations"`
	// Additional fields for search
	LocationList []string `json:"-"` // Will be populated from relations
}

// Validate ensures the artist data is valid
func (a Artist) Validate() error {
	if a.ID == 0 {
		return fmt.Errorf("artist ID cannot be zero")
	}
	if a.Name == "" {
		return fmt.Errorf("artist name cannot be empty")
	}
	return nil
}

// GetSearchableText returns all searchable text for this artist
func (a Artist) GetSearchableText() string {
	texts := []string{
		a.Name,
		strings.Join(a.Members, " "),
		a.FirstAlbum,
		fmt.Sprintf("%d", a.CreationDate),
		strings.Join(a.LocationList, " "),
	}
	return strings.ToLower(strings.Join(texts, " "))
}

// CleanLocationName removes slashes and formats location names properly
func CleanLocationName(location string) string {
	// Remove slashes and replace with spaces
	cleaned := strings.ReplaceAll(location, "-", " ")
	cleaned = strings.ReplaceAll(cleaned, "_", " ")

	// Capitalize words
	words := strings.Fields(cleaned)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

// Location represents a concert location
type Location struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	Dates     string   `json:"dates"`
}

// LocationIndex wraps the array of locations from the /locations endpoint.
type LocationIndex struct {
	Index []Location `json:"index"`
}

// Date represents concert dates
type Date struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

// DateIndex wraps the array of dates from the /dates endpoint.
type DateIndex struct {
	Index []Date `json:"index"`
}

// Relation represents relationships between artists and their data
type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// RelationIndex wraps the array of relations from the /relation endpoint.
type RelationIndex struct {
	Index []Relation `json:"index"`
}
