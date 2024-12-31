// Authentication utilities
function getAuthToken() {
    // Get token from cookie
    const cookies = document.cookie.split(';');
    for (let cookie of cookies) {
        const [name, value] = cookie.trim().split('=');
        if (name === 'authToken') {
            return value;
        }
    }
    return null;
}

function isAuthenticated() {
    return !!getAuthToken();
}

// Handle logout
function logout() {
    // Remove auth cookie
    document.cookie = 'authToken=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
    window.location.href = '/login';
}

// Fetch utilities with authentication
async function authenticatedFetch(url, options = {}) {
    const token = getAuthToken();
    if (!token) {
        // Only redirect for page requests, not API calls
        if (!url.startsWith('/api/')) {
            sessionStorage.setItem('redirectUrl', window.location.pathname);
            window.location.href = '/login';
        }
        return null;
    }

    const defaultOptions = {
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
        credentials: 'include',
    };

    const mergedOptions = {
        ...defaultOptions,
        ...options,
        headers: {
            ...defaultOptions.headers,
            ...options.headers,
        },
    };

    try {
        const response = await fetch(url, mergedOptions);
        if (response.status === 401) {
            // Clear token
            document.cookie = 'authToken=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
            
            // Only redirect for page requests, not API calls
            if (!url.startsWith('/api/')) {
                sessionStorage.setItem('redirectUrl', window.location.pathname);
                window.location.href = '/login';
            }
            return null;
        }
        return response;
    } catch (error) {
        console.error('Fetch error:', error);
        return null;
    }
}
