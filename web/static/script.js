document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('search-input');
    const searchButton = document.getElementById('search-button');
    const resultsContainer = document.getElementById('results-container');

    const performSearch = async () => {
        const query = searchInput.value.trim();
        if (!query) {
            resultsContainer.innerHTML = '<p>Please enter a search query.</p>';
            return;
        }

        resultsContainer.innerHTML = '<p>Searching...</p>';

        try {
            const response = await fetch(`/api/v1/search?q=${encodeURIComponent(query)}`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            displayResults(data.results);
        } catch (error) {
            resultsContainer.innerHTML = `<p>Error performing search: ${error.message}</p>`;
            console.error('Search failed:', error);
        }
    };

    const displayResults = (results) => {
        if (!results || results.length === 0) {
            resultsContainer.innerHTML = '<p>No results found.</p>';
            return;
        }

        resultsContainer.innerHTML = results.map(result => `
            <div class="result">
                <h3><a href="${result.url}" target="_blank" rel="noopener noreferrer">${result.title}</a></h3>
                <p><a href="${result.url}" target="_blank" rel="noopener noreferrer">${result.url}</a></p>
            </div>
        `).join('');
    };

    searchButton.addEventListener('click', performSearch);
    searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            performSearch();
        }
    });
});