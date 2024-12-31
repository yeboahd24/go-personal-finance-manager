package model

import (
	"time"
)

type RecurrenceInterval string

const (
	RecurrenceDaily   RecurrenceInterval = "daily"
	RecurrenceWeekly  RecurrenceInterval = "weekly"
	RecurrenceMonthly RecurrenceInterval = "monthly"
	RecurrenceYearly  RecurrenceInterval = "yearly"
)

type RecurringTransaction struct {
	ID          string             `json:"id"`
	UserID      string             `json:"user_id"`
	AccountID   string             `json:"account_id"`
	CategoryID  string             `json:"category_id"`
	Amount      float64            `json:"amount"`
	Description string             `json:"description"`
	Interval    RecurrenceInterval `json:"interval"`
	DayOfMonth  *int              `json:"day_of_month,omitempty"` // For monthly recurrence
	DayOfWeek   *int              `json:"day_of_week,omitempty"`  // For weekly recurrence (0 = Sunday)
	StartDate   time.Time         `json:"start_date"`
	EndDate     *time.Time        `json:"end_date,omitempty"`
	LastRun     *time.Time        `json:"last_run,omitempty"`
	NextRun     time.Time         `json:"next_run"`
	Active      bool              `json:"active"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`

	// Populated fields
	Account  *Account  `json:"account,omitempty"`
	Category *Category `json:"category,omitempty"`
}

type RecurringTransactionFilter struct {
	UserID     string
	AccountID  string
	CategoryID string
	Active     *bool
}

// CalculateNextRun calculates the next run date based on the interval and last run
func (r *RecurringTransaction) CalculateNextRun(from time.Time) time.Time {
	if r.LastRun == nil {
		return r.StartDate
	}

	var next time.Time
	switch r.Interval {
	case RecurrenceDaily:
		next = r.LastRun.AddDate(0, 0, 1)

	case RecurrenceWeekly:
		next = r.LastRun.AddDate(0, 0, 7)
		if r.DayOfWeek != nil {
			// Adjust to the specified day of week
			for next.Weekday() != time.Weekday(*r.DayOfWeek) {
				next = next.AddDate(0, 0, 1)
			}
		}

	case RecurrenceMonthly:
		next = r.LastRun.AddDate(0, 1, 0)
		if r.DayOfMonth != nil {
			// Adjust to the specified day of month
			day := *r.DayOfMonth
			if day > 28 {
				// Handle months with different number of days
				lastDay := time.Date(next.Year(), next.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
				if day > lastDay {
					day = lastDay
				}
			}
			next = time.Date(next.Year(), next.Month(), day, next.Hour(), next.Minute(), next.Second(), next.Nanosecond(), next.Location())
		}

	case RecurrenceYearly:
		next = r.LastRun.AddDate(1, 0, 0)
	}

	// If next run is before the from date (which could happen if we missed some runs),
	// keep advancing until we get a future date
	for next.Before(from) {
		next = r.CalculateNextRun(next)
	}

	// If we have an end date and the next run would be after it, return zero time
	if r.EndDate != nil && next.After(*r.EndDate) {
		return time.Time{}
	}

	return next
}

// IsDue checks if the recurring transaction is due to run
func (r *RecurringTransaction) IsDue(now time.Time) bool {
	if !r.Active {
		return false
	}

	if r.EndDate != nil && now.After(*r.EndDate) {
		return false
	}

	if r.LastRun == nil {
		return !now.Before(r.StartDate)
	}

	return !now.Before(r.NextRun)
}
