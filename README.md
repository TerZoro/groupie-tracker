# Groupie Tracker

A secure web application for tracking music artists and bands, built with Go and following Schneiderman's 8 Golden Rules for UI design.

## Security Features

### Static File Security
- **Directory Traversal Protection**: Blocks attempts to access files outside the static directory
- **File Type Whitelist**: Only allows specific file extensions (.css, .js, .png, .jpg, .jpeg, .gif, .svg, .ico)
- **Path Validation**: Ensures all requests are within the allowed directory structure

### Template Security
- Templates are stored in `internal/templates/` and cannot be accessed directly
- Server-side rendering prevents client-side template injection
- Input validation and sanitization on all user inputs

### Startup Safety Checks
- Validates required directories exist and are not empty
- Prevents server startup if critical files are missing
- Clear error messages for configuration issues

## API Endpoints

### GET /api/artists
Returns all artists data as JSON for frontend search functionality.

**Response:**
```json
[
  {
    "id": 1,
    "name": "Artist Name",
    "image": "image_url",
    "members": ["Member 1", "Member 2"],
    "creationDate": 1990,
    "firstAlbum": "Album Name"
  }
]
```

**Headers:**
- `Content-Type: application/json`
- `Cache-Control: public, max-age=300` (5 minutes cache)

## UI/UX Features (Schneiderman's 8 Golden Rules)

### 1. Consistency
- Unified design tokens in `app.css` (colors, typography, spacing)
- Consistent button styles and interactions
- Standardized card layouts and hover effects

### 2. Shortcuts
- Press `/` to focus the search bar
- Arrow keys to navigate suggestions
- Enter to select, Escape to close

### 3. Informative Feedback
- Loading spinner with descriptive messages
- Clear error states with retry options
- Real-time search suggestions

### 4. Dialog Closure
- Suggestions auto-hide when clicking outside
- Clear visual feedback for all interactions

### 5. Simple Error Handling
- User-friendly error messages
- Retry buttons for failed operations
- Graceful degradation

### 6. Easy Reversal
- "Clear Filters" button resets all search
- Undo functionality for user actions

### 7. User Control
- No auto-redirects
- Explicit confirmation for actions
- User-initiated navigation

### 8. Reduce Memory Load
- Show only top 10 results by default
- "Show More" button for additional results
- Clean, uncluttered interface

## Search Features

### Type-Tagged Suggestions
- Artist/band names
- Band members
- First album titles
- Creation dates

### Case-Insensitive Search
- Searches across all artist fields
- Real-time filtering with debouncing
- Keyboard navigation support

## File Structure

```
groupie-tracker/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   └── api.go           # External API integration
│   ├── handlers/
│   │   └── handlers.go      # HTTP handlers with security
│   ├── models/
│   │   └── models.go        # Data structures
│   └── templates/
│       ├── index.html       # Main page template
│       ├── artist.html      # Artist detail template
│       └── error.html       # Error page template
├── static/
│   ├── css/
│   │   └── app.css          # Unified styles with design tokens
│   ├── js/
│   │   └── search.js        # Clean search implementation
│   └── images/
│       └── icon.png         # App icon
├── security_test.go         # Security validation tests
└── README.md               # This file
```

## Running the Application

```bash
go run cmd/main.go
```

The server will start on port 8080 (or the PORT environment variable).

## Testing

Run the security tests:
```bash
go test -v
```

## Security Audit Checklist

- [x] Directory traversal protection
- [x] File type whitelist
- [x] Template injection prevention
- [x] Input validation
- [x] Error handling
- [x] Startup safety checks
- [x] Content-Type headers
- [x] Path validation
- [x] Access control

## Performance Features

- Static file caching (5 minutes for API responses)
- Debounced search (200ms delay)
- Lazy loading of results (10 at a time)
- Optimized CSS with design tokens
- Minimal JavaScript footprint

