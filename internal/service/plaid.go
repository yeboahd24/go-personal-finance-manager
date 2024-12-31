package service

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/plaid/plaid-go/v31/plaid"
)

type PlaidService struct {
	client *plaid.APIClient
}

func NewPlaidService() (*PlaidService, error) {
	clientID := os.Getenv("PLAID_CLIENT_ID")
	secret := os.Getenv("PLAID_SECRET")
	environment := os.Getenv("PLAID_ENVIRONMENT")

	if clientID == "" || secret == "" {
		log.Printf("Warning: Plaid credentials not found, some features will be disabled")
		return &PlaidService{}, nil
	}

	// Define environment mapping
	environments := map[string]plaid.Environment{
		"sandbox":    plaid.Sandbox,
		"production": plaid.Production,
	}

	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	configuration.AddDefaultHeader("PLAID-SECRET", secret)

	// Set environment from the mapping, default to sandbox if not found
	env, ok := environments[environment]
	if !ok {
		env = plaid.Sandbox
	}
	configuration.UseEnvironment(env)

	client := plaid.NewAPIClient(configuration)

	return &PlaidService{
		client: client,
	}, nil
}

func (s *PlaidService) CreateLinkToken(ctx context.Context, userID string) (string, error) {
	configs := plaid.LinkTokenCreateRequest{
		User: plaid.LinkTokenCreateRequestUser{
			ClientUserId: userID,
		},
		ClientName:   "Personal Finance Manager",
		Products:     []plaid.Products{plaid.PRODUCTS_TRANSACTIONS},
		CountryCodes: []plaid.CountryCode{plaid.COUNTRYCODE_US},
		Language:     "en",
	}

	resp, _, err := s.client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(configs).Execute()
	if err != nil {
		return "", err
	}

	return resp.LinkToken, nil
}

func (s *PlaidService) ExchangePublicToken(ctx context.Context, publicToken string) (string, string, error) {
	request := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	resp, _, err := s.client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(*request).Execute()
	if err != nil {
		return "", "", err
	}

	return resp.GetAccessToken(), resp.GetItemId(), nil
}

func (s *PlaidService) GetAccounts(ctx context.Context, accessToken string) ([]*plaid.AccountBase, error) {
	accountsGetResp, _, err := s.client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
		*plaid.NewAccountsGetRequest(accessToken),
	).Execute()
	if err != nil {
		return nil, err
	}

	accounts := accountsGetResp.GetAccounts()
	result := make([]*plaid.AccountBase, len(accounts))
	for i := range accounts {
		result[i] = &accounts[i]
	}
	return result, nil
}

func (s *PlaidService) GetTransactions(ctx context.Context, accessToken string, startDate, endDate string) ([]plaid.Transaction, error) {
	var allTransactions []plaid.Transaction
	cursor := ""
	hasMore := true

	for hasMore {
		request := plaid.NewTransactionsSyncRequest(accessToken)
		if cursor != "" {
			request.SetCursor(cursor)
		}

		resp, raw, err := s.client.PlaidApi.TransactionsSync(ctx).TransactionsSyncRequest(*request).Execute()
		if err != nil {
			// Print detailed error information
			if plaidErr, ok := err.(plaid.GenericOpenAPIError); ok {
				log.Printf("Plaid API error details: %+v", string(plaidErr.Body()))
			}
			log.Printf("Full error: %+v", err)
			if raw != nil {
				log.Printf("Raw response: %+v", raw)
			}
			return nil, fmt.Errorf("plaid sync error: %w", err)
		}

		allTransactions = append(allTransactions, resp.GetAdded()...)
		cursor = resp.GetNextCursor()
		hasMore = resp.GetHasMore()
	}

	return allTransactions, nil
}
