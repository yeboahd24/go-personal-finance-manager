// Format currency values
function formatCurrency(amount) {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD'
    }).format(amount);
}

// Format dates
function formatDate(dateStr) {
    return new Date(dateStr).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
    });
}

// Show toast notification
function showToast(message, type = 'success') {
    const toast = document.createElement('div');
    const bgColor = type === 'success' ? 'bg-green-500' : 'bg-red-500';
    toast.className = `${bgColor} text-white px-6 py-4 rounded-lg mb-2 transition-opacity duration-500`;
    toast.textContent = message;
    
    const container = document.getElementById('toast-container');
    if (container) {
        container.appendChild(toast);
        setTimeout(() => {
            toast.style.opacity = '0';
            setTimeout(() => container.removeChild(toast), 500);
        }, 3000);
    }
}

// Get auth token from cookie
function getAuthToken() {
    const cookies = document.cookie.split(';');
    for (let cookie of cookies) {
        const [name, value] = cookie.trim().split('=');
        if (name === 'authToken') {
            return value;
        }
    }
    return null;
}

// Check if user is authenticated
function isAuthenticated() {
    return !!getAuthToken();
}

// Handle logout
function logout() {
    // Remove auth cookie
    document.cookie = 'authToken=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; secure; samesite=strict';
    window.location.href = '/login';
}

// Initialize app and HTMX
document.addEventListener('DOMContentLoaded', function() {
    console.log('Initializing app...');

    // Initialize HTMX
    htmx.logger = function(elt, event, data) {
        if (console) {
            console.log(event, elt, data);
        }
    };

    // Add HTMX request headers
    document.body.addEventListener('htmx:configRequest', function(evt) {
        const token = getAuthToken();
        if (token) {
            evt.detail.headers['Authorization'] = `Bearer ${token}`;
        }
    });

    // Handle HTMX response errors
    document.body.addEventListener('htmx:responseError', function(evt) {
        const status = evt.detail.xhr.status;
        if (status === 401) {
            console.log('Unauthorized request, redirecting to login...');
            window.location.href = '/login';
        } else {
            const error = evt.detail.xhr.responseText || 'An error occurred. Please try again.';
            showToast(error, 'error');
            evt.preventDefault(); // Prevent default error handling
        }
    });

    // Handle HTMX after swap
    document.body.addEventListener('htmx:afterSwap', function(evt) {
        // Re-initialize any components that need it after content swap
        const path = evt.detail.pathInfo.requestPath;
        console.log('Content swapped for path:', path);
    });

    // Handle HTMX after request
    document.body.addEventListener('htmx:afterRequest', function(evt) {
        const path = evt.detail.pathInfo.requestPath;
        console.log('Request completed for path:', path);
        
        // Handle specific paths
        if (path === '/api/accounts' && !evt.detail.failed) {
            // Successful accounts request, ensure we stay on the accounts page
            evt.preventDefault(); // Prevent any default navigation
        }
    });
});

// Handle transaction amount formatting
document.addEventListener('input', function(e) {
    if (e.target.matches('input[name="amount"]')) {
        const value = e.target.value.replace(/[^\d.]/g, '');
        if (value) {
            const formatted = formatCurrency(parseFloat(value));
            e.target.value = formatted;
        }
    }
});
