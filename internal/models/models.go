package models

// Artist represents the data structure for an artist from the API
type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	LocationsURL string   `json:"locations"`    // URL to locations data
	DatesURL     string   `json:"concertDates"` // URL to concert dates data
	RelationsURL string   `json:"relations"`    // URL to relations data
}

// Location represents the data structure for concert locations
type Location struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	DatesURL  string   `json:"dates"` // URL to corresponding dates
}

// LocationIndex wraps the array of locations from the /locations endpoint
type LocationIndex struct {
	Index []Location `json:"index"`
}

// Date represents the data structure for concert dates
type Date struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

// DateIndex wraps the array of dates from the /dates endpoint
type DateIndex struct {
	Index []Date `json:"index"`
}

// Relation represents the data structure for date-location relations
type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// RelationIndex wraps the array of relations from the /relation endpoint
type RelationIndex struct {
	Index []Relation `json:"index"`
}
