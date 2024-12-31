package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/errors"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"github.com/yeboahd24/personal-finance-manager/internal/repository"
)

type TransactionService struct {
	repo  repository.Repository
	plaid *PlaidService // Optional Plaid integration
}

func NewTransactionService(repo repository.Repository, plaid *PlaidService) *TransactionService {
	return &TransactionService{
		repo:  repo,
		plaid: plaid,
	}
}

func (s *TransactionService) SyncTransactions(ctx context.Context, userID string) error {
	if s.plaid == nil {
		return errors.New("Plaid service not configured", 501)
	}

	// Get Plaid credentials
	creds, err := s.repo.GetPlaidCredentials(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "Failed to get Plaid credentials", 500)
	}

	// Get transactions from Plaid
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, -1, 0).Format("2006-01-02") // Last 30 days

	transactions, err := s.plaid.GetTransactions(ctx, creds.AccessToken, startDate, endDate)
	if err != nil {
		return errors.Wrap(err, "Failed to get transactions from Plaid", 500)
	}

	// Get user's accounts
	accounts, err := s.repo.GetAccountsByUserID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "Failed to get accounts", 500)
	}

	accountMap := make(map[string]string) // plaidAccountID -> accountID
	for _, account := range accounts {
		accountMap[account.PlaidAccountID] = account.ID
	}

	// Save transactions
	for _, plaidTx := range transactions {
		accountID, ok := accountMap[plaidTx.AccountId]
		if !ok {
			continue // Skip transactions for unlinked accounts
		}

		date, err := time.Parse("2006-01-02", plaidTx.Date)
		if err != nil {
			continue
		}

		tx := &model.Transaction{
			AccountID:          accountID,
			Amount:             plaidTx.Amount,
			Description:        plaidTx.Name,
			Date:               date,
			Type:               s.determineTransactionType(plaidTx.Amount),
			PlaidTransactionID: &plaidTx.TransactionId,
			MerchantName:       plaidTx.MerchantName.Get(),
			Categories:         plaidTx.Category,
			UserID:             userID,
		}

		// Try to create transaction, ignore if it already exists
		_ = s.repo.CreateTransaction(ctx, tx)
	}

	return nil
}

func (s *TransactionService) GetTransactions(ctx context.Context, userID string, filter model.TransactionFilter) ([]*model.Transaction, error) {
	// Always set the user_id in the filter
	filter.UserID = userID

	transactions, err := s.repo.GetTransactions(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*model.Transaction{}, nil
		}
		return nil, fmt.Errorf("transaction service - GetTransactions: %w", err)
	}
	
	if transactions == nil {
		return []*model.Transaction{}, nil
	}

	return transactions, nil
}

func (s *TransactionService) UpdateTransaction(ctx context.Context, userID string, tx *model.Transaction) error {
	// Verify transaction belongs to user
	existing, err := s.repo.GetTransactionByID(ctx, tx.ID)
	if err != nil {
		return err
	}

	account, err := s.repo.GetAccountByID(ctx, existing.AccountID)
	if err != nil {
		return err
	}

	if account.UserID != userID {
		return errors.ErrUnauthorized
	}

	// Only allow updating certain fields
	existing.CategoryID = tx.CategoryID
	existing.Description = tx.Description

	return s.repo.UpdateTransaction(ctx, existing)
}

func (s *TransactionService) CreateTransaction(ctx context.Context, userID string, tx *model.Transaction) error {
	// Verify account belongs to user
	account, err := s.repo.GetAccountByID(ctx, tx.AccountID)
	if err != nil {
		log.Printf("Error getting account %s: %+v", tx.AccountID, err)
		return fmt.Errorf("failed to get account: %w", err)
	}
	if account.UserID != userID {
		log.Printf("Account %s does not belong to user %s (belongs to %s)", tx.AccountID, userID, account.UserID)
		return errors.ErrUnauthorized
	}

	// Set the transaction type based on amount
	tx.Type = s.determineTransactionType(tx.Amount)

	// Set the user ID
	tx.UserID = userID

	log.Printf("Creating transaction: %+v", tx)
	if err := s.repo.CreateTransaction(ctx, tx); err != nil {
		log.Printf("Error in repository.CreateTransaction: %+v", err)
		return fmt.Errorf("failed to create transaction in repository: %w", err)
	}

	return nil
}

func (s *TransactionService) determineTransactionType(amount float64) string {
	if amount >= 0 {
		return "credit"
	}
	return "debit"
}

func (s *TransactionService) GetUserAccounts(ctx context.Context, userID string) ([]*model.Account, error) {
	return s.repo.GetAccountsByUserID(ctx, userID)
}

func (s *TransactionService) GetCategories(ctx context.Context, userID string) ([]*model.Category, error) {
	// First try to get existing categories
	categories, err := s.repo.GetCategories(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get categories", 500)
	}

	// If no categories exist, initialize default ones
	if len(categories) == 0 {
		if err := s.repo.InitializeDefaultCategories(ctx, userID); err != nil {
			return nil, errors.Wrap(err, "Failed to initialize default categories", 500)
		}
		// Get the newly created categories
		categories, err = s.repo.GetCategories(ctx, userID)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to get categories after initialization", 500)
		}
	}

	return categories, nil
}

func (s *TransactionService) GetRecentTransactions(ctx context.Context, userID string, limit int) ([]*model.Transaction, error) {
	if limit <= 0 {
		limit = 5 // default to 5 recent transactions
	}

	filter := model.TransactionFilter{
		UserID: userID,
		Limit:  limit,
	}

	transactions, err := s.repo.GetTransactions(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("transaction service - GetRecentTransactions: %w", err)
	}

	if transactions == nil {
		return []*model.Transaction{}, nil
	}

	return transactions, nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
