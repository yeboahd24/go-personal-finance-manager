package model

import (
	"time"
)

// Goal represents a financial goal
type Goal struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	TargetAmount float64   `json:"target_amount"`
	CurrentAmount float64  `json:"current_amount"`
	Deadline    time.Time `json:"deadline"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
