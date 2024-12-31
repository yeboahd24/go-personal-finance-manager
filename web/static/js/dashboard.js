// Create a flag to track if we're already redirecting
let isRedirecting = false;

// Function to handle auth redirect
function redirectToLogin() {
    if (isRedirecting) return;
    isRedirecting = true;
    window.location.replace('/login');
}

document.addEventListener('DOMContentLoaded', function() {
    // Check if we're on the dashboard page
    if (window.location.pathname !== '/' && window.location.pathname !== '/dashboard') {
        return;
    }

    // Load dashboard data
    loadDashboardData();
});

async function loadDashboardData() {
    try {
        // Load metrics first
        await Promise.all([
            loadMetric('net-worth', '/api/metrics?type=net_worth'),
            loadMetric('monthly-income', '/api/analytics/income?period=month'),
            loadMetric('monthly-expenses', '/api/analytics/expenses?period=month'),
            loadMetric('monthly-savings', '/api/analytics/savings?period=month')
        ]);

        // Then load transactions
        await loadRecentTransactions();
    } catch (error) {
        console.error('Error loading dashboard data:', error);
        if (error.status === 401) {
            window.location.href = '/login';
        }
    }
}

async function loadMetric(elementId, url) {
    const element = document.querySelector(`[data-metric="${elementId}"]`);
    if (!element) return;

    try {
        const response = await fetch(url, {
            credentials: 'include',
            headers: {
                'Accept': 'application/json'
            }
        });

        if (!response.ok) {
            const data = await response.json();
            if (response.status === 401) {
                throw { status: 401, message: data.error };
            }
            element.textContent = '$0.00';
            return;
        }

        const data = await response.json();
        const value = data.value || data.amount || 0;
        element.textContent = formatCurrency(value);
    } catch (error) {
        console.error(`Error loading ${elementId}:`, error);
        element.textContent = '$0.00';
        if (error.status === 401) {
            throw error;
        }
    }
}

async function loadRecentTransactions() {
    const transactionsBody = document.getElementById('recent-transactions');
    if (!transactionsBody) return;

    try {
        const response = await fetch('/api/transactions/recent', {
            credentials: 'include',
            headers: {
                'Accept': 'application/json'
            }
        });

        if (!response.ok) {
            const data = await response.json();
            if (response.status === 401) {
                throw { status: 401, message: data.error };
            }
            transactionsBody.innerHTML = `
                <tr>
                    <td colspan="4" class="px-6 py-4 text-sm text-gray-500 text-center">Failed to load transactions</td>
                </tr>`;
            return;
        }

        const data = await response.json();
        if (!data.transactions || data.transactions.length === 0) {
            transactionsBody.innerHTML = `
                <tr>
                    <td colspan="4" class="px-6 py-4 text-sm text-gray-500 text-center">No recent transactions</td>
                </tr>`;
            return;
        }

        transactionsBody.innerHTML = data.transactions.map(transaction => `
            <tr>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    ${new Date(transaction.date).toLocaleDateString()}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    ${transaction.description}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    ${transaction.category || 'Uncategorized'}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-right ${transaction.type === 'debit' ? 'text-red-600' : 'text-green-600'}">
                    ${formatCurrency(transaction.amount)}
                </td>
            </tr>
        `).join('');
    } catch (error) {
        console.error('Error loading transactions:', error);
        if (error.status === 401) {
            throw error;
        }
        transactionsBody.innerHTML = `
            <tr>
                <td colspan="4" class="px-6 py-4 text-sm text-gray-500 text-center">Error loading transactions</td>
            </tr>`;
    }
}

function formatCurrency(value) {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD',
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
    }).format(value);
}
