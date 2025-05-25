# Groupie Tracker


**Groupie Tracker** is a web application that lets users explore music artists and their concert schedules. It fetches data from an external API and presents it in a clean, responsive interface built with Go, HTML, CSS, JavaScript, and Bootstrap. Key features include a searchable artist list, sorted concert schedules.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Project Structure](#project-structure)

## Features
- **Artist Cards**: Displays artists with centered group names, creation dates, and "View Details" buttons on the main page.
- **Search Bar**: Located in the navbar, allows filtering artists by name (e.g., "Scorpions").
- **Concert Schedules**: Sorted by most recent date on artist detail pages (e.g., Scorpions' "Auckland, New Zealand" on 2020-02-27 appears first).
- **Refresh Data**: Button clears cached API data and reloads the page.
- **Caching**: Stores API data for 10 minutes to reduce load times.


## Installation

### Prerequisites
- [Go](https://golang.org/dl/) (1.20 or later)
- [Git](https://git-scm.com/downloads)

### Steps
1. **Clone the Repository**:
   ```bash
   git clone https://01.tomorrow-school.ai/git/bzhaksyba/groupie-tracker
   cd groupie-tracker
   ```

2. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

3. **Build and Run**:
   ```bash
   go build
   go run .
   ```
   The server starts at `http://localhost:8080`.

## Usage
1. **Browse Artists**:
   - Visit `http://localhost:8080/` to see a grid of artist cards.
   - Each card shows the artist’s name, creation date, and image, with a centered "View Details" button.

2. **Search Artists**:
   - Use the search bar in the navbar to filter artists (e.g., type "Scorpions" and press Enter).
   - Clear the search to reset the list.

3. **View Artist Details**:
   - Click "View Details" to visit `/artist/<id>` (e.g., `/artist/2` for Scorpions).
   - See artist details (members, creation date, first album) and a sorted concert schedule.

4. **Refresh Data**:
   - Click "Refresh Data" in the navbar to clear the cache and reload fresh API data.

## Project Structure
```
groupie-tracker/
├── internal/
│   ├── api/            # API client for fetching artist data
│   ├── handlers/       # HTTP handlers and tests
│   ├── models/         # Data models (Artist, Relation, etc.)
│   ├── templates/      # HTML templates (index.html, artist.html)
│   └── cache/          # Caching logic
├── static/
│   ├── css/            # Styles (styles.css)
│   ├── js/             # Scripts (scripts.js)
│   └── images/         # Favicon (icon.png)
├── go.mod              # Go module file
├── cmd/server
│   └── main.go         # Entry point
└── README.md           # This file
```

