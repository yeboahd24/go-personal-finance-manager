document.addEventListener('DOMContentLoaded', function() {
    // Initialize analytics page
    initializeAnalytics();
});

async function initializeAnalytics() {
    try {
        // Fetch analytics data from the API
        const response = await fetch('/api/analytics', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            throw new Error('Failed to fetch analytics data');
        }

        const data = await response.json();
        
        // TODO: Add visualization code here
        // You might want to use a library like Chart.js to display:
        // - Spending trends
        // - Category breakdown
        // - Monthly comparisons
        // - Income vs. Expenses
        
    } catch (error) {
        console.error('Error initializing analytics:', error);
        showToast('Failed to load analytics data', 'error');
    }
}
