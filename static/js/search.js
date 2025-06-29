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

// Clean, Simple Search Implementation
class SearchManager {
  constructor() {
    this.searchInput = document.getElementById('search');
    this.suggestionsBox = document.getElementById('suggestions');
    this.artistList = document.getElementById('artistList');
    this.alertContainer = document.getElementById('alertContainer');
    this.allArtists = [];
    this.filteredArtists = [];
    this.currentIndex = -1;
    this.debounceTimer = null;
    
    this.init();
  }

  init() {
    // Load data on startup
    this.loadArtists();
    
    // Event listeners
    this.searchInput.addEventListener('input', (e) => this.handleInput(e));
    this.searchInput.addEventListener('keydown', (e) => this.handleKeydown(e));
    this.searchInput.addEventListener('focus', () => this.showSuggestions());
    this.searchInput.addEventListener('blur', () => this.hideSuggestions());
    
    // Keyboard shortcut: '/' to focus search (only when search is not focused)
    document.addEventListener('keydown', (e) => {
      if (e.key === '/' && document.activeElement !== this.searchInput) {
        e.preventDefault();
        this.searchInput.focus();
      }
    });
  }

  async loadArtists() {
    try {
      this.showLoading('Loading artists...');
      
      // Check cache first
      if (searchCache.artists.length > 0 && searchCache.lastUpdated) {
        const cacheAge = Date.now() - searchCache.lastUpdated;
        if (cacheAge < 5 * 60 * 1000) { // 5 minutes
          this.allArtists = [...searchCache.artists];
          this.filteredArtists = [...this.allArtists];
          this.renderArtists();
          this.hideLoading();
          return;
        }
      }
      
      const response = await fetch('/api/artists');
      if (!response.ok) throw new Error('Failed to fetch artists');
      
      this.allArtists = await response.json();
      this.filteredArtists = [...this.allArtists];
      
      // Update cache
      searchCache.artists = [...this.allArtists];
      searchCache.lastUpdated = Date.now();
      
      this.renderArtists();
      this.hideLoading();
    } catch (error) {
      this.showError('Unable to fetch data—try again?', () => this.loadArtists());
    }
  }

  handleInput(e) {
    const query = e.target.value.trim();
    
    // Clear previous timer
    clearTimeout(this.debounceTimer);
    
    // Debounce search
    this.debounceTimer = setTimeout(() => {
      this.performSearch(query);
    }, 200);
  }

  performSearch(query) {
    if (!query || query.trim() === '') {
      this.filteredArtists = [...this.allArtists];
      this.renderArtists();
      this.hideSuggestions();
      return;
    }

    const queryLower = query.toLowerCase().trim();
    
    // First, search in basic fields (fast)
    this.filteredArtists = this.allArtists.filter(artist => {
      return (
        artist.name.toLowerCase().includes(queryLower) ||
        artist.members.some(member => member.toLowerCase().includes(queryLower)) ||
        artist.firstAlbum.toLowerCase().includes(queryLower) ||
        artist.creationDate.toString().includes(queryLower)
      );
    });

    this.renderArtists();
    this.showSuggestions(query);
    
    // If no results found, try location search
    if (this.filteredArtists.length === 0) {
      this.performLocationSearch(query);
    }
  }

  async performLocationSearch(query) {
    try {
      this.showLoading('Searching locations...');
      const response = await fetch(`/api/search/locations?q=${encodeURIComponent(query)}`);
      if (!response.ok) throw new Error('Location search failed');
      
      const locationResults = await response.json();
      this.filteredArtists = locationResults;
      this.renderArtists();
      this.hideLoading();
      
      if (locationResults.length > 0) {
        this.showAlert(`Found ${locationResults.length} artist(s) performing in locations matching "${query}"`, 'success');
      }
    } catch (error) {
      console.error('Location search error:', error);
      this.hideLoading();
      this.showAlert('Location search failed. Please try again.', 'warning');
    }
  }

  showSuggestions(query = '') {
    if (!query) {
      this.hideSuggestions();
      return;
    }

    const suggestions = this.getSuggestions(query);
    this.renderSuggestions(suggestions);
    this.suggestionsBox.classList.add('show');
  }

  getSuggestions(query) {
    const queryLower = query.toLowerCase();
    const results = [];

    this.allArtists.forEach(artist => {
      // Check artist name
      if (artist.name.toLowerCase().includes(queryLower)) {
        results.push({ text: artist.name, type: 'artist/band', artistId: artist.id });
      }
      
      // Check members
      artist.members.forEach(member => {
        if (member.toLowerCase().includes(queryLower)) {
          results.push({ text: member, type: 'member', artistId: artist.id });
        }
      });
      
      // Check first album
      if (artist.firstAlbum.toLowerCase().includes(queryLower)) {
        results.push({ text: artist.firstAlbum, type: 'first album', artistId: artist.id });
      }
      
      // Check creation date
      if (artist.creationDate.toString().includes(queryLower)) {
        results.push({ text: artist.creationDate.toString(), type: 'creation date', artistId: artist.id });
      }
    });

    // Add location suggestions (async)
    this.addLocationSuggestions(query, results);

    // Deduplicate and limit to 5
    const unique = [...new Map(results.map(r => [r.text + r.type, r])).values()];
    return unique.slice(0, 5);
  }

  async addLocationSuggestions(query, results) {
    try {
      const response = await fetch(`/api/suggestions/locations?q=${encodeURIComponent(query)}&limit=3`);
      if (response.ok) {
        const locationSuggestions = await response.json();
        locationSuggestions.forEach(location => {
          results.push({ text: location, type: 'location' });
        });
        
        // Re-render suggestions with locations
        this.renderSuggestions(results.slice(0, 5));
      }
    } catch (error) {
      console.error('Error fetching location suggestions:', error);
    }
  }

  renderSuggestions(suggestions) {
    this.suggestionsBox.innerHTML = '';
    
    suggestions.forEach(suggestion => {
      const li = document.createElement('li');
      li.innerHTML = `${suggestion.text} <span class="suggestion-type">— ${suggestion.type}</span>`;
      li.addEventListener('click', () => {
        // Navigate to artist page if we have an artistId
        if (suggestion.artistId) {
          window.location.href = `/artist/${suggestion.artistId}`;
        } else {
          // Fallback to search
          this.searchInput.value = suggestion.text;
          this.performSearch(suggestion.text);
        }
        this.hideSuggestions();
      });
      this.suggestionsBox.appendChild(li);
    });
  }

  hideSuggestions() {
    this.suggestionsBox.classList.remove('show');
  }

  handleKeydown(e) {
    const suggestions = this.suggestionsBox.querySelectorAll('li');
    
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        this.currentIndex = Math.min(this.currentIndex + 1, suggestions.length - 1);
        this.highlightSuggestion(suggestions);
        break;
        
      case 'ArrowUp':
        e.preventDefault();
        this.currentIndex = Math.max(this.currentIndex - 1, -1);
        this.highlightSuggestion(suggestions);
        break;
        
      case 'Enter':
        e.preventDefault();
        if (this.currentIndex >= 0 && suggestions[this.currentIndex]) {
          suggestions[this.currentIndex].click();
        } else {
          this.performSearch(this.searchInput.value);
        }
        break;
        
      case 'Escape':
        this.hideSuggestions();
        this.searchInput.blur();
        break;
    }
  }

  highlightSuggestion(suggestions) {
    suggestions.forEach((suggestion, index) => {
      if (index === this.currentIndex) {
        suggestion.classList.add('highlighted');
      } else {
        suggestion.classList.remove('highlighted');
      }
    });
  }

  renderArtists() {
    if (!this.artistList) return;
    
    this.artistList.innerHTML = '';
    
    if (this.filteredArtists.length === 0) {
      this.artistList.innerHTML = `
        <div class="no-results">
          <p>No artists found matching your search.</p>
          <button class="btn btn-secondary" onclick="searchManager.clearFilters()">Clear Search</button>
        </div>
      `;
      return;
    }
    
    this.filteredArtists.forEach(artist => {
      const artistCard = document.createElement('div');
      artistCard.className = 'artist-card';
      artistCard.innerHTML = `
        <img src="${artist.image}" alt="${artist.name}" class="artist-image">
        <div class="card-body">
          <h5 class="card-title">${artist.name}</h5>
          <p class="card-text">
            <strong>Members:</strong> ${artist.members.join(', ')}<br>
            <strong>Creation Date:</strong> ${artist.creationDate}<br>
            <strong>First Album:</strong> ${artist.firstAlbum || 'Unknown'}
          </p>
          <button class="btn btn-primary" onclick="goToArtist(${artist.id})">View Details</button>
        </div>
      `;
      this.artistList.appendChild(artistCard);
    });
  }

  showMore() {
    // Implementation for showing more results if needed
  }

  clearFilters() {
    this.searchInput.value = '';
    this.filteredArtists = [...this.allArtists];
    this.renderArtists();
    this.hideSuggestions();
  }

  showLoading(message) {
    if (this.alertContainer) {
      this.alertContainer.innerHTML = `
        <div class="alert alert-info">
          <div class="loading">
            <div class="spinner"></div>
            <span>${message}</span>
          </div>
        </div>
      `;
    }
  }

  hideLoading() {
    if (this.alertContainer) {
      this.alertContainer.innerHTML = '';
    }
  }

  showError(message, retryCallback) {
    if (this.alertContainer) {
      this.alertContainer.innerHTML = `
        <div class="alert alert-danger">
          <p>${message}</p>
          ${retryCallback ? `<button class="btn btn-secondary" onclick="searchManager.loadArtists()">Retry</button>` : ''}
        </div>
      `;
    }
  }
}

// Initialize search manager
let searchManager;

document.addEventListener('DOMContentLoaded', () => {
  searchManager = new SearchManager();
});

// Utility functions
function goToArtist(artistId) {
  window.location.href = `/artist/${artistId}`;
}

async function fetchArtists() {
  try {
    const response = await fetch('/api/artists');
    if (!response.ok) throw new Error('Failed to fetch artists');
    return await response.json();
  } catch (error) {
    console.error('Error fetching artists:', error);
    return [];
  }
}

function showAlert(message, type = 'info') {
  const alertContainer = document.getElementById('alertContainer');
  if (alertContainer) {
    alertContainer.innerHTML = `
      <div class="alert alert-${type}">
        <p>${message}</p>
      </div>
    `;
  }
}

async function checkApiAvailability() {
  try {
    const response = await fetch('/api/cache/status');
    if (response.ok) {
      const status = await response.json();
      console.log('Cache status:', status);
    }
  } catch (error) {
    console.error('Error checking cache status:', error);
  }
}