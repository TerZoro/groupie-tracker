package models

import "fmt"

// Artist represents a band or artist from the Groupie Tracker API.
type Artist struct {
	ID           int      `json:"id"`           // Unique identifier for the artist
	Image        string   `json:"image"`        // URL to the artist's image
	Name         string   `json:"name"`         // Name of the artist or band
	Members      []string `json:"members"`      // List of band members
	CreationDate int      `json:"creationDate"` // Year the artist was formed
	FirstAlbum   string   `json:"firstAlbum"`   // Date of the first album
	LocationsURL string   `json:"locations"`    // URL to locations data
	DatesURL     string   `json:"concertDates"` // URL to concert dates data
	RelationsURL string   `json:"relations"`    // URL to relations data
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

// Location represents the data structure for concert locations.
type Location struct {
	ID        int      `json:"id"`        // Unique identifier for the location set
	Locations []string `json:"locations"` // List of concert locations
	DatesURL  string   `json:"dates"`     // URL to corresponding dates
}

// LocationIndex wraps the array of locations from the /locations endpoint.
type LocationIndex struct {
	Index []Location `json:"index"`
}

// Date represents the data structure for concert dates.
type Date struct {
	ID    int      `json:"id"`    // Unique identifier for the date set
	Dates []string `json:"dates"` // List of concert dates
}

// DateIndex wraps the array of dates from the /dates endpoint.
type DateIndex struct {
	Index []Date `json:"index"`
}

// Relation represents the data structure for date-location relations.
type Relation struct {
	ID             int                 `json:"id"`             // Unique identifier for the relation set
	DatesLocations map[string][]string `json:"datesLocations"` // Mapping of locations to dates
}

// RelationIndex wraps the array of relations from the /relation endpoint.
type RelationIndex struct {
	Index []Relation `json:"index"`
}
