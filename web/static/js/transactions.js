document.addEventListener('DOMContentLoaded', function() {
    // Check authentication
    const token = getAuthToken();
    if (!token) {
        window.location.href = '/login';
        return;
    }

    // Load transactions
    loadTransactions();

    // Load accounts for the form
    loadAccounts();

    // Load categories for the form
    loadCategories();
});

function loadTransactions() {
    const token = getAuthToken();
    fetch('/api/transactions', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (!response.ok) throw new Error('Failed to load transactions');
        return response.json();
    })
    .then(data => {
        const transactionsList = document.getElementById('transactions-list');
        const loadingElement = document.getElementById('loading-transactions');
        const noTransactionsElement = document.getElementById('no-transactions');
        
        // Hide loading state
        if (loadingElement) {
            loadingElement.classList.add('hidden');
        }

        if (!data.transactions || data.transactions.length === 0) {
            if (noTransactionsElement) {
                noTransactionsElement.classList.remove('hidden');
            }
            if (transactionsList) {
                transactionsList.innerHTML = `
                    <tr>
                        <td colspan="5" class="px-3 py-4 text-sm text-gray-500 text-center">
                            No transactions available
                        </td>
                    </tr>`;
            }
            return;
        }

        if (noTransactionsElement) {
            noTransactionsElement.classList.add('hidden');
        }

        transactionsList.innerHTML = data.transactions.map(transaction => `
            <tr>
                <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm text-gray-900 sm:pl-0">
                    ${new Date(transaction.date).toLocaleDateString()}
                </td>
                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                    ${transaction.description}
                </td>
                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                    ${transaction.category}
                </td>
                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                    ${transaction.account_name}
                </td>
                <td class="whitespace-nowrap px-3 py-4 text-sm text-right ${transaction.type === 'expense' ? 'text-red-600' : 'text-green-600'}">
                    ${formatCurrency(transaction.amount)}
                </td>
                <td class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-0">
                    <button onclick="editTransaction('${transaction.id}')" class="text-primary-600 hover:text-primary-900 mr-4">
                        Edit
                    </button>
                    <button onclick="deleteTransaction('${transaction.id}')" class="text-red-600 hover:text-red-900">
                        Delete
                    </button>
                </td>
            </tr>
        `).join('');
    })
    .catch(error => {
        console.error('Error loading transactions:', error);
        showToast(error.message, 'error');
    });
}

function loadAccounts() {
    const token = getAuthToken();
    fetch('/api/accounts', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (!response.ok) throw new Error('Failed to load accounts');
        return response.json();
    })
    .then(data => {
        const accountSelect = document.getElementById('account_id');
        if (!data.accounts || data.accounts.length === 0) {
            accountSelect.innerHTML = '<option value="">No accounts available</option>';
            return;
        }

        accountSelect.innerHTML = data.accounts.map(account => `
            <option value="${account.id}">${account.name}</option>
        `).join('');
    })
    .catch(error => {
        console.error('Error loading accounts:', error);
        showToast(error.message, 'error');
    });
}

function loadCategories() {
    const token = getAuthToken();
    fetch('/api/categories', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (!response.ok) throw new Error('Failed to load categories');
        return response.json();
    })
    .then(data => {
        const categorySelect = document.getElementById('category');
        if (!data.categories || data.categories.length === 0) {
            categorySelect.innerHTML = '<option value="">No categories available</option>';
            return;
        }

        categorySelect.innerHTML = data.categories.map(category => `
            <option value="${category.id}">${category.name}</option>
        `).join('');
    })
    .catch(error => {
        console.error('Error loading categories:', error);
        showToast(error.message, 'error');
    });
}

function showAddTransactionModal() {
    document.getElementById('add-transaction-modal').classList.remove('hidden');
    document.getElementById('date').valueAsDate = new Date();
}

function hideAddTransactionModal() {
    document.getElementById('add-transaction-modal').classList.add('hidden');
    document.getElementById('add-transaction-form').reset();
}

function addTransaction() {
    const form = document.getElementById('add-transaction-form');
    const formData = new FormData(form);
    const data = Object.fromEntries(formData.entries());
    
    // Convert amount to number
    data.amount = parseFloat(data.amount);
    if (data.type === 'expense') {
        data.amount = -Math.abs(data.amount);
    }

    const token = getAuthToken();
    fetch('/api/transactions', {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data)
    })
    .then(response => {
        if (!response.ok) throw new Error('Failed to add transaction');
        return response.json();
    })
    .then(() => {
        hideAddTransactionModal();
        showToast('Transaction added successfully');
        loadTransactions();
    })
    .catch(error => {
        console.error('Error adding transaction:', error);
        showToast(error.message, 'error');
    });
}

function editTransaction(id) {
    // TODO: Implement edit transaction functionality
    console.log('Edit transaction:', id);
}

function deleteTransaction(id) {
    if (!confirm('Are you sure you want to delete this transaction?')) {
        return;
    }

    const token = getAuthToken();
    fetch(`/api/transactions/${id}`, {
        method: 'DELETE',
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (!response.ok) throw new Error('Failed to delete transaction');
        showToast('Transaction deleted successfully');
        loadTransactions();
    })
    .catch(error => {
        console.error('Error deleting transaction:', error);
        showToast(error.message, 'error');
    });
}
