package model

import (
	"time"
)

type NotificationType string

const (
	NotificationTypeRecurringFailed    NotificationType = "recurring_failed"
	NotificationTypeRecurringRetry     NotificationType = "recurring_retry"
	NotificationTypePermanentFail      NotificationType = "permanent_fail"
	NotificationTypeRecurringUpcoming  NotificationType = "recurring_upcoming"
)

type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityMedium NotificationPriority = "medium"
	NotificationPriorityHigh   NotificationPriority = "high"
)

type Notification struct {
	ID          string              `json:"id"`
	UserID      string              `json:"user_id"`
	Type        NotificationType    `json:"type"`
	Priority    NotificationPriority `json:"priority"`
	Title       string              `json:"title"`
	Message     string              `json:"message"`
	Data        map[string]interface{} `json:"data"`
	Read        bool                `json:"read"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type NotificationPreferences struct {
	ID                   string `json:"id"`
	UserID              string `json:"user_id"`
	EmailEnabled        bool   `json:"email_enabled"`
	PushEnabled         bool   `json:"push_enabled"`
	InAppEnabled        bool   `json:"in_app_enabled"`
	MinPriority        NotificationPriority `json:"min_priority"`
	RecurringFailures  bool   `json:"recurring_failures"`
	UpcomingRecurring  bool   `json:"upcoming_recurring"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
