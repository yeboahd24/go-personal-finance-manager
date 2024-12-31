package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/errors"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
)

type AnalyticsRepository interface {
	GetSpendingByCategory(ctx context.Context, filter model.AnalyticsFilter) ([]model.SpendingByCategory, error)
	GetMonthlySpending(ctx context.Context, filter model.AnalyticsFilter) ([]model.MonthlySpending, error)
	GetCashFlow(ctx context.Context, filter model.AnalyticsFilter) (*model.CashFlow, error)
	GetTopMerchants(ctx context.Context, filter model.AnalyticsFilter) ([]model.MerchantSpending, error)
	GetFinancialReport(ctx context.Context, filter model.AnalyticsFilter) (*model.FinancialReport, error)
	GetBudgetPerformance(ctx context.Context, filter model.AnalyticsFilter) (*model.BudgetPerformance, error)
	GetIncomeVsExpenses(ctx context.Context, filter model.AnalyticsFilter) (*model.IncomeVsExpenses, error)
	GetIncomeVsExpensesByUser(ctx context.Context, userID string) (*model.IncomeVsExpenses, error)
	GetMonthlyDebtPayments(ctx context.Context) (float64, error)
	GetMonthlyDebtPaymentsByUser(ctx context.Context, userID string) (float64, error)
	GetEmergencyFundBalance(ctx context.Context) (float64, error)
	GetEmergencyFundBalanceByUser(ctx context.Context, userID string) (float64, error)
	GetAverageMonthlyExpenses(ctx context.Context) (float64, error)
	GetAverageMonthlyExpensesByUser(ctx context.Context, userID string) (float64, error)
	GetTotalIncome(ctx context.Context) (float64, error)
	GetTotalExpenses(ctx context.Context) (float64, error)
	GetAverageDailyExpenses(ctx context.Context) (float64, error)
	GetUserCount(ctx context.Context) (int64, error)
}

type AnalyticsSQL struct {
	db *sql.DB
	tx *sql.Tx
}

func (r *AnalyticsSQL) query() QueryExecutor {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *AnalyticsSQL) GetSpendingByCategory(ctx context.Context, filter model.AnalyticsFilter) ([]model.SpendingByCategory, error) {
	sqlQuery := `
		WITH category_spending AS (
			SELECT 
				c.id as category_id,
				c.name as category_name,
				COALESCE(ABS(SUM(CASE WHEN t.type = 'debit' THEN t.amount ELSE 0 END)), 0) as amount
			FROM categories c
			LEFT JOIN transactions t ON t.category_id = c.id 
				AND t.user_id::text = $1::text
				AND t.date >= $2 
				AND t.date <= $3
				AND ($4::text IS NULL OR c.id::text = $4::text)
			WHERE c.type = 'expense'
				AND (c.user_id IS NULL OR c.user_id::text = $1::text)  -- Include both default and user-specific categories
			GROUP BY c.id, c.name
			HAVING COALESCE(ABS(SUM(CASE WHEN t.type = 'debit' THEN t.amount ELSE 0 END)), 0) > 0
		),
		total_spending AS (
			SELECT COALESCE(SUM(amount), 0) as total
			FROM category_spending
		)
		SELECT 
			cs.category_id,
			cs.category_name,
			cs.amount,
			CASE 
				WHEN ts.total = 0 THEN 0::numeric
				ELSE ROUND((cs.amount::numeric / ts.total::numeric * 100)::numeric, 2)
			END as percentage
		FROM category_spending cs
		CROSS JOIN total_spending ts
		ORDER BY cs.amount DESC
		LIMIT $5`

	rows, err := r.query().QueryContext(
		ctx,
		sqlQuery,
		filter.UserID,
		filter.StartDate,
		filter.EndDate,
		filter.CategoryID,
		filter.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var spending []model.SpendingByCategory
	for rows.Next() {
		var s model.SpendingByCategory
		if err := rows.Scan(&s.CategoryID, &s.CategoryName, &s.Amount, &s.Percentage); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		spending = append(spending, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after scanning rows: %w", err)
	}

	return spending, nil
}

func (r *AnalyticsSQL) GetMonthlySpending(ctx context.Context, filter model.AnalyticsFilter) ([]model.MonthlySpending, error) {
	query := `
		SELECT 
			DATE_TRUNC('month', t.date) as month,
			COALESCE(SUM(CASE WHEN t.type = 'expense' THEN ABS(t.amount) ELSE 0 END), 0) as amount
		FROM transactions t
		WHERE t.user_id = $1
			AND t.date >= $2
			AND t.date <= $3
			AND ($4::uuid IS NULL OR t.category_id = $4)
		GROUP BY DATE_TRUNC('month', t.date)
		ORDER BY month DESC`

	rows, err := r.query().QueryContext(ctx, query,
		filter.UserID,
		filter.StartDate,
		filter.EndDate,
		filter.CategoryID,
	)
	if err != nil {
		return nil, errors.New("Failed to get monthly spending", http.StatusInternalServerError)
	}
	defer rows.Close()

	var spending []model.MonthlySpending
	for rows.Next() {
		var s model.MonthlySpending
		if err := rows.Scan(&s.Month, &s.Amount); err != nil {
			return nil, errors.New("Failed to scan monthly spending", http.StatusInternalServerError)
		}
		spending = append(spending, s)
	}

	return spending, nil
}

func (r *AnalyticsSQL) GetCashFlow(ctx context.Context, filter model.AnalyticsFilter) (*model.CashFlow, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN ABS(amount) ELSE 0 END), 0) as expenses
		FROM transactions
		WHERE user_id = $1
			AND date >= $2
			AND date <= $3
			AND ($4::uuid IS NULL OR category_id = $4)`

	row := r.query().QueryRowContext(ctx, query,
		filter.UserID,
		filter.StartDate,
		filter.EndDate,
		filter.CategoryID,
	)
	cashFlow := &model.CashFlow{}
	err := row.Scan(&cashFlow.Income, &cashFlow.Expenses)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.New("Failed to get cash flow", http.StatusInternalServerError)
	}
	cashFlow.NetIncome = cashFlow.Income - cashFlow.Expenses
	return cashFlow, nil
}

func (r *AnalyticsSQL) GetTopMerchants(ctx context.Context, filter model.AnalyticsFilter) ([]model.MerchantSpending, error) {
	query := `
		WITH merchant_spending AS (
			SELECT 
				merchant,
				COALESCE(ABS(SUM(amount)), 0) as amount
			FROM transactions
			WHERE user_id = $1
				AND date >= $2
				AND date <= $3
				AND type = 'expense'
				AND ($4::uuid IS NULL OR category_id = $4)
				AND merchant IS NOT NULL
			GROUP BY merchant
			HAVING COALESCE(ABS(SUM(amount)), 0) > 0
		),
		total_spending AS (
			SELECT COALESCE(SUM(amount), 0) as total
			FROM merchant_spending
		)
		SELECT 
			ms.merchant,
			ms.amount,
			CASE 
				WHEN ts.total = 0 THEN 0
				ELSE ROUND((ms.amount::float / ts.total::float) * 100, 2)
			END as percentage
		FROM merchant_spending ms
		CROSS JOIN total_spending ts
		ORDER BY ms.amount DESC
		LIMIT 10`

	rows, err := r.query().QueryContext(ctx, query,
		filter.UserID,
		filter.StartDate,
		filter.EndDate,
		filter.CategoryID,
	)
	if err != nil {
		return nil, errors.New("Failed to get top merchants", http.StatusInternalServerError)
	}
	defer rows.Close()

	var merchants []model.MerchantSpending
	for rows.Next() {
		var m model.MerchantSpending
		if err := rows.Scan(&m.MerchantName, &m.Amount, &m.Percentage); err != nil {
			return nil, errors.New("Failed to scan merchant spending", http.StatusInternalServerError)
		}
		merchants = append(merchants, m)
	}

	return merchants, nil
}

func (r *AnalyticsSQL) GetFinancialReport(ctx context.Context, filter model.AnalyticsFilter) (*model.FinancialReport, error) {
	query := `
		WITH monthly_metrics AS (
			SELECT
				DATE_TRUNC('month', t.date) as month,
				COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE 0 END), 0) as income,
				COALESCE(SUM(CASE WHEN t.type = 'expense' THEN ABS(t.amount) ELSE 0 END), 0) as expenses
			FROM transactions t
			WHERE t.user_id = $1
				AND t.date >= $2
				AND t.date <= $3
				AND ($4::uuid IS NULL OR t.category_id = $4)
			GROUP BY DATE_TRUNC('month', t.date)
		),
		aggregated_metrics AS (
			SELECT
				COUNT(DISTINCT month) as num_months,
				SUM(income) as total_income,
				SUM(expenses) as total_expenses,
				AVG(income) as avg_monthly_income,
				AVG(expenses) as avg_monthly_expenses,
				MIN(expenses) as min_monthly_expenses,
				MAX(expenses) as max_monthly_expenses
			FROM monthly_metrics
		)
		SELECT
			total_income,
			total_expenses,
			avg_monthly_income,
			avg_monthly_expenses,
			min_monthly_expenses,
			max_monthly_expenses,
			CASE 
				WHEN total_expenses = 0 THEN 100
				ELSE (total_income - total_expenses) / total_expenses * 100
			END as savings_rate
		FROM aggregated_metrics`

	report := &model.FinancialReport{
		CashFlow: model.CashFlow{},
	}
	err := r.query().QueryRowContext(ctx, query,
		filter.UserID,
		filter.StartDate,
		filter.EndDate,
		filter.CategoryID,
	).Scan(
		&report.CashFlow.Income,
		&report.CashFlow.Expenses,
		&report.AvgMonthlyIncome,
		&report.AvgMonthlyExpenses,
		&report.MinMonthlyExpenses,
		&report.MaxMonthlyExpenses,
		&report.SavingsRate,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.New("Failed to get financial report", http.StatusInternalServerError)
	}

	// Get top spending categories
	categories, err := r.GetSpendingByCategory(ctx, model.AnalyticsFilter{
		UserID:    filter.UserID,
		StartDate: filter.StartDate,
		EndDate:   filter.EndDate,
		Limit:     5,
	})
	if err != nil {
		return nil, err
	}
	report.TopCategories = categories

	// Get top merchants
	merchants, err := r.GetTopMerchants(ctx, model.AnalyticsFilter{
		UserID:    filter.UserID,
		StartDate: filter.StartDate,
		EndDate:   filter.EndDate,
		Limit:     5,
	})
	if err != nil {
		return nil, err
	}
	report.TopMerchants = merchants

	// Get monthly trend
	monthlySpending, err := r.GetMonthlySpending(ctx, model.AnalyticsFilter{
		UserID:    filter.UserID,
		StartDate: filter.StartDate,
		EndDate:   filter.EndDate,
		Limit:     12,
	})
	if err != nil {
		return nil, err
	}
	report.MonthlyTrend = monthlySpending

	// Get average daily spend
	query = `
		SELECT COALESCE(ABS(SUM(amount)) / NULLIF(COUNT(DISTINCT DATE(date)), 0), 0)
		FROM transactions
		WHERE user_id = $1
			AND date >= $2
			AND date <= $3
			AND amount < 0`

	err = r.query().QueryRowContext(
		ctx,
		query,
		filter.UserID,
		filter.StartDate,
		filter.EndDate,
	).Scan(&report.AverageDailySpend)

	if err != nil && err != sql.ErrNoRows {
		return nil, errors.New("Failed to get average daily spend", http.StatusInternalServerError)
	}

	// Get largest expense
	query = `
		SELECT 
			id, user_id, account_id, category_id, 
			amount, date, description, merchant_name,
			created_at, updated_at
		FROM transactions
		WHERE user_id = $1
			AND date >= $2
			AND date <= $3
			AND amount < 0
		ORDER BY amount ASC
		LIMIT 1`

	err = r.query().QueryRowContext(
		ctx,
		query,
		filter.UserID,
		filter.StartDate,
		filter.EndDate,
	).Scan(
		&report.LargestExpense.ID,
		&report.LargestExpense.UserID,
		&report.LargestExpense.AccountID,
		&report.LargestExpense.CategoryID,
		&report.LargestExpense.Amount,
		&report.LargestExpense.Date,
		&report.LargestExpense.Description,
		&report.LargestExpense.MerchantName,
		&report.LargestExpense.CreatedAt,
		&report.LargestExpense.UpdatedAt,
	)

	if err != nil && err != sql.ErrNoRows {
		return nil, errors.New("Failed to get largest expense", http.StatusInternalServerError)
	}

	return report, nil
}

func (r *AnalyticsSQL) GetBudgetPerformance(ctx context.Context, filter model.AnalyticsFilter) (*model.BudgetPerformance, error) {
	// Query to get category-wise budget performance
	query := `
		WITH category_spending AS (
			SELECT 
				c.id as category_id,
				c.name as category_name,
				b.amount as budget_amount,
				COALESCE(ABS(SUM(CASE WHEN t.type = 'debit' THEN t.amount ELSE 0 END)), 0) as spent_amount
			FROM categories c
			LEFT JOIN budgets b ON b.category_id = c.id 
				AND b.user_id = $1 
				AND b.period_start <= $2 
				AND b.period_end >= $3
			LEFT JOIN transactions t ON t.category_id = c.id
				AND t.user_id = $1
				AND t.date >= $2
				AND t.date <= $3
			WHERE c.type = 'expense'
			GROUP BY c.id, c.name, b.amount
			HAVING b.amount > 0 OR COALESCE(ABS(SUM(CASE WHEN t.type = 'debit' THEN t.amount ELSE 0 END)), 0) > 0
		)
		SELECT 
			category_id,
			category_name,
			COALESCE(budget_amount, 0) as budget_amount,
			spent_amount,
			GREATEST(COALESCE(budget_amount, 0) - spent_amount, 0) as remaining_amount,
			CASE 
				WHEN COALESCE(budget_amount, 0) > 0 THEN 
					ROUND((spent_amount / budget_amount) * 100, 2)
				ELSE 100
			END as spending_progress
		FROM category_spending
		ORDER BY spent_amount DESC`

	rows, err := r.query().QueryContext(ctx, query, filter.UserID, filter.StartDate, filter.EndDate)
	if err != nil {
		return nil, errors.New("Failed to query budget performance", http.StatusInternalServerError)
	}
	defer rows.Close()

	var categories []model.CategoryBudgetPerformance
	var totalBudget, totalSpent float64

	for rows.Next() {
		var cat model.CategoryBudgetPerformance
		err := rows.Scan(
			&cat.CategoryID,
			&cat.CategoryName,
			&cat.BudgetAmount,
			&cat.SpentAmount,
			&cat.RemainingAmount,
			&cat.SpendingProgress,
		)
		if err != nil {
			return nil, errors.New("Failed to scan budget performance row", http.StatusInternalServerError)
		}
		categories = append(categories, cat)
		totalBudget += cat.BudgetAmount
		totalSpent += cat.SpentAmount
	}

	if err = rows.Err(); err != nil {
		return nil, errors.New("Error iterating budget performance rows", http.StatusInternalServerError)
	}

	performance := &model.BudgetPerformance{
		Period:          getPeriodName(filter.StartDate, filter.EndDate),
		TotalBudget:     totalBudget,
		TotalSpent:      totalSpent,
		RemainingBudget: totalBudget - totalSpent,
		SpendingProgress: func() float64 {
			if totalBudget > 0 {
				return (totalSpent / totalBudget) * 100
			}
			return 100
		}(),
		Categories: categories,
	}

	return performance, nil
}

func getPeriodName(start, end time.Time) string {
	if start.Year() != end.Year() {
		return "custom"
	}
	if start.Month() == end.Month() {
		return "month"
	}
	if end.Sub(start).Hours()/24 >= 89 && end.Sub(start).Hours()/24 <= 92 {
		return "quarter"
	}
	if start.Month() == time.January && end.Month() == time.December {
		return "year"
	}
	return "custom"
}

func (r *AnalyticsSQL) GetIncomeVsExpenses(ctx context.Context, filter model.AnalyticsFilter) (*model.IncomeVsExpenses, error) {
	sqlQuery := `
		WITH income_expenses AS (
			SELECT 
				COALESCE(SUM(CASE WHEN t.type = 'income' THEN amount ELSE 0 END), 0) as total_income,
				COALESCE(ABS(SUM(CASE WHEN t.type = 'expense' THEN amount ELSE 0 END)), 0) as total_expenses
			FROM transactions t
			WHERE t.user_id = $1
				AND t.date >= $2
				AND t.date <= $3
		)
		SELECT 
			total_income,
			total_expenses,
			(total_income - total_expenses) as net_amount
		FROM income_expenses`

	var result model.IncomeVsExpenses
	err := r.query().QueryRowContext(
		ctx,
		sqlQuery,
		filter.UserID,
		filter.StartDate,
		filter.EndDate,
	).Scan(&result.TotalIncome, &result.TotalExpenses, &result.NetAmount)

	if err != nil {
		if err == sql.ErrNoRows {
			return &model.IncomeVsExpenses{}, nil
		}
		return nil, errors.New("Failed to get income vs expenses", http.StatusInternalServerError)
	}

	return &result, nil
}

func (r *AnalyticsSQL) GetIncomeVsExpensesByUser(ctx context.Context, userID string) (*model.IncomeVsExpenses, error) {
	filter := model.AnalyticsFilter{
		UserID:    userID,
		StartDate: time.Now().AddDate(0, 0, -30), // Last 30 days
		EndDate:   time.Now(),
	}
	return r.GetIncomeVsExpenses(ctx, filter)
}

func (r *AnalyticsSQL) GetMonthlyDebtPayments(ctx context.Context) (float64, error) {
	var total float64
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE category_type = 'expense'
		AND category_name IN ('Debt Payment', 'Loan Payment', 'Credit Card Payment')
		AND date_trunc('month', transaction_date) = date_trunc('month', CURRENT_DATE)`
	
	err := r.db.QueryRowContext(ctx, query).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get monthly debt payments: %w", err)
	}
	return total, nil
}

func (r *AnalyticsSQL) GetMonthlyDebtPaymentsByUser(ctx context.Context, userID string) (float64, error) {
	var total float64
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE category_type = 'expense'
		AND category_name IN ('Debt Payment', 'Loan Payment', 'Credit Card Payment')
		AND date_trunc('month', transaction_date) = date_trunc('month', CURRENT_DATE)
		AND user_id = $1`
	
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get monthly debt payments for user: %w", err)
	}
	return total, nil
}

func (r *AnalyticsSQL) GetEmergencyFundBalance(ctx context.Context) (float64, error) {
	var balance float64
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM accounts
		WHERE account_type = 'Emergency Fund'`
	
	err := r.db.QueryRowContext(ctx, query).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to get emergency fund balance: %w", err)
	}
	return balance, nil
}

func (r *AnalyticsSQL) GetEmergencyFundBalanceByUser(ctx context.Context, userID string) (float64, error) {
	var balance float64
	query := `
		SELECT COALESCE(SUM(balance), 0)
		FROM accounts
		WHERE account_type = 'savings'
		AND is_emergency_fund = true
		AND user_id = $1`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to get emergency fund balance for user: %w", err)
	}
	return balance, nil
}

func (r *AnalyticsSQL) GetAverageMonthlyExpenses(ctx context.Context) (float64, error) {
	var average float64
	query := `
		SELECT COALESCE(AVG(monthly_total), 0)
		FROM (
			SELECT date_trunc('month', transaction_date) as month,
				   SUM(amount) as monthly_total
			FROM transactions
			WHERE category_type = 'expense'
			AND transaction_date >= date_trunc('month', CURRENT_DATE - INTERVAL '6 months')
			GROUP BY date_trunc('month', transaction_date)
		) monthly_expenses`
	
	err := r.db.QueryRowContext(ctx, query).Scan(&average)
	if err != nil {
		return 0, fmt.Errorf("failed to get average monthly expenses: %w", err)
	}
	return average, nil
}

func (r *AnalyticsSQL) GetAverageMonthlyExpensesByUser(ctx context.Context, userID string) (float64, error) {
	var average float64
	query := `
		WITH monthly_expenses AS (
			SELECT date_trunc('month', transaction_date) as month,
				   SUM(amount) as total_expenses
			FROM transactions
			WHERE type = 'expense'
			AND user_id = $1
			GROUP BY date_trunc('month', transaction_date)
		)
		SELECT COALESCE(AVG(total_expenses), 0)
		FROM monthly_expenses`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&average)
	if err != nil {
		return 0, fmt.Errorf("failed to get average monthly expenses for user: %w", err)
	}
	return average, nil
}

func (r *AnalyticsSQL) GetTotalIncome(ctx context.Context) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions t
		JOIN categories c ON t.category_id = c.id
		WHERE c.type = 'income'
	`

	var totalIncome float64
	err := r.query().QueryRowContext(ctx, query).Scan(&totalIncome)
	if err != nil {
		return 0, errors.New("failed to get total income", http.StatusInternalServerError)
	}

	return totalIncome, nil
}

func (r *AnalyticsSQL) GetTotalExpenses(ctx context.Context) (float64, error) {
	query := `
		SELECT COALESCE(ABS(SUM(amount)), 0)
		FROM transactions
		WHERE amount < 0
	`

	var total float64
	err := r.query().QueryRowContext(ctx, query).Scan(&total)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get total expenses", http.StatusInternalServerError)
	}

	return total, nil
}

func (r *AnalyticsSQL) GetAverageDailyExpenses(ctx context.Context) (float64, error) {
	query := `
		SELECT COALESCE(ABS(SUM(amount)) / NULLIF(DATE_PART('day', NOW() - MIN(date)), 0), 0)
		FROM transactions
		WHERE amount < 0
	`

	var avgDailyExpenses float64
	err := r.query().QueryRowContext(ctx, query).Scan(&avgDailyExpenses)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, errors.New("failed to get average daily expenses", http.StatusInternalServerError)
	}

	return avgDailyExpenses, nil
}

func (r *AnalyticsSQL) GetUserCount(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get user count: %w", err)
	}
	return count, nil
}
