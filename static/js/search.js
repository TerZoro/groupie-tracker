// Search suggestion types
const SUGGESTION_TYPES = {
    ARTIST: 'artist/band',
    MEMBER: 'member',
    LOCATION: 'location',
    FIRST_ALBUM: 'first album',
    CREATION_DATE: 'creation date'
};

// Cache for search results
let searchCache = {
    artists: [],
    lastUpdated: null
};

// Search functionality
document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('searchInput');
    const suggestionsContainer = document.getElementById('searchSuggestions');
    const searchButton = document.getElementById('searchButton');
    
    let artists = [];
    
    // Add search button next to input
    addSearchButton();
    
    // Fetch artists data on page load if we're on the home page
    if (window.location.pathname === '/') {
        fetchArtistsFromPage();
    }
    
    // Add input event listener for suggestions
    searchInput.addEventListener('input', handleSearchInput);
    
    // Remove old search button click handler binding here
    // Add enter key handler (use keydown instead of keypress)
    searchInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            performSearch();
        }
    });
    
    // Close suggestions when clicking outside
    document.addEventListener('click', (e) => {
        if (!searchInput.contains(e.target) && !suggestionsContainer.contains(e.target)) {
            suggestionsContainer.style.display = 'none';
        }
    });

    function addSearchButton() {
        const searchContainer = document.querySelector('.search-container');
        if (!searchContainer) return;
        
        const button = document.createElement('button');
        button.id = 'searchButton';
        button.className = 'search-btn';
        button.innerHTML = 'Search';
        button.type = 'button';
        searchContainer.appendChild(button);

        // Attach the click handler here!
        button.addEventListener('click', performSearch);
    }

    function fetchArtistsFromPage() {
        // Extract artist data from the current page
        const artistCards = document.querySelectorAll('.artist-card');
        artists = Array.from(artistCards).map(card => {
            const link = card.querySelector('a[href^="/artist/"]');
            const name = card.querySelector('.card-title');
            const year = card.querySelector('.card-text');
            
            if (link && name) {
                const id = link.getAttribute('href').split('/')[2];
                return {
                    id: parseInt(id),
                    name: name.textContent,
                    creationDate: year ? year.textContent.replace('Formed: ', '') : '',
                    image: card.querySelector('img') ? card.querySelector('img').src : ''
                };
            }
            return null;
        }).filter(artist => artist !== null);
    }
    
    function handleSearchInput(e) {
        const query = e.target.value.trim();
        if (query.length === 0) {
            suggestionsContainer.style.display = 'none';
            return;
        }
        if (query.length >= 2) {
            showSuggestions(query);
        }
    }
    
    function showSuggestions(query) {
        const suggestions = generateSuggestions(query);
        
        if (suggestions.length === 0) {
            suggestionsContainer.style.display = 'none';
            return;
        }
        
        const suggestionsHTML = suggestions.map(suggestion => `
            <div class="suggestion-item" onclick="goToArtist(${suggestion.id})">
                <span class="suggestion-text">${suggestion.text}</span>
                <span class="suggestion-type">${suggestion.type}</span>
            </div>
        `).join('');
        
        suggestionsContainer.innerHTML = suggestionsHTML;
        suggestionsContainer.style.display = 'block';
    }
    
    function generateSuggestions(query) {
        query = query.toLowerCase();
        const suggestions = [];
        
        artists.forEach(artist => {
            // Check artist name
            if (artist.name.toLowerCase().includes(query)) {
                suggestions.push({
                    id: artist.id,
                    text: artist.name,
                    type: 'artist'
                });
            }
            
            // Check creation date
            if (artist.creationDate.includes(query)) {
                suggestions.push({
                    id: artist.id,
                    text: `${artist.name} (${artist.creationDate})`,
                    type: 'year'
                });
            }
        });
        
        return suggestions.slice(0, 5); // Limit to 5 suggestions
    }
    
    function performSearch() {
        const query = searchInput.value.trim();
        if (query.length === 0) return;
        
        // Use our existing /search endpoint
        window.location.href = `/search?q=${encodeURIComponent(query)}`;
    }
});

// Global function for suggestion clicks
function goToArtist(artistId) {
    window.location.href = `/artist/${artistId}`;
}

// Fetch artists data
async function fetchArtists() {
    try {
        const response = await fetch('/api/artists');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        if (!Array.isArray(data)) {
            throw new Error('Invalid data format received');
        }
        searchCache.artists = data;
        searchCache.lastUpdated = Date.now();
    } catch (error) {
        console.error('Error fetching artists:', error);
        showAlert('Failed to load artists data. Please refresh the page.', 'danger');
        // Retry after 5 seconds
        setTimeout(fetchArtists, 5000);
    }
}

// Show alert message
function showAlert(message, type = 'info') {
    const alertContainer = document.getElementById('alertContainer');
    const alert = document.createElement('div');
    alert.className = `alert alert-${type} alert-dismissible fade show`;
    alert.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    `;
    alertContainer.appendChild(alert);
    setTimeout(() => alert.remove(), 5000);
}

// Add a function to check if the API is available
async function checkApiAvailability() {
    try {
        const response = await fetch('/api/artists');
        return response.ok;
    } catch (error) {
        return false;
    }
}