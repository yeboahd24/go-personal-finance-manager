package service

import (
	"context"
	"fmt"

	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"github.com/yeboahd24/personal-finance-manager/internal/repository"
)

type AccountService struct {
	repo  repository.Repository
	plaid *PlaidService
}

func NewAccountService(repo repository.Repository, plaid *PlaidService) *AccountService {
	return &AccountService{
		repo:  repo,
		plaid: plaid,
	}
}

func (s *AccountService) CreateLinkToken(ctx context.Context, userID string) (string, error) {
	return s.plaid.CreateLinkToken(ctx, userID)
}

func (s *AccountService) LinkAccount(ctx context.Context, userID string, publicToken string) error {
	// Exchange public token for access token
	accessToken, itemID, err := s.plaid.ExchangePublicToken(ctx, publicToken)
	if err != nil {
		return fmt.Errorf("failed to exchange public token: %w", err)
	}

	// Save Plaid credentials
	creds := &model.PlaidCredentials{
		UserID:      userID,
		AccessToken: accessToken,
		ItemID:      itemID,
	}
	if err := s.repo.SavePlaidCredentials(ctx, creds); err != nil {
		return fmt.Errorf("failed to save plaid credentials: %w", err)
	}

	// Get accounts from Plaid
	plaidAccounts, err := s.plaid.GetAccounts(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("failed to get accounts from plaid: %w", err)
	}

	// Save accounts to database
	for _, plaidAccount := range plaidAccounts {
		account := &model.Account{
			UserID:         userID,
			PlaidAccountID: plaidAccount.GetAccountId(),
			Name:           plaidAccount.GetName(),
			Type:           string(plaidAccount.GetType()),
			Balance:        *plaidAccount.GetBalances().Current.Get(),
			Currency:       *plaidAccount.GetBalances().IsoCurrencyCode.Get(),
		}
		if err := s.repo.CreateAccount(ctx, account); err != nil {
			return fmt.Errorf("failed to create account: %w", err)
		}
	}

	return nil
}

func (s *AccountService) GetAccounts(ctx context.Context, userID string) ([]*model.Account, error) {
	return s.repo.GetAccountsByUserID(ctx, userID)
}

func (s *AccountService) CreateAccount(ctx context.Context, account *model.Account) error {
	return s.repo.CreateAccount(ctx, account)
}

func (s *AccountService) UpdateAccount(ctx context.Context, account *model.Account) error {
	return s.repo.UpdateAccount(ctx, account)
}

func (s *AccountService) GetAccountByID(ctx context.Context, id string) (*model.Account, error) {
	return s.repo.GetAccountByID(ctx, id)
}

func (s *AccountService) SyncAccounts(ctx context.Context, userID string) error {
	// Get Plaid credentials
	creds, err := s.repo.GetPlaidCredentials(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get plaid credentials: %w", err)
	}

	// Get updated account information from Plaid
	plaidAccounts, err := s.plaid.GetAccounts(ctx, creds.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get accounts from plaid: %w", err)
	}

	// Update account balances
	for _, plaidAccount := range plaidAccounts {
		accounts, err := s.repo.GetAccountsByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get accounts: %w", err)
		}

		for _, account := range accounts {
			if account.PlaidAccountID == plaidAccount.AccountId {
				account.Balance = *plaidAccount.GetBalances().Current.Get()
				if err := s.repo.UpdateAccount(ctx, account); err != nil {
					return fmt.Errorf("failed to update account: %w", err)
				}
			}
		}
	}

	return nil
}
