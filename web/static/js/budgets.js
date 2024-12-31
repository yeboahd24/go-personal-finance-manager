document.addEventListener('DOMContentLoaded', function() {
    // Initialize budgets page
    initializeBudgets();
});

async function initializeBudgets() {
    try {
        // Fetch budgets data from the API
        const response = await fetch('/api/budgets', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            throw new Error('Failed to fetch budgets data');
        }

        const data = await response.json();
        
        // TODO: Add budget management UI code here
        // Features to implement:
        // - Create new budgets
        // - Edit existing budgets
        // - Delete budgets
        // - View budget progress
        // - Category-wise budget allocation
        
    } catch (error) {
        console.error('Error initializing budgets:', error);
        showToast('Failed to load budgets data', 'error');
    }
}

async function createBudget(budgetData) {
    try {
        const response = await fetch('/api/budgets', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(budgetData)
        });

        if (!response.ok) {
            throw new Error('Failed to create budget');
        }

        showToast('Budget created successfully');
        initializeBudgets(); // Refresh the budgets list
    } catch (error) {
        console.error('Error creating budget:', error);
        showToast('Failed to create budget', 'error');
    }
}
