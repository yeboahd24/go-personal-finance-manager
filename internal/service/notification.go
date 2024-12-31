package service

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"github.com/yeboahd24/personal-finance-manager/internal/repository"
)

type NotificationService struct {
	repo         repository.Repository
	emailService *EmailService
}

type NotificationRequest struct {
	UserID   string                 `json:"user_id"`
	Title    string                 `json:"title"`
	Message  string                 `json:"message"`
	Type     string                 `json:"type"`
	Priority string                 `json:"priority"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func NewNotificationService(repo repository.Repository, emailService *EmailService) *NotificationService {
	return &NotificationService{
		repo:         repo,
		emailService: emailService,
	}
}

func (s *NotificationService) CreateNotification(ctx context.Context, notification *model.Notification) error {
	if err := s.repo.CreateNotification(ctx, notification); err != nil {
		return err
	}

	// Get user's notification preferences
	prefs, err := s.repo.GetNotificationPreferences(ctx, notification.UserID)
	if err != nil {
		return err
	}

	// Check if notification meets priority threshold
	if !s.shouldSendNotification(notification, prefs) {
		return nil
	}

	// Send notifications based on user preferences
	if prefs.EmailEnabled {
		if err := s.sendEmailNotification(ctx, notification); err != nil {
			// Log error but don't fail the notification creation
			fmt.Printf("Failed to send email notification: %v\n", err)
		}
	}

	if prefs.PushEnabled {
		if err := s.sendPushNotification(ctx, notification); err != nil {
			fmt.Printf("Failed to send push notification: %v\n", err)
		}
	}

	return nil
}

func (s *NotificationService) NotifyRecurringTransactionFailure(ctx context.Context, userID string, tx *model.RecurringTransaction, err error, retryCount int) error {
	notification := &model.Notification{
		UserID:   userID,
		Type:     model.NotificationTypeRecurringFailed,
		Priority: model.NotificationPriorityHigh,
		Title:    "Recurring Transaction Failed",
		Message:  fmt.Sprintf("Failed to process recurring transaction: %s", tx.Description),
		Data: map[string]interface{}{
			"transaction_id": tx.ID,
			"error":          err.Error(),
			"retry_count":    retryCount,
			"amount":         tx.Amount,
		},
		Read:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.CreateNotification(ctx, notification)
}

func (s *NotificationService) NotifyPermanentFailure(ctx context.Context, userID string, tx *model.RecurringTransaction, err error) error {
	notification := &model.Notification{
		UserID:   userID,
		Type:     model.NotificationTypePermanentFail,
		Priority: model.NotificationPriorityHigh,
		Title:    "Recurring Transaction Permanently Failed",
		Message:  fmt.Sprintf("Recurring transaction has permanently failed after multiple retries: %s", tx.Description),
		Data: map[string]interface{}{
			"transaction_id": tx.ID,
			"error":          err.Error(),
			"amount":         tx.Amount,
		},
		Read:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.CreateNotification(ctx, notification)
}

func (s *NotificationService) NotifyUpcomingRecurring(ctx context.Context, userID string, tx *model.RecurringTransaction) error {
	notification := &model.Notification{
		UserID:   userID,
		Type:     model.NotificationTypeRecurringUpcoming,
		Priority: model.NotificationPriorityMedium,
		Title:    "Upcoming Recurring Transaction",
		Message:  fmt.Sprintf("Upcoming recurring transaction: %s for %.2f", tx.Description, tx.Amount),
		Data: map[string]interface{}{
			"transaction_id": tx.ID,
			"amount":         tx.Amount,
			"due_date":       tx.NextRun,
		},
		Read:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.CreateNotification(ctx, notification)
}

func (s *NotificationService) shouldSendNotification(notification *model.Notification, prefs *model.NotificationPreferences) bool {
	// Check priority threshold
	switch prefs.MinPriority {
	case model.NotificationPriorityHigh:
		if notification.Priority != model.NotificationPriorityHigh {
			return false
		}
	case model.NotificationPriorityMedium:
		if notification.Priority == model.NotificationPriorityLow {
			return false
		}
	}

	// Check notification type preferences
	switch notification.Type {
	case model.NotificationTypeRecurringFailed, model.NotificationTypePermanentFail:
		return prefs.RecurringFailures
	case model.NotificationTypeRecurringUpcoming:
		return prefs.UpcomingRecurring
	default:
		return true
	}
}

func (s *NotificationService) sendEmailNotification(ctx context.Context, notification *model.Notification) error {
	// Get user's email
	user, err := s.repo.GetUserByID(ctx, notification.UserID)
	if err != nil {
		return err
	}

	// Create email template
	tmpl := template.Must(template.New("notification").Parse(`
		<html>
			<body>
				<h2>{{.Title}}</h2>
				<p>{{.Message}}</p>
				{{if .Data.amount}}
				<p>Amount: ${{.Data.amount}}</p>
				{{end}}
				{{if .Data.due_date}}
				<p>Due Date: {{.Data.due_date}}</p>
				{{end}}
				<p>Please check your account for more details.</p>
			</body>
		</html>
	`))

	var body bytes.Buffer
	if err := tmpl.Execute(&body, notification); err != nil {
		return err
	}

	return s.emailService.SendEmail(ctx, user.Email, notification.Title, body.String())
}

func (s *NotificationService) sendPushNotification(ctx context.Context, notification *model.Notification) error {
	// Implement push notification logic here
	// This could integrate with Firebase Cloud Messaging or another push service
	return nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, userID string, notificationID string) error {
	return s.repo.MarkNotificationAsRead(ctx, userID, notificationID)
}

func (s *NotificationService) GetUserNotifications(ctx context.Context, userID string, unreadOnly bool) ([]*model.Notification, error) {
	return s.repo.GetUserNotifications(ctx, userID, unreadOnly)
}

func (s *NotificationService) UpdateNotificationPreferences(ctx context.Context, userID string, prefs *model.NotificationPreferences) error {
	prefs.UserID = userID
	prefs.UpdatedAt = time.Now()
	return s.repo.UpdateNotificationPreferences(ctx, prefs)
}

func (s *NotificationService) GetNotificationPreferences(ctx context.Context, userID string) (*model.NotificationPreferences, error) {
	return s.repo.GetNotificationPreferences(ctx, userID)
}
