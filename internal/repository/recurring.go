package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/errors"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
)

type RecurringTransactionRepository interface {
	CreateRecurringTransaction(ctx context.Context, tx *model.RecurringTransaction) error
	GetRecurringTransactionByID(ctx context.Context, id string) (*model.RecurringTransaction, error)
	GetRecurringTransactions(ctx context.Context, filter model.RecurringTransactionFilter) ([]*model.RecurringTransaction, error)
	UpdateRecurringTransaction(ctx context.Context, tx *model.RecurringTransaction) error
	DeleteRecurringTransaction(ctx context.Context, id string) error
	GetDueRecurringTransactions(ctx context.Context, before time.Time) ([]*model.RecurringTransaction, error)
	UpdateLastRun(ctx context.Context, id string, lastRun, nextRun time.Time) error
}

type RecurringTransactionSQL struct {
	db *sql.DB
	tx *sql.Tx
}

func (r *RecurringTransactionSQL) query() QueryExecutor {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *RecurringTransactionSQL) CreateRecurringTransaction(ctx context.Context, tx *model.RecurringTransaction) error {
	query := `
		INSERT INTO recurring_transactions (
			user_id, account_id, category_id, amount, description,
			interval, day_of_month, day_of_week, start_date, end_date,
			last_run, next_run, active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		RETURNING id, created_at, updated_at`

	err := r.query().QueryRowContext(
		ctx,
		query,
		tx.UserID,
		tx.AccountID,
		tx.CategoryID,
		tx.Amount,
		tx.Description,
		tx.Interval,
		tx.DayOfMonth,
		tx.DayOfWeek,
		tx.StartDate,
		tx.EndDate,
		tx.LastRun,
		tx.NextRun,
		tx.Active,
	).Scan(&tx.ID, &tx.CreatedAt, &tx.UpdatedAt)

	if err != nil {
		return errors.Wrap(err, "Failed to create recurring transaction", 500)
	}
	return nil
}

func (r *RecurringTransactionSQL) GetRecurringTransactionByID(ctx context.Context, id string) (*model.RecurringTransaction, error) {
	tx := &model.RecurringTransaction{}
	query := `
		SELECT 
			rt.id, rt.user_id, rt.account_id, rt.category_id,
			rt.amount, rt.description, rt.interval,
			rt.day_of_month, rt.day_of_week,
			rt.start_date, rt.end_date, rt.last_run,
			rt.next_run, rt.active, rt.created_at, rt.updated_at,
			a.name as account_name, a.type as account_type,
			c.name as category_name, c.type as category_type
		FROM recurring_transactions rt
		LEFT JOIN accounts a ON rt.account_id = a.id
		LEFT JOIN categories c ON rt.category_id = c.id
		WHERE rt.id = $1`

	var account model.Account
	var category model.Category
	err := r.query().QueryRowContext(ctx, query, id).Scan(
		&tx.ID,
		&tx.UserID,
		&tx.AccountID,
		&tx.CategoryID,
		&tx.Amount,
		&tx.Description,
		&tx.Interval,
		&tx.DayOfMonth,
		&tx.DayOfWeek,
		&tx.StartDate,
		&tx.EndDate,
		&tx.LastRun,
		&tx.NextRun,
		&tx.Active,
		&tx.CreatedAt,
		&tx.UpdatedAt,
		&account.Name,
		&account.Type,
		&category.Name,
		&category.Type,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get recurring transaction", 500)
	}

	tx.Account = &account
	tx.Category = &category
	return tx, nil
}

func (r *RecurringTransactionSQL) GetRecurringTransactions(ctx context.Context, filter model.RecurringTransactionFilter) ([]*model.RecurringTransaction, error) {
	query := `
		SELECT 
			rt.id, rt.user_id, rt.account_id, rt.category_id,
			rt.amount, rt.description, rt.interval,
			rt.day_of_month, rt.day_of_week,
			rt.start_date, rt.end_date, rt.last_run,
			rt.next_run, rt.active, rt.created_at, rt.updated_at,
			a.name as account_name, a.type as account_type,
			c.name as category_name, c.type as category_type
		FROM recurring_transactions rt
		LEFT JOIN accounts a ON rt.account_id = a.id
		LEFT JOIN categories c ON rt.category_id = c.id
		WHERE rt.user_id = $1
			AND ($2::uuid IS NULL OR rt.account_id = $2)
			AND ($3::uuid IS NULL OR rt.category_id = $3)
			AND ($4::boolean IS NULL OR rt.active = $4)
		ORDER BY rt.next_run ASC`

	rows, err := r.query().QueryContext(
		ctx,
		query,
		filter.UserID,
		filter.AccountID,
		filter.CategoryID,
		filter.Active,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get recurring transactions", 500)
	}
	defer rows.Close()

	var transactions []*model.RecurringTransaction
	for rows.Next() {
		tx := &model.RecurringTransaction{}
		account := &model.Account{}
		category := &model.Category{}

		err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.AccountID,
			&tx.CategoryID,
			&tx.Amount,
			&tx.Description,
			&tx.Interval,
			&tx.DayOfMonth,
			&tx.DayOfWeek,
			&tx.StartDate,
			&tx.EndDate,
			&tx.LastRun,
			&tx.NextRun,
			&tx.Active,
			&tx.CreatedAt,
			&tx.UpdatedAt,
			&account.Name,
			&account.Type,
			&category.Name,
			&category.Type,
		)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to scan recurring transaction", 500)
		}

		tx.Account = account
		tx.Category = category
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (r *RecurringTransactionSQL) UpdateRecurringTransaction(ctx context.Context, tx *model.RecurringTransaction) error {
	query := `
		UPDATE recurring_transactions
		SET account_id = $2,
			category_id = $3,
			amount = $4,
			description = $5,
			interval = $6,
			day_of_month = $7,
			day_of_week = $8,
			start_date = $9,
			end_date = $10,
			active = $11,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	err := r.query().QueryRowContext(
		ctx,
		query,
		tx.ID,
		tx.AccountID,
		tx.CategoryID,
		tx.Amount,
		tx.Description,
		tx.Interval,
		tx.DayOfMonth,
		tx.DayOfWeek,
		tx.StartDate,
		tx.EndDate,
		tx.Active,
	).Scan(&tx.UpdatedAt)

	if err == sql.ErrNoRows {
		return errors.ErrNotFound
	}
	if err != nil {
		return errors.Wrap(err, "Failed to update recurring transaction", 500)
	}
	return nil
}

func (r *RecurringTransactionSQL) DeleteRecurringTransaction(ctx context.Context, id string) error {
	query := "DELETE FROM recurring_transactions WHERE id = $1"
	result, err := r.query().ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "Failed to delete recurring transaction", 500)
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

func (r *RecurringTransactionSQL) GetDueRecurringTransactions(ctx context.Context, before time.Time) ([]*model.RecurringTransaction, error) {
	query := `
		SELECT 
			rt.id, rt.user_id, rt.account_id, rt.category_id,
			rt.amount, rt.description, rt.interval,
			rt.day_of_month, rt.day_of_week,
			rt.start_date, rt.end_date, rt.last_run,
			rt.next_run, rt.active, rt.created_at, rt.updated_at,
			a.name as account_name, a.type as account_type,
			c.name as category_name, c.type as category_type
		FROM recurring_transactions rt
		LEFT JOIN accounts a ON rt.account_id = a.id
		LEFT JOIN categories c ON rt.category_id = c.id
		WHERE rt.active = true
			AND rt.next_run <= $1
			AND (rt.end_date IS NULL OR rt.next_run <= rt.end_date)
		ORDER BY rt.next_run ASC`

	rows, err := r.query().QueryContext(ctx, query, before)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get due recurring transactions", 500)
	}
	defer rows.Close()

	var transactions []*model.RecurringTransaction
	for rows.Next() {
		tx := &model.RecurringTransaction{}
		account := &model.Account{}
		category := &model.Category{}

		err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.AccountID,
			&tx.CategoryID,
			&tx.Amount,
			&tx.Description,
			&tx.Interval,
			&tx.DayOfMonth,
			&tx.DayOfWeek,
			&tx.StartDate,
			&tx.EndDate,
			&tx.LastRun,
			&tx.NextRun,
			&tx.Active,
			&tx.CreatedAt,
			&tx.UpdatedAt,
			&account.Name,
			&account.Type,
			&category.Name,
			&category.Type,
		)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to scan recurring transaction", 500)
		}

		tx.Account = account
		tx.Category = category
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (r *RecurringTransactionSQL) UpdateLastRun(ctx context.Context, id string, lastRun, nextRun time.Time) error {
	query := `
		UPDATE recurring_transactions
		SET last_run = $2,
			next_run = $3,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.query().ExecContext(ctx, query, id, lastRun, nextRun)
	if err != nil {
		return errors.Wrap(err, "Failed to update last run", 500)
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
