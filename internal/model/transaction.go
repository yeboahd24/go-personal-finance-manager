package model

import (
	"time"
)

type Transaction struct {
	ID                string    `json:"id"`
	UserID           string    `json:"user_id"`
	AccountID        string    `json:"account_id"`
	CategoryID       *string   `json:"category_id,omitempty"`
	Amount           float64   `json:"amount"`
	Description      string    `json:"description"`
	Date            time.Time `json:"date"`
	Type            string    `json:"type"` // credit or debit
	Status          string    `json:"status"`
	PlaidTransactionID *string   `json:"plaid_transaction_id,omitempty"`
	MerchantName    *string   `json:"merchant_name,omitempty"`
	Categories      []string  `json:"categories,omitempty"`
	Location        *TransactionLocation `json:"location,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Joined fields
	Category string `json:"category,omitempty"`
	Account  string `json:"account,omitempty"`
}

type TransactionLocation struct {
	Address     string  `json:"address,omitempty"`
	City        string  `json:"city,omitempty"`
	Region      string  `json:"region,omitempty"`
	PostalCode  string  `json:"postal_code,omitempty"`
	Country     string  `json:"country,omitempty"`
	Lat        float64 `json:"lat,omitempty"`
	Lon        float64 `json:"lon,omitempty"`
}

type TransactionFilter struct {
	UserID     string
	AccountID  string
	StartDate  time.Time
	EndDate    time.Time
	MinAmount  *float64
	MaxAmount  *float64
	Categories []string
	Type       string
	Search     string
	Limit      int
	Offset     int
}
