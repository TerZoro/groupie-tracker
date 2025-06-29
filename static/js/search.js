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
      const response = await fetch('/api/artists');
      if (!response.ok) throw new Error('Failed to fetch artists');
      
      this.allArtists = await response.json();
      this.filteredArtists = [...this.allArtists];
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

    // Deduplicate and limit to 5
    const unique = [...new Map(results.map(r => [r.text + r.type, r])).values()];
    return unique.slice(0, 5);
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
    suggestions.forEach((li, index) => {
      li.style.backgroundColor = index === this.currentIndex ? 'var(--light)' : '';
    });
  }

  renderArtists() {
    if (this.filteredArtists.length === 0) {
      this.artistList.innerHTML = `
        <div style="display: flex; justify-content: center; align-items: center; min-height: 400px;">
          <div class="alert alert-warning" style="margin: 0;">
            No artists found matching your search.
          </div>
        </div>
      `;
      return;
    }

    // Show all results (removed the 10-result limit)
    let html = '';
    this.filteredArtists.forEach(artist => {
      html += `
        <div class="artist-card">
          <a href="/artist/${artist.id}">
            <img src="${artist.image}" class="artist-image" alt="${artist.name}">
          </a>
          <div class="card-body">
            <h5 class="card-title">${artist.name}</h5>
            <p class="card-text">Formed: ${artist.creationDate}</p>
            <a href="/artist/${artist.id}" class="btn btn-primary">View Details</a>
          </div>
        </div>
      `;
    });

    this.artistList.innerHTML = html;
  }

  showMore() {
    // Removed - not needed anymore
  }

  clearFilters() {
    this.searchInput.value = '';
    this.filteredArtists = [...this.allArtists];
    this.renderArtists();
    this.hideSuggestions();
  }

  showLoading(message) {
    this.artistList.innerHTML = `
      <div class="loading" style="display: flex; justify-content: center; align-items: center; min-height: 400px;">
        <div style="text-align: center;">
          <div class="spinner"></div>
          <p>${message}</p>
        </div>
      </div>
    `;
  }

  hideLoading() {
    // Loading is hidden when renderArtists is called
  }

  showError(message, retryCallback) {
    this.alertContainer.innerHTML = `
      <div class="alert alert-danger" style="margin: var(--spacing-md) auto; text-align: center;">
        ${message}
        ${retryCallback ? '<button class="btn btn-danger" onclick="this.parentElement.remove(); searchManager.loadArtists()">Retry</button>' : ''}
      </div>
    `;
  }
}

// Initialize search when DOM is loaded
let searchManager;
document.addEventListener('DOMContentLoaded', () => {
  searchManager = new SearchManager();
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