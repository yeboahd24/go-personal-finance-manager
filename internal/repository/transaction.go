package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/yeboahd24/personal-finance-manager/internal/errors"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
)

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, tx *model.Transaction) error
	GetTransactionByID(ctx context.Context, id string) (*model.Transaction, error)
	GetTransactions(ctx context.Context, filter model.TransactionFilter) ([]*model.Transaction, error)
	UpdateTransaction(ctx context.Context, tx *model.Transaction) error
	DeleteTransaction(ctx context.Context, id string) error
	GetAccountsByUserID(ctx context.Context, userID string) ([]*model.Account, error)
	GetCategoriesByUserID(ctx context.Context, userID string) ([]*model.Category, error)
}

type TransactionSQL struct {
	db *sql.DB
	tx *sql.Tx
}

func (r *TransactionSQL) query() QueryExecutor {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *TransactionSQL) CreateTransaction(ctx context.Context, tx *model.Transaction) error {
	query := `
		INSERT INTO transactions (
			user_id, account_id, category_id, amount, description, 
			date, type, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	log.Printf("Executing query: %s with values: user_id=%s, account_id=%s, category_id=%v, amount=%f, description=%s, date=%v, type=%s",
		query, tx.UserID, tx.AccountID, tx.CategoryID, tx.Amount, tx.Description, tx.Date, tx.Type)

	err := r.query().QueryRowContext(
		ctx,
		query,
		tx.UserID,
		tx.AccountID,
		tx.CategoryID,
		tx.Amount,
		tx.Description,
		tx.Date,
		tx.Type,
		"completed", // default status
	).Scan(&tx.ID, &tx.CreatedAt, &tx.UpdatedAt)

	if err != nil {
		log.Printf("Error executing transaction insert query: %+v", err)
		return fmt.Errorf("failed to create transaction in database: %w", err)
	}

	return nil
}

func (r *TransactionSQL) GetTransactionByID(ctx context.Context, id string) (*model.Transaction, error) {
	tx := &model.Transaction{}
	query := `
		SELECT 
			id, user_id, account_id, category_id, amount, description, date, type,
			plaid_transaction_id, merchant_name, categories, location,
			created_at, updated_at
		FROM transactions
		WHERE id = $1`

	err := r.query().QueryRowContext(ctx, query, id).Scan(
		&tx.ID,
		&tx.UserID,
		&tx.AccountID,
		&tx.CategoryID,
		&tx.Amount,
		&tx.Description,
		&tx.Date,
		&tx.Type,
		&tx.PlaidTransactionID,
		&tx.MerchantName,
		&tx.Categories,
		&tx.Location,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get transaction", 500)
	}
	return tx, nil
}

func (r *TransactionSQL) GetTransactions(ctx context.Context, filter model.TransactionFilter) ([]*model.Transaction, error) {
	var conditions []string
	var args []interface{}
	argCount := 1

	// Always filter by user_id if provided
	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("t.user_id::text = $%d::text", argCount))
		args = append(args, filter.UserID)
		argCount++
	}

	if filter.AccountID != "" {
		conditions = append(conditions, fmt.Sprintf("t.account_id = $%d", argCount))
		args = append(args, filter.AccountID)
		argCount++
	}

	if !filter.StartDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("t.date >= $%d", argCount))
		args = append(args, filter.StartDate)
		argCount++
	}

	if !filter.EndDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("t.date <= $%d", argCount))
		args = append(args, filter.EndDate)
		argCount++
	}

	if filter.MinAmount != nil {
		conditions = append(conditions, fmt.Sprintf("t.amount >= $%d", argCount))
		args = append(args, *filter.MinAmount)
		argCount++
	}

	if filter.MaxAmount != nil {
		conditions = append(conditions, fmt.Sprintf("t.amount <= $%d", argCount))
		args = append(args, *filter.MaxAmount)
		argCount++
	}

	if filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("t.type = $%d", argCount))
		args = append(args, filter.Type)
		argCount++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("t.description ILIKE $%d", argCount))
		args = append(args, "%"+filter.Search+"%")
		argCount++
	}

	query := `
		SELECT 
			t.id, t.user_id, t.account_id, t.category_id, t.amount, t.description, 
			t.date, t.type, t.status, t.created_at, t.updated_at,
			c.name as category_name,
			a.name as account_name
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id
		LEFT JOIN accounts a ON t.account_id = a.id`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY t.date DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := r.query().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get transactions", 500)
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		tx := &model.Transaction{}
		var categoryName, accountName sql.NullString
		var status string
		err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.AccountID,
			&tx.CategoryID,
			&tx.Amount,
			&tx.Description,
			&tx.Date,
			&tx.Type,
			&status,
			&tx.CreatedAt,
			&tx.UpdatedAt,
			&categoryName,
			&accountName,
		)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to scan transaction", 500)
		}
		tx.Category = categoryName.String
		tx.Account = accountName.String
		tx.Status = status
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (r *TransactionSQL) UpdateTransaction(ctx context.Context, tx *model.Transaction) error {
	query := `
		UPDATE transactions
		SET 
			category_id = $2,
			description = $3,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	err := r.query().QueryRowContext(
		ctx,
		query,
		tx.ID,
		tx.CategoryID,
		tx.Description,
	).Scan(&tx.UpdatedAt)

	if err == sql.ErrNoRows {
		return errors.ErrNotFound
	}
	if err != nil {
		return errors.Wrap(err, "Failed to update transaction", 500)
	}
	return nil
}

func (r *TransactionSQL) DeleteTransaction(ctx context.Context, id string) error {
	query := "DELETE FROM transactions WHERE id = $1"

	result, err := r.query().ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "Failed to delete transaction", 500)
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

func (r *TransactionSQL) GetAccountsByUserID(ctx context.Context, userID string) ([]*model.Account, error) {
	query := `
		SELECT id, user_id, name, type, balance, currency, plaid_account_id, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY name ASC`

	rows, err := r.query().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*model.Account
	for rows.Next() {
		var account model.Account
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Name,
			&account.Type,
			&account.Balance,
			&account.Currency,
			&account.PlaidAccountID,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, &account)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating accounts: %w", err)
	}

	return accounts, nil
}

func (r *TransactionSQL) GetCategoriesByUserID(ctx context.Context, userID string) ([]*model.Category, error) {
	query := `
		SELECT id, user_id, name, created_at, updated_at
		FROM categories
		WHERE user_id = $1 OR user_id IS NULL
		ORDER BY name ASC`

	rows, err := r.query().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []*model.Category
	for rows.Next() {
		var category model.Category
		err := rows.Scan(
			&category.ID,
			&category.UserID,
			&category.Name,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}
