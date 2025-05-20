let debounceTimer;

async function searchArtists() {
    const query = document.getElementById('searchInput').value.trim();
    if (!query) {
        showAlert('Please enter a search query', 'warning');
        return;
    }
    if (query.length > 100) {
        showAlert('Search query must be under 100 characters', 'warning');
        return;
    }

    try {
        const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`, {
            method: 'GET',
            headers: { 'Content-Type': 'application/json' }
        });

        if (!response.ok) {
            throw new Error(`Search failed with status ${response.status}`);
        }

        const artists = await response.json();
        const artistList = document.getElementById('artistList');
        artistList.innerHTML = '';

        if (artists.length === 0) {
            artistList.innerHTML = `
                <div class="col-12">
                    <div class="alert alert-warning text-center" role="alert">
                        No artists found for "${query}".
                    </div>
                </div>
            `;
            return;
        }

        artists.forEach(artist => {
            const card = `
                <div class="col-md-4 col-sm-6 mb-4">
                    <div class="card h-100">
                        <img src="${artist.Image}" class="card-img-top" alt="${artist.Name}" style="height: 200px; object-fit: cover;">
                        <div class="card-body">
                            <h5 class="card-title">${artist.Name}</h5>
                            <p class="card-text">Formed: ${artist.CreationDate}</p>
                            <a href="/artist/${artist.ID}" class="btn btn-primary">View Details</a>
                        </div>
                    </div>
                </div>
            `;
            artistList.innerHTML += card;
        });
    } catch (error) {
        console.error('Search error:', error);
        showAlert('Failed to search artists. Please try again.', 'danger');
    }
}

async function refreshCache() {
    try {
        const response = await fetch('/api/refresh-cache', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' }
        });

        if (!response.ok) {
            throw new Error('Cache refresh failed');
        }

        showAlert('Data refreshed successfully', 'success');
        // Reload the page to show updated data
        window.location.reload();
    } catch (error) {
        console.error('Cache refresh error:', error);
        showAlert('Failed to refresh data. Please try again.', 'danger');
    }
}

function showAlert(message, type) {
    const artistList = document.getElementById('artistList');
    artistList.innerHTML = `
        <div class="col-12">
            <div class="alert alert-${type} text-center" role="alert">
                ${message}
            </div>
        </div>
    `;
}

// Debounced search
function debounceSearch() {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(searchArtists, 300);
}

// Add Enter key and input event listeners
document.getElementById('searchInput').addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
        searchArtists();
    }
});

document.getElementById('searchInput').addEventListener('input', debounceSearch);