package model

import (
	"time"
)

type Category struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // income, expense, transfer
	Icon      string    `json:"icon,omitempty"`
	Color     string    `json:"color,omitempty"`
	ParentID  *string   `json:"parent_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Default category types
const (
	CategoryTypeIncome   = "income"
	CategoryTypeExpense  = "expense"
	CategoryTypeTransfer = "transfer"
)

// Default categories
var DefaultCategories = []Category{
	// Income categories
	{Name: "Income", Type: CategoryTypeIncome},
	{Name: "Salary", Type: CategoryTypeIncome, ParentID: strPtr("Income")},
	{Name: "Investment", Type: CategoryTypeIncome, ParentID: strPtr("Income")},
	
	// Housing categories
	{Name: "Housing", Type: CategoryTypeExpense},
	{Name: "Rent/Mortgage", Type: CategoryTypeExpense, ParentID: strPtr("Housing")},
	{Name: "Utilities", Type: CategoryTypeExpense, ParentID: strPtr("Housing")},
	
	// Transportation categories
	{Name: "Transportation", Type: CategoryTypeExpense},
	{Name: "Public Transit", Type: CategoryTypeExpense, ParentID: strPtr("Transportation")},
	{Name: "Fuel", Type: CategoryTypeExpense, ParentID: strPtr("Transportation")},
	
	// Food categories
	{Name: "Food", Type: CategoryTypeExpense},
	{Name: "Groceries", Type: CategoryTypeExpense, ParentID: strPtr("Food")},
	{Name: "Restaurants", Type: CategoryTypeExpense, ParentID: strPtr("Food")},
	
	// Shopping categories
	{Name: "Shopping", Type: CategoryTypeExpense},
	{Name: "Clothing", Type: CategoryTypeExpense, ParentID: strPtr("Shopping")},
	{Name: "Electronics", Type: CategoryTypeExpense, ParentID: strPtr("Shopping")},
	
	// Entertainment categories
	{Name: "Entertainment", Type: CategoryTypeExpense},
	{Name: "Movies", Type: CategoryTypeExpense, ParentID: strPtr("Entertainment")},
	{Name: "Games", Type: CategoryTypeExpense, ParentID: strPtr("Entertainment")},
	
	// Health categories
	{Name: "Health", Type: CategoryTypeExpense},
	{Name: "Medical", Type: CategoryTypeExpense, ParentID: strPtr("Health")},
	{Name: "Pharmacy", Type: CategoryTypeExpense, ParentID: strPtr("Health")},
	
	// Transfer categories
	{Name: "Transfer", Type: CategoryTypeTransfer},
	{Name: "Account Transfer", Type: CategoryTypeTransfer, ParentID: strPtr("Transfer")},
}

func strPtr(s string) *string {
	return &s
}
