package model

import (
	"time"
)

type Account struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	PlaidAccountID string    `json:"plaid_account_id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Balance        float64   `json:"balance"`
	Currency       string    `json:"currency"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PlaidCredentials struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	AccessToken string    `json:"-"`
	ItemID      string    `json:"item_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
