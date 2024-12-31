package model

import (
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/errors"
)

type Budget struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	CategoryID  string    `json:"category_id"`
	Amount      float64   `json:"amount"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Populated fields
	Category     *Category `json:"category,omitempty"`
	SpentAmount  *float64  `json:"spent_amount,omitempty"`
	SpentPercent *float64  `json:"spent_percent,omitempty"`
}

type BudgetSummary struct {
	TotalBudget     float64 `json:"total_budget"`
	TotalSpent      float64 `json:"total_spent"`
	RemainingBudget float64 `json:"remaining_budget"`
	SpentPercent    float64 `json:"spent_percent"`
}

type BudgetFilter struct {
	UserID      string
	CategoryID  string
	PeriodStart time.Time
	PeriodEnd   time.Time
}

// GetPeriodDates converts a period string into start and end times
// Valid periods are: "daily", "weekly", "monthly", "yearly"
func GetPeriodDates(period string) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	switch period {
	case "daily":
		end := start.Add(24 * time.Hour)
		return start, end, nil
	case "weekly":
		end := start.Add(7 * 24 * time.Hour)
		return start, end, nil
	case "monthly":
		end := start.AddDate(0, 1, 0)
		return start, end, nil
	case "yearly":
		end := start.AddDate(1, 0, 0)
		return start, end, nil
	default:
		return time.Time{}, time.Time{}, errors.New("invalid period: must be daily, weekly, monthly, or yearly", 400)
	}
}
