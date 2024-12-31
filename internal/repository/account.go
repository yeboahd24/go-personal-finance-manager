package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"log"
)

type AccountRepository interface {
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
}

type AccountSQL struct {
	db *sql.DB
	tx *sql.Tx
}

func (r *AccountSQL) query() QueryExecutor {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *AccountSQL) CreateAccount(ctx context.Context, account *model.Account) error {
	log.Printf("Starting CreateAccount in repository for user %s", account.UserID)
    
    // Generate a new UUID for the account ID
    if account.ID == "" {
        account.ID = uuid.New().String()
    }

    query := `
        INSERT INTO accounts (id, user_id, plaid_account_id, name, type, balance, currency, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING created_at, updated_at`

    log.Printf("Generated UUID for account: %s", account.ID)
    log.Printf("Executing query with values: id=%s, user_id=%s, name=%s, type=%s, balance=%f, currency=%s",
        account.ID, account.UserID, account.Name, account.Type, account.Balance, account.Currency)

    var err error
    if r.tx != nil {
        // If we already have a transaction, use it
        err = r.executeAccountCreation(ctx, r.tx, query, account)
    } else {
        // Start a new transaction
        tx, err := r.db.BeginTx(ctx, nil)
        if err != nil {
            log.Printf("Error starting transaction: %v", err)
            return fmt.Errorf("failed to start transaction: %w", err)
        }
        defer tx.Rollback() // Rollback if we don't commit

        if err = r.executeAccountCreation(ctx, tx, query, account); err != nil {
            return err
        }

        // Commit the transaction
        if err = tx.Commit(); err != nil {
            log.Printf("Error committing transaction: %v", err)
            return fmt.Errorf("failed to commit transaction: %w", err)
        }
    }

    if err != nil {
        log.Printf("Error creating account: %v", err)
        return fmt.Errorf("failed to create account: %w", err)
    }

    log.Printf("Successfully created account in database with ID: %s", account.ID)
    return nil
}

func (r *AccountSQL) executeAccountCreation(ctx context.Context, tx *sql.Tx, query string, account *model.Account) error {
    // Execute the query within the transaction
    err := tx.QueryRowContext(
        ctx,
        query,
        account.ID,
        account.UserID,
        account.PlaidAccountID,
        account.Name,
        account.Type,
        account.Balance,
        account.Currency,
    ).Scan(&account.CreatedAt, &account.UpdatedAt)

    if err != nil {
        log.Printf("Error executing query: %v", err)
        return fmt.Errorf("failed to create account: %w", err)
    }

    // Verify the account was created
    var count int
    err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM accounts WHERE id = $1", account.ID).Scan(&count)
    if err != nil {
        log.Printf("Error verifying account creation: %v", err)
        return fmt.Errorf("failed to verify account creation: %w", err)
    }

    if count != 1 {
        log.Printf("Account was not created. Count = %d", count)
        return fmt.Errorf("account was not created")
    }

    return nil
}

func (r *AccountSQL) GetAccountByID(ctx context.Context, id string) (*model.Account, error) {
	account := &model.Account{}
	query := `
		SELECT id, user_id, plaid_account_id, name, type, balance, currency, created_at, updated_at
		FROM accounts
		WHERE id = $1`

	err := r.query().QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.UserID,
		&account.PlaidAccountID,
		&account.Name,
		&account.Type,
		&account.Balance,
		&account.Currency,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *AccountSQL) GetAccountsByUserID(ctx context.Context, userID string) ([]*model.Account, error) {
    log.Printf("Getting accounts for user: %s", userID)

    query := `
        SELECT id, user_id, plaid_account_id, name, type, balance, currency, created_at, updated_at
        FROM accounts
        WHERE user_id = $1
        ORDER BY created_at DESC`

    // Execute query directly without transaction since we're just reading
    rows, err := r.db.QueryContext(ctx, query, userID)
    if err != nil {
        log.Printf("Error querying accounts: %v", err)
        return nil, fmt.Errorf("failed to query accounts: %w", err)
    }
    defer rows.Close()

    var accounts []*model.Account
    for rows.Next() {
        account := &model.Account{}
        err := rows.Scan(
            &account.ID,
            &account.UserID,
            &account.PlaidAccountID,
            &account.Name,
            &account.Type,
            &account.Balance,
            &account.Currency,
            &account.CreatedAt,
            &account.UpdatedAt,
        )
        if err != nil {
            log.Printf("Error scanning account row: %v", err)
            return nil, fmt.Errorf("failed to scan account: %w", err)
        }
        accounts = append(accounts, account)
    }

    if err = rows.Err(); err != nil {
        log.Printf("Error iterating accounts: %v", err)
        return nil, fmt.Errorf("error iterating accounts: %w", err)
    }

    // Return empty slice instead of nil if no accounts found
    if accounts == nil {
        log.Printf("No accounts found for user %s", userID)
        accounts = []*model.Account{}
    } else {
        log.Printf("Found %d accounts for user %s", len(accounts), userID)
    }

    return accounts, nil
}

func (r *AccountSQL) UpdateAccount(ctx context.Context, account *model.Account) error {
	query := `
		UPDATE accounts
		SET balance = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	return r.query().QueryRowContext(
		ctx,
		query,
		account.ID,
		account.Balance,
	).Scan(&account.UpdatedAt)
}

func (r *AccountSQL) SavePlaidCredentials(ctx context.Context, creds *model.PlaidCredentials) error {
	query := `
		INSERT INTO plaid_credentials (user_id, access_token, item_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE
		SET access_token = EXCLUDED.access_token,
			item_id = EXCLUDED.item_id,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at`

	return r.query().QueryRowContext(
		ctx,
		query,
		creds.UserID,
		creds.AccessToken,
		creds.ItemID,
	).Scan(&creds.ID, &creds.CreatedAt, &creds.UpdatedAt)
}

func (r *AccountSQL) GetPlaidCredentials(ctx context.Context, userID string) (*model.PlaidCredentials, error) {
	creds := &model.PlaidCredentials{}
	query := `
		SELECT id, user_id, access_token, item_id, created_at, updated_at
		FROM plaid_credentials
		WHERE user_id = $1`

	err := r.query().QueryRowContext(ctx, query, userID).Scan(
		&creds.ID,
		&creds.UserID,
		&creds.AccessToken,
		&creds.ItemID,
		&creds.CreatedAt,
		&creds.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return creds, nil
}

// GetTotalAssets returns the sum of all account balances that are assets
func (r *AccountSQL) GetTotalAssets(ctx context.Context) (float64, error) {
	query := `
		SELECT COALESCE(SUM(balance), 0)
		FROM accounts
		WHERE type IN ('checking', 'savings', 'investment')`
	
	var total float64
	err := r.query().QueryRowContext(ctx, query).Scan(&total)
	if err != nil {
		return 0, err
	}
	
	return total, nil
}

// GetTotalAssetsByUser returns the sum of all account balances that are assets for a specific user
func (r *AccountSQL) GetTotalAssetsByUser(ctx context.Context, userID string) (float64, error) {
	var total float64
	query := `
		SELECT COALESCE(SUM(balance), 0)
		FROM accounts
		WHERE type IN ('checking', 'savings', 'investment')
		AND user_id = $1`
	
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total assets for user: %w", err)
	}
	return total, nil
}

// GetTotalLiabilities returns the sum of all account balances that are liabilities
func (r *AccountSQL) GetTotalLiabilities(ctx context.Context) (float64, error) {
	var total float64
	query := `
		SELECT COALESCE(SUM(balance), 0)
		FROM accounts
		WHERE type IN ('credit', 'loan', 'mortgage')`

	err := r.query().QueryRowContext(ctx, query).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// GetTotalLiabilitiesByUser returns the sum of all account balances that are liabilities for a specific user
func (r *AccountSQL) GetTotalLiabilitiesByUser(ctx context.Context, userID string) (float64, error) {
	var total float64
	query := `
		SELECT COALESCE(SUM(balance), 0)
		FROM accounts
		WHERE type IN ('credit', 'loan', 'mortgage')
		AND user_id = $1`
	
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total liabilities for user: %w", err)
	}
	return total, nil
}
