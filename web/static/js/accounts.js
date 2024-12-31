// Initialize the accounts page
document.addEventListener('DOMContentLoaded', function() {
    // Check if we're on the accounts page
    if (window.location.pathname === '/accounts') {
        loadAccounts();
    }
});

// Load accounts from the API
async function loadAccounts() {
    const accountsList = document.getElementById('accounts-list');
    const noAccounts = document.getElementById('no-accounts');
    if (!accountsList) return;

    try {
        console.log('Loading accounts...');
        const response = await fetch('/api/accounts', {
            credentials: 'include',
            headers: {
                'Accept': 'application/json'
            }
        });

        if (!response.ok) {
            const data = await response.json();
            console.error('Error response:', data);
            if (response.status === 401) {
                window.location.href = '/login';
                return;
            }
            throw new Error(data.error || 'Failed to load accounts');
        }
        
        const data = await response.json();
        console.log('Accounts loaded:', data);
        
        updateAccountsList(data.accounts);
    } catch (error) {
        console.error('Error loading accounts:', error);
        showToast('Failed to load accounts. Please try again.', 'error');
    }
}

// Update accounts list
function updateAccountsList(accounts) {
    const accountsList = document.getElementById('accounts-list');
    const noAccounts = document.getElementById('no-accounts');
    if (!accountsList) return;

    if (!accounts || accounts.length === 0) {
        accountsList.innerHTML = '';
        noAccounts.classList.remove('hidden');
        return;
    }

    noAccounts.classList.add('hidden');
    accountsList.innerHTML = accounts.map(account => `
        <div class="bg-white overflow-hidden shadow rounded-lg divide-y divide-gray-200">
            <div class="px-4 py-5 sm:p-6">
                <div class="flex justify-between items-start">
                    <div class="space-y-1">
                        <h3 class="text-lg font-medium text-gray-900">${account.name}</h3>
                        <p class="text-sm text-gray-500">${account.type}</p>
                    </div>
                    <div class="text-right">
                        <p class="text-lg font-semibold ${account.balance >= 0 ? 'text-green-600' : 'text-red-600'}">
                            ${formatCurrency(account.balance)}
                        </p>
                        <p class="text-sm text-gray-500">${account.currency}</p>
                    </div>
                </div>
            </div>
            <div class="px-4 py-4 sm:px-6">
                <button onclick="loadAccountForm('${account.id}')" class="text-primary-600 hover:text-primary-900">
                    Edit
                </button>
            </div>
        </div>
    `).join('');
}

// Format currency with 2 decimal places
function formatCurrency(amount) {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD'
    }).format(amount);
}

// Show add account modal
function showAddAccountModal() {
    const modal = document.getElementById('add-account-modal');
    if (!modal) return;
    modal.classList.remove('hidden');
}

// Hide add account modal
function hideAddAccountModal() {
    const modal = document.getElementById('add-account-modal');
    if (!modal) return;
    modal.classList.add('hidden');
}

// Add new account
async function addAccount(event) {
    console.log('Form submission started');
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);

    try {
        const accountData = {
            name: formData.get('name'),
            type: formData.get('type'),
            balance: parseFloat(formData.get('balance')),
            currency: formData.get('currency')
        };
        
        console.log('Creating account with data:', accountData);

        const response = await fetch('/api/accounts', {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify(accountData)
        });

        console.log('Response status:', response.status);
        console.log('Response headers:', Object.fromEntries(response.headers.entries()));

        if (!response.ok) {
            const data = await response.json();
            console.error('Error response:', data);
            if (response.status === 401) {
                window.location.href = '/login';
                return;
            }
            throw new Error(data.error || 'Failed to create account');
        }

        const data = await response.json();
        console.log('Account created:', data);

        // Update the accounts list with the new data
        updateAccountsList(data.accounts);
        
        hideAddAccountModal();
        showToast('Account created successfully!', 'success');
        form.reset();
    } catch (error) {
        console.error('Error creating account:', error);
        console.error('Error details:', error.stack);
        showToast(error.message || 'Failed to create account. Please try again.', 'error');
    }
}

// Load account form
async function loadAccountForm(id) {
    const isNew = id === 'new';
    let account = { name: '', type: 'checking', balance: '0.00' };
    
    if (!isNew) {
        try {
            const response = await fetch(`/api/accounts/${id}`, {
                credentials: 'include',
                headers: {
                    'Accept': 'application/json'
                }
            });

            if (!response.ok) {
                const data = await response.json();
                if (response.status === 401) {
                    window.location.href = '/login';
                    return;
                }
                throw new Error(data.error || 'Failed to load account');
            }

            account = await response.json();
        } catch (error) {
            console.error('Error loading account:', error);
            showToast('Error loading account', 'error');
            return;
        }
    }

    const modal = document.getElementById('account-modal');
    modal.innerHTML = `
        <div class="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"></div>
            <div class="relative inline-block align-bottom bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full sm:p-6">
                <div class="sm:flex sm:items-start">
                    <div class="mt-3 text-center sm:mt-0 sm:text-left w-full">
                        <h3 class="text-lg leading-6 font-medium text-gray-900" id="modal-title">
                            ${isNew ? 'Add New Account' : 'Edit Account'}
                        </h3>
                        <div class="mt-4">
                            <form id="account-form" onsubmit="event.preventDefault(); saveAccount('${id}');">
                                <div class="space-y-4">
                                    <div>
                                        <label for="name" class="block text-sm font-medium text-gray-700">Account Name</label>
                                        <input type="text" name="name" id="name" required
                                               class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
                                               value="${account.name}">
                                    </div>
                                    <div>
                                        <label for="type" class="block text-sm font-medium text-gray-700">Account Type</label>
                                        <select name="type" id="type" required
                                                class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm">
                                            <option value="checking" ${account.type === 'checking' ? 'selected' : ''}>Checking</option>
                                            <option value="savings" ${account.type === 'savings' ? 'selected' : ''}>Savings</option>
                                            <option value="credit" ${account.type === 'credit' ? 'selected' : ''}>Credit Card</option>
                                            <option value="investment" ${account.type === 'investment' ? 'selected' : ''}>Investment</option>
                                        </select>
                                    </div>
                                    <div>
                                        <label for="balance" class="block text-sm font-medium text-gray-700">Balance</label>
                                        <div class="mt-1 relative rounded-md shadow-sm">
                                            <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                                <span class="text-gray-500 sm:text-sm">$</span>
                                            </div>
                                            <input type="text" name="balance" id="balance" required
                                                   class="block w-full pl-7 pr-12 border border-gray-300 rounded-md focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
                                                   value="${parseFloat(account.balance).toFixed(2)}">
                                        </div>
                                    </div>
                                </div>
                                <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                                    <button type="submit"
                                            class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-primary-600 text-base font-medium text-white hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 sm:ml-3 sm:w-auto sm:text-sm">
                                        ${isNew ? 'Create Account' : 'Save Changes'}
                                    </button>
                                    <button type="button"
                                            onclick="closeModal()"
                                            class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 sm:mt-0 sm:w-auto sm:text-sm">
                                        Cancel
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `;
    modal.classList.remove('hidden');
}

// Close the modal
function closeModal() {
    const modal = document.getElementById('account-modal');
    modal.classList.add('hidden');
}

// Save account
async function saveAccount(id) {
    const form = document.getElementById('account-form');
    const formData = new FormData(form);
    
    const accountData = {
        name: formData.get('name'),
        type: formData.get('type'),
        balance: parseFloat(formData.get('balance'))
    };

    try {
        const url = id === 'new' ? '/api/accounts' : `/api/accounts/${id}`;
        const method = id === 'new' ? 'POST' : 'PUT';
        
        const response = await fetch(url, {
            method: method,
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify(accountData)
        });

        if (!response.ok) {
            const data = await response.json();
            if (response.status === 401) {
                window.location.href = '/login';
                return;
            }
            throw new Error(data.error || 'Failed to save account');
        }

        const data = await response.json();
        console.log('Account saved:', data);

        closeModal();
        await loadAccounts();
        showToast(`Account ${id === 'new' ? 'created' : 'updated'} successfully`);
    } catch (error) {
        console.error('Error saving account:', error);
        showToast('Error saving account', 'error');
    }
}

// Handle currency input formatting
document.addEventListener('input', function(e) {
    if (e.target.matches('input[name="balance"]')) {
        const value = e.target.value.replace(/[^\d.]/g, '');
        if (value === '') {
            e.target.value = '';
        } else {
            const num = parseFloat(value);
            if (!isNaN(num)) {
                e.target.value = num.toFixed(2);
            }
        }
    }
});
