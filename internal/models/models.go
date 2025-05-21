package models

import "fmt"

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
