package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/model"
)

// Repository defines the interface for all repository operations
type Repository interface {
	// User methods
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error

	// Account methods
	CreateAccount(ctx context.Context, account *model.Account) error
	GetAccountByID(ctx context.Context, id string) (*model.Account, error)
	GetAccountsByUserID(ctx context.Context, userID string) ([]*model.Account, error)
	UpdateAccount(ctx context.Context, account *model.Account) error
	SavePlaidCredentials(ctx context.Context, creds *model.PlaidCredentials) error
	GetPlaidCredentials(ctx context.Context, userID string) (*model.PlaidCredentials, error)
	GetTotalAssets(ctx context.Context) (float64, error)
	GetTotalAssetsByUser(ctx context.Context, userID string) (float64, error)
	GetTotalLiabilities(ctx context.Context) (float64, error)
	GetTotalLiabilitiesByUser(ctx context.Context, userID string) (float64, error)

	// Transaction methods
	CreateTransaction(ctx context.Context, transaction *model.Transaction) error
	GetTransactionByID(ctx context.Context, id string) (*model.Transaction, error)
	GetTransactionsByAccountID(ctx context.Context, accountID string) ([]*model.Transaction, error)
	GetTransactionsByUserID(ctx context.Context, userID string) ([]*model.Transaction, error)
	GetTransactions(ctx context.Context, filter model.TransactionFilter) ([]*model.Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *model.Transaction) error
	DeleteTransaction(ctx context.Context, id string) error

	// Budget methods
	CreateBudget(ctx context.Context, budget *model.Budget) error
	GetBudgetByID(ctx context.Context, id string) (*model.Budget, error)
	GetBudgetsByUserID(ctx context.Context, userID string) ([]*model.Budget, error)
	GetBudgets(ctx context.Context, userID string, filter model.BudgetFilter) ([]*model.Budget, error)
	GetBudgetSummary(ctx context.Context, userID string, period string) (*model.BudgetSummary, error)
	UpdateBudget(ctx context.Context, budget *model.Budget) error
	DeleteBudget(ctx context.Context, id string) error

	// Goal methods
	CreateGoal(ctx context.Context, goal *model.Goal) error
	GetGoalByID(ctx context.Context, id string) (*model.Goal, error)
	GetGoalsByUserID(ctx context.Context, userID string) ([]*model.Goal, error)
	UpdateGoal(ctx context.Context, goal *model.Goal) error
	DeleteGoal(ctx context.Context, id string, userID string) error

	// Notification methods
	CreateNotification(ctx context.Context, notification *model.Notification) error
	GetNotificationByID(ctx context.Context, id string) (*model.Notification, error)
	GetUserNotifications(ctx context.Context, userID string, unreadOnly bool) ([]*model.Notification, error)
	MarkNotificationAsRead(ctx context.Context, userID string, notificationID string) error
	DeleteNotification(ctx context.Context, id string) error
	GetNotificationPreferences(ctx context.Context, userID string) (*model.NotificationPreferences, error)
	UpdateNotificationPreferences(ctx context.Context, prefs *model.NotificationPreferences) error

	// Category methods
	GetCategoryByID(ctx context.Context, id string) (*model.Category, error)
	CreateCategory(ctx context.Context, category *model.Category) error
	GetCategories(ctx context.Context, userID string) ([]*model.Category, error)
	UpdateCategory(ctx context.Context, category *model.Category) error
	DeleteCategory(ctx context.Context, id string) error
	InitializeDefaultCategories(ctx context.Context, userID string) error

	// Recurring Transaction methods
	CreateRecurringTransaction(ctx context.Context, transaction *model.RecurringTransaction) error
	GetRecurringTransactionByID(ctx context.Context, id string) (*model.RecurringTransaction, error)
	GetRecurringTransactions(ctx context.Context, filter model.RecurringTransactionFilter) ([]*model.RecurringTransaction, error)
	UpdateRecurringTransaction(ctx context.Context, transaction *model.RecurringTransaction) error
	DeleteRecurringTransaction(ctx context.Context, id string) error
	GetDueRecurringTransactions(ctx context.Context, before time.Time) ([]*model.RecurringTransaction, error)
	UpdateLastRun(ctx context.Context, id string, lastRun, nextRun time.Time) error

	// Analytics methods
	GetTotalIncome(ctx context.Context) (float64, error)
	GetTotalExpenses(ctx context.Context) (float64, error)
	GetAverageDailyExpenses(ctx context.Context) (float64, error)
	GetMonthlySpending(ctx context.Context, filter model.AnalyticsFilter) ([]model.MonthlySpending, error)
	GetCashFlow(ctx context.Context, filter model.AnalyticsFilter) (*model.CashFlow, error)
	GetTopMerchants(ctx context.Context, filter model.AnalyticsFilter) ([]model.MerchantSpending, error)
	GetFinancialReport(ctx context.Context, filter model.AnalyticsFilter) (*model.FinancialReport, error)
	GetIncomeVsExpenses(ctx context.Context, filter model.AnalyticsFilter) (*model.IncomeVsExpenses, error)
	GetIncomeVsExpensesByUser(ctx context.Context, userID string) (*model.IncomeVsExpenses, error)
	GetBudgetPerformance(ctx context.Context, filter model.AnalyticsFilter) (*model.BudgetPerformance, error)
	GetSpendingByCategory(ctx context.Context, filter model.AnalyticsFilter) ([]model.SpendingByCategory, error)
	GetUserCount(ctx context.Context) (int64, error)
	GetMonthlyDebtPayments(ctx context.Context) (float64, error)
	GetMonthlyDebtPaymentsByUser(ctx context.Context, userID string) (float64, error)
	GetEmergencyFundBalance(ctx context.Context) (float64, error)
	GetEmergencyFundBalanceByUser(ctx context.Context, userID string) (float64, error)
	GetAverageMonthlyExpenses(ctx context.Context) (float64, error)
	GetAverageMonthlyExpensesByUser(ctx context.Context, userID string) (float64, error)
}

// SQLRepository struct
type SQLRepository struct {
	db           *sql.DB
	user         *UserSQL
	account      *AccountSQL
	transaction  *TransactionSQL
	budget       *BudgetSQL
	goal         *GoalSQL
	notification *NotificationSQL
	analytics    *AnalyticsSQL
	category     *CategorySQL
	recurring    *RecurringTransactionSQL
	recurringTx  *RecurringTransactionSQL
}

// NewRepository creates a new SQLRepository
func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{
		db:           db,
		user:         &UserSQL{db: db},
		account:      &AccountSQL{db: db},
		transaction:  &TransactionSQL{db: db},
		budget:       &BudgetSQL{db: db},
		goal:         &GoalSQL{db: db},
		notification: &NotificationSQL{db: db},
		analytics:    &AnalyticsSQL{db: db},
		category:     &CategorySQL{db: db},
		recurring:    &RecurringTransactionSQL{db: db},
		recurringTx:  &RecurringTransactionSQL{db: db},
	}
}

// WithTx wraps repository with transaction
func (r *SQLRepository) WithTx(tx *sql.Tx) *SQLRepository {
	return &SQLRepository{
		db:           r.db,
		user:         &UserSQL{db: r.db, tx: tx},
		account:      &AccountSQL{db: r.db, tx: tx},
		transaction:  &TransactionSQL{db: r.db, tx: tx},
		budget:       &BudgetSQL{db: r.db, tx: tx},
		goal:         &GoalSQL{db: r.db, tx: tx},
		notification: &NotificationSQL{db: r.db, tx: tx},
		analytics:    &AnalyticsSQL{db: r.db, tx: tx},
		category:     &CategorySQL{db: r.db, tx: tx},
		recurring:    &RecurringTransactionSQL{db: r.db, tx: tx},
		recurringTx:  &RecurringTransactionSQL{db: r.db, tx: tx},
	}
}

// User methods
func (r *SQLRepository) CreateUser(ctx context.Context, user *model.User) error {
	return r.user.CreateUser(ctx, user)
}

func (r *SQLRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	return r.user.GetUserByID(ctx, id)
}

func (r *SQLRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.user.GetUserByEmail(ctx, email)
}

func (r *SQLRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return r.user.GetUserByUsername(ctx, username)
}

func (r *SQLRepository) UpdateUser(ctx context.Context, user *model.User) error {
	return r.user.UpdateUser(ctx, user)
}

func (r *SQLRepository) DeleteUser(ctx context.Context, id string) error {
	return r.user.DeleteUser(ctx, id)
}

// Account methods
func (r *SQLRepository) CreateAccount(ctx context.Context, account *model.Account) error {
	return r.account.CreateAccount(ctx, account)
}

func (r *SQLRepository) GetAccountByID(ctx context.Context, id string) (*model.Account, error) {
	return r.account.GetAccountByID(ctx, id)
}

func (r *SQLRepository) GetAccountsByUserID(ctx context.Context, userID string) ([]*model.Account, error) {
	return r.account.GetAccountsByUserID(ctx, userID)
}

func (r *SQLRepository) UpdateAccount(ctx context.Context, account *model.Account) error {
	return r.account.UpdateAccount(ctx, account)
}

func (r *SQLRepository) SavePlaidCredentials(ctx context.Context, creds *model.PlaidCredentials) error {
	return r.account.SavePlaidCredentials(ctx, creds)
}

func (r *SQLRepository) GetPlaidCredentials(ctx context.Context, userID string) (*model.PlaidCredentials, error) {
	return r.account.GetPlaidCredentials(ctx, userID)
}

func (r *SQLRepository) GetTotalAssets(ctx context.Context) (float64, error) {
	return r.account.GetTotalAssets(ctx)
}

func (r *SQLRepository) GetTotalAssetsByUser(ctx context.Context, userID string) (float64, error) {
	return r.account.GetTotalAssetsByUser(ctx, userID)
}

func (r *SQLRepository) GetTotalLiabilities(ctx context.Context) (float64, error) {
	return r.account.GetTotalLiabilities(ctx)
}

func (r *SQLRepository) GetTotalLiabilitiesByUser(ctx context.Context, userID string) (float64, error) {
	return r.account.GetTotalLiabilitiesByUser(ctx, userID)
}

// Transaction methods
func (r *SQLRepository) CreateTransaction(ctx context.Context, transaction *model.Transaction) error {
	return r.transaction.CreateTransaction(ctx, transaction)
}

func (r *SQLRepository) GetTransactionByID(ctx context.Context, id string) (*model.Transaction, error) {
	return r.transaction.GetTransactionByID(ctx, id)
}

func (r *SQLRepository) GetTransactionsByAccountID(ctx context.Context, accountID string) ([]*model.Transaction, error) {
	filter := model.TransactionFilter{
		AccountID: accountID,
	}
	return r.transaction.GetTransactions(ctx, filter)
}

func (r *SQLRepository) GetTransactionsByUserID(ctx context.Context, userID string) ([]*model.Transaction, error) {
	// First get all accounts for the user
	accounts, err := r.account.GetAccountsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// If user has no accounts, return empty slice
	if len(accounts) == 0 {
		return []*model.Transaction{}, nil
	}

	// Get transactions for all user accounts
	var allTransactions []*model.Transaction
	for _, account := range accounts {
		filter := model.TransactionFilter{
			AccountID: account.ID,
		}
		transactions, err := r.transaction.GetTransactions(ctx, filter)
		if err != nil {
			return nil, err
		}
		allTransactions = append(allTransactions, transactions...)
	}

	return allTransactions, nil
}

func (r *SQLRepository) GetTransactions(ctx context.Context, filter model.TransactionFilter) ([]*model.Transaction, error) {
	return r.transaction.GetTransactions(ctx, filter)
}

func (r *SQLRepository) UpdateTransaction(ctx context.Context, transaction *model.Transaction) error {
	return r.transaction.UpdateTransaction(ctx, transaction)
}

func (r *SQLRepository) DeleteTransaction(ctx context.Context, id string) error {
	return r.transaction.DeleteTransaction(ctx, id)
}

// Budget methods
func (r *SQLRepository) CreateBudget(ctx context.Context, budget *model.Budget) error {
	return r.budget.CreateBudget(ctx, budget)
}

func (r *SQLRepository) GetBudgetByID(ctx context.Context, id string) (*model.Budget, error) {
	return r.budget.GetBudgetByID(ctx, id)
}

func (r *SQLRepository) GetBudgetsByUserID(ctx context.Context, userID string) ([]*model.Budget, error) {
	filter := model.BudgetFilter{
		UserID: userID,
	}
	return r.budget.GetBudgets(ctx, filter)
}

func (r *SQLRepository) GetBudgets(ctx context.Context, userID string, filter model.BudgetFilter) ([]*model.Budget, error) {
	return r.budget.GetBudgets(ctx, filter)
}

func (r *SQLRepository) GetBudgetSummary(ctx context.Context, userID string, period string) (*model.BudgetSummary, error) {
	// Convert period string to start and end times
	start, end, err := model.GetPeriodDates(period)
	if err != nil {
		return nil, err
	}
	return r.budget.GetBudgetSummary(ctx, userID, start, end)
}

func (r *SQLRepository) UpdateBudget(ctx context.Context, budget *model.Budget) error {
	return r.budget.UpdateBudget(ctx, budget)
}

func (r *SQLRepository) DeleteBudget(ctx context.Context, id string) error {
	return r.budget.DeleteBudget(ctx, id)
}

// Goal methods
func (r *SQLRepository) CreateGoal(ctx context.Context, goal *model.Goal) error {
	return r.goal.CreateGoal(ctx, goal)
}

func (r *SQLRepository) GetGoalByID(ctx context.Context, id string) (*model.Goal, error) {
	return r.goal.GetGoalByID(ctx, id)
}

func (r *SQLRepository) GetGoalsByUserID(ctx context.Context, userID string) ([]*model.Goal, error) {
	return r.goal.GetUserGoals(ctx, userID)
}

func (r *SQLRepository) UpdateGoal(ctx context.Context, goal *model.Goal) error {
	return r.goal.UpdateGoal(ctx, goal)
}

func (r *SQLRepository) DeleteGoal(ctx context.Context, id string, userID string) error {
	return r.goal.DeleteGoal(ctx, id, userID)
}

// Notification methods
func (r *SQLRepository) CreateNotification(ctx context.Context, notification *model.Notification) error {
	return r.notification.CreateNotification(ctx, notification)
}

func (r *SQLRepository) GetNotificationByID(ctx context.Context, id string) (*model.Notification, error) {
	return r.notification.GetNotificationByID(ctx, id)
}

func (r *SQLRepository) GetUserNotifications(ctx context.Context, userID string, unreadOnly bool) ([]*model.Notification, error) {
	return r.notification.GetUserNotifications(ctx, userID, unreadOnly)
}

func (r *SQLRepository) MarkNotificationAsRead(ctx context.Context, userID string, notificationID string) error {
	return r.notification.MarkNotificationAsRead(ctx, userID, notificationID)
}

func (r *SQLRepository) DeleteNotification(ctx context.Context, id string) error {
	return r.notification.DeleteNotification(ctx, id)
}

func (r *SQLRepository) GetNotificationPreferences(ctx context.Context, userID string) (*model.NotificationPreferences, error) {
	return r.notification.GetNotificationPreferences(ctx, userID)
}

func (r *SQLRepository) UpdateNotificationPreferences(ctx context.Context, prefs *model.NotificationPreferences) error {
	return r.notification.UpdateNotificationPreferences(ctx, prefs)
}

// Category methods
func (r *SQLRepository) GetCategoryByID(ctx context.Context, id string) (*model.Category, error) {
	return r.category.GetCategoryByID(ctx, id)
}

func (r *SQLRepository) CreateCategory(ctx context.Context, category *model.Category) error {
	return r.category.CreateCategory(ctx, category)
}

func (r *SQLRepository) GetCategories(ctx context.Context, userID string) ([]*model.Category, error) {
	return r.category.GetCategories(ctx, userID)
}

func (r *SQLRepository) UpdateCategory(ctx context.Context, category *model.Category) error {
	return r.category.UpdateCategory(ctx, category)
}

func (r *SQLRepository) DeleteCategory(ctx context.Context, id string) error {
	return r.category.DeleteCategory(ctx, id)
}

func (r *SQLRepository) InitializeDefaultCategories(ctx context.Context, userID string) error {
	return r.category.InitializeDefaultCategories(ctx, userID)
}

// Recurring Transaction methods
func (r *SQLRepository) CreateRecurringTransaction(ctx context.Context, transaction *model.RecurringTransaction) error {
	return r.recurring.CreateRecurringTransaction(ctx, transaction)
}

func (r *SQLRepository) GetRecurringTransactionByID(ctx context.Context, id string) (*model.RecurringTransaction, error) {
	return r.recurring.GetRecurringTransactionByID(ctx, id)
}

func (r *SQLRepository) GetRecurringTransactions(ctx context.Context, filter model.RecurringTransactionFilter) ([]*model.RecurringTransaction, error) {
	return r.recurring.GetRecurringTransactions(ctx, filter)
}

func (r *SQLRepository) UpdateRecurringTransaction(ctx context.Context, transaction *model.RecurringTransaction) error {
	return r.recurring.UpdateRecurringTransaction(ctx, transaction)
}

func (r *SQLRepository) DeleteRecurringTransaction(ctx context.Context, id string) error {
	return r.recurring.DeleteRecurringTransaction(ctx, id)
}

func (r *SQLRepository) GetDueRecurringTransactions(ctx context.Context, before time.Time) ([]*model.RecurringTransaction, error) {
	return r.recurring.GetDueRecurringTransactions(ctx, before)
}

func (r *SQLRepository) UpdateLastRun(ctx context.Context, id string, lastRun, nextRun time.Time) error {
	return r.recurring.UpdateLastRun(ctx, id, lastRun, nextRun)
}

// Analytics methods
func (r *SQLRepository) GetTotalIncome(ctx context.Context) (float64, error) {
	return r.analytics.GetTotalIncome(ctx)
}

func (r *SQLRepository) GetTotalExpenses(ctx context.Context) (float64, error) {
	return r.analytics.GetTotalExpenses(ctx)
}

func (r *SQLRepository) GetAverageDailyExpenses(ctx context.Context) (float64, error) {
	return r.analytics.GetAverageDailyExpenses(ctx)
}

func (r *SQLRepository) GetMonthlySpending(ctx context.Context, filter model.AnalyticsFilter) ([]model.MonthlySpending, error) {
	return r.analytics.GetMonthlySpending(ctx, filter)
}

func (r *SQLRepository) GetCashFlow(ctx context.Context, filter model.AnalyticsFilter) (*model.CashFlow, error) {
	return r.analytics.GetCashFlow(ctx, filter)
}

func (r *SQLRepository) GetTopMerchants(ctx context.Context, filter model.AnalyticsFilter) ([]model.MerchantSpending, error) {
	return r.analytics.GetTopMerchants(ctx, filter)
}

func (r *SQLRepository) GetFinancialReport(ctx context.Context, filter model.AnalyticsFilter) (*model.FinancialReport, error) {
	return r.analytics.GetFinancialReport(ctx, filter)
}

func (r *SQLRepository) GetIncomeVsExpenses(ctx context.Context, filter model.AnalyticsFilter) (*model.IncomeVsExpenses, error) {
	return r.analytics.GetIncomeVsExpenses(ctx, filter)
}

func (r *SQLRepository) GetIncomeVsExpensesByUser(ctx context.Context, userID string) (*model.IncomeVsExpenses, error) {
	return r.analytics.GetIncomeVsExpensesByUser(ctx, userID)
}

func (r *SQLRepository) GetBudgetPerformance(ctx context.Context, filter model.AnalyticsFilter) (*model.BudgetPerformance, error) {
	return r.analytics.GetBudgetPerformance(ctx, filter)
}

func (r *SQLRepository) GetSpendingByCategory(ctx context.Context, filter model.AnalyticsFilter) ([]model.SpendingByCategory, error) {
	return r.analytics.GetSpendingByCategory(ctx, filter)
}

func (r *SQLRepository) GetUserCount(ctx context.Context) (int64, error) {
	return r.analytics.GetUserCount(ctx)
}

func (r *SQLRepository) GetMonthlyDebtPayments(ctx context.Context) (float64, error) {
	return r.analytics.GetMonthlyDebtPayments(ctx)
}

func (r *SQLRepository) GetMonthlyDebtPaymentsByUser(ctx context.Context, userID string) (float64, error) {
	return r.analytics.GetMonthlyDebtPaymentsByUser(ctx, userID)
}

func (r *SQLRepository) GetEmergencyFundBalance(ctx context.Context) (float64, error) {
	return r.analytics.GetEmergencyFundBalance(ctx)
}

func (r *SQLRepository) GetEmergencyFundBalanceByUser(ctx context.Context, userID string) (float64, error) {
	return r.analytics.GetEmergencyFundBalanceByUser(ctx, userID)
}

func (r *SQLRepository) GetAverageMonthlyExpenses(ctx context.Context) (float64, error) {
	return r.analytics.GetAverageMonthlyExpenses(ctx)
}

func (r *SQLRepository) GetAverageMonthlyExpensesByUser(ctx context.Context, userID string) (float64, error) {
	return r.analytics.GetAverageMonthlyExpensesByUser(ctx, userID)
}
