package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/errors"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
)

type BudgetRepository interface {
	CreateBudget(ctx context.Context, budget *model.Budget) error
	GetBudgetByID(ctx context.Context, id string) (*model.Budget, error)
	GetBudgets(ctx context.Context, filter model.BudgetFilter) ([]*model.Budget, error)
	UpdateBudget(ctx context.Context, budget *model.Budget) error
	DeleteBudget(ctx context.Context, id string) error
	GetBudgetSummary(ctx context.Context, userID string, start, end time.Time) (*model.BudgetSummary, error)
}

type BudgetSQL struct {
	db *sql.DB
	tx *sql.Tx
}

func (r *BudgetSQL) query() QueryExecutor {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *BudgetSQL) CreateBudget(ctx context.Context, budget *model.Budget) error {
	query := `
		INSERT INTO budgets (user_id, category_id, amount, period_start, period_end)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := r.query().QueryRowContext(
		ctx,
		query,
		budget.UserID,
		budget.CategoryID,
		budget.Amount,
		budget.PeriodStart,
		budget.PeriodEnd,
	).Scan(&budget.ID, &budget.CreatedAt, &budget.UpdatedAt)

	if err != nil {
		return errors.Wrap(err, "Failed to create budget", 500)
	}
	return nil
}

func (r *BudgetSQL) GetBudgetByID(ctx context.Context, id string) (*model.Budget, error) {
	budget := &model.Budget{}
	query := `
		SELECT 
			b.id, b.user_id, b.category_id, b.amount, b.period_start, b.period_end,
			b.created_at, b.updated_at,
			c.id, c.name, c.type, c.icon, c.color, c.parent_id,
			COALESCE(SUM(t.amount), 0) as spent_amount
		FROM budgets b
		LEFT JOIN categories c ON b.category_id = c.id
		LEFT JOIN transactions t ON t.category_id = c.id 
			AND t.date >= b.period_start 
			AND t.date <= b.period_end
		WHERE b.id = $1
		GROUP BY b.id, c.id`

	var category model.Category
	var spentAmount float64
	err := r.query().QueryRowContext(ctx, query, id).Scan(
		&budget.ID,
		&budget.UserID,
		&budget.CategoryID,
		&budget.Amount,
		&budget.PeriodStart,
		&budget.PeriodEnd,
		&budget.CreatedAt,
		&budget.UpdatedAt,
		&category.ID,
		&category.Name,
		&category.Type,
		&category.Icon,
		&category.Color,
		&category.ParentID,
		&spentAmount,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get budget", 500)
	}

	budget.Category = &category
	budget.SpentAmount = &spentAmount
	spentPercent := (spentAmount / budget.Amount) * 100
	budget.SpentPercent = &spentPercent

	return budget, nil
}

func (r *BudgetSQL) GetBudgets(ctx context.Context, filter model.BudgetFilter) ([]*model.Budget, error) {
	query := `
		SELECT 
			b.id, b.user_id, b.category_id, b.amount, b.period_start, b.period_end,
			b.created_at, b.updated_at,
			c.id, c.name, c.type, c.icon, c.color, c.parent_id,
			COALESCE(SUM(t.amount), 0) as spent_amount
		FROM budgets b
		LEFT JOIN categories c ON b.category_id = c.id
		LEFT JOIN transactions t ON t.category_id = c.id 
			AND t.date >= b.period_start 
			AND t.date <= b.period_end
		WHERE b.user_id = $1
			AND ($2::uuid IS NULL OR b.category_id = $2)
			AND ($3::timestamp IS NULL OR b.period_end >= $3)
			AND ($4::timestamp IS NULL OR b.period_start <= $4)
		GROUP BY b.id, c.id
		ORDER BY b.period_start DESC, c.name`

	rows, err := r.query().QueryContext(
		ctx,
		query,
		filter.UserID,
		filter.CategoryID,
		filter.PeriodStart,
		filter.PeriodEnd,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get budgets", 500)
	}
	defer rows.Close()

	var budgets []*model.Budget
	for rows.Next() {
		budget := &model.Budget{}
		category := &model.Category{}
		var spentAmount float64

		err := rows.Scan(
			&budget.ID,
			&budget.UserID,
			&budget.CategoryID,
			&budget.Amount,
			&budget.PeriodStart,
			&budget.PeriodEnd,
			&budget.CreatedAt,
			&budget.UpdatedAt,
			&category.ID,
			&category.Name,
			&category.Type,
			&category.Icon,
			&category.Color,
			&category.ParentID,
			&spentAmount,
		)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to scan budget", 500)
		}

		budget.Category = category
		budget.SpentAmount = &spentAmount
		spentPercent := (spentAmount / budget.Amount) * 100
		budget.SpentPercent = &spentPercent

		budgets = append(budgets, budget)
	}

	return budgets, nil
}

func (r *BudgetSQL) UpdateBudget(ctx context.Context, budget *model.Budget) error {
	query := `
		UPDATE budgets
		SET amount = $2, period_start = $3, period_end = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	err := r.query().QueryRowContext(
		ctx,
		query,
		budget.ID,
		budget.Amount,
		budget.PeriodStart,
		budget.PeriodEnd,
	).Scan(&budget.UpdatedAt)

	if err == sql.ErrNoRows {
		return errors.ErrNotFound
	}
	if err != nil {
		return errors.Wrap(err, "Failed to update budget", 500)
	}
	return nil
}

func (r *BudgetSQL) DeleteBudget(ctx context.Context, id string) error {
	query := "DELETE FROM budgets WHERE id = $1"
	result, err := r.query().ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "Failed to delete budget", 500)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "Failed to get rows affected", 500)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *BudgetSQL) GetBudgetSummary(ctx context.Context, userID string, start, end time.Time) (*model.BudgetSummary, error) {
	query := `
		WITH budget_summary AS (
			SELECT 
				COALESCE(SUM(b.amount), 0) as total_budget,
				COALESCE(SUM(
					CASE 
						WHEN t.amount > 0 THEN t.amount 
						ELSE 0 
					END
				), 0) as total_spent
			FROM budgets b
			LEFT JOIN transactions t ON t.category_id = b.category_id 
				AND t.date >= b.period_start 
				AND t.date <= b.period_end
			WHERE b.user_id = $1
				AND b.period_start >= $2
				AND b.period_end <= $3
		)
		SELECT 
			total_budget,
			total_spent,
			total_budget - total_spent as remaining_budget,
			CASE 
				WHEN total_budget > 0 THEN (total_spent / total_budget) * 100
				ELSE 0
			END as spent_percent
		FROM budget_summary`

	summary := &model.BudgetSummary{}
	err := r.query().QueryRowContext(ctx, query, userID, start, end).Scan(
		&summary.TotalBudget,
		&summary.TotalSpent,
		&summary.RemainingBudget,
		&summary.SpentPercent,
	)

	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "Failed to get budget summary", 500)
	}

	return summary, nil
}
