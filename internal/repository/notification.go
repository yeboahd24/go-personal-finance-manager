package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/yeboahd24/personal-finance-manager/internal/errors"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
)

type NotificationSQL struct {
	db *sql.DB
	tx *sql.Tx
}

func (r *NotificationSQL) query() QueryExecutor {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *NotificationSQL) CreateNotification(ctx context.Context, notification *model.Notification) error {
	query := `
		INSERT INTO notifications (
			user_id, type, priority, title, message, data, read, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		RETURNING id`

	data, err := json.Marshal(notification.Data)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal notification data", 500)
	}

	err = r.query().QueryRowContext(
		ctx,
		query,
		notification.UserID,
		notification.Type,
		notification.Priority,
		notification.Title,
		notification.Message,
		data,
		notification.Read,
		notification.CreatedAt,
		notification.UpdatedAt,
	).Scan(&notification.ID)

	if err != nil {
		return errors.Wrap(err, "Failed to create notification", 500)
	}
	return nil
}

func (r *NotificationSQL) GetNotificationByID(ctx context.Context, id string) (*model.Notification, error) {
	query := `
		SELECT 
			id, user_id, type, priority, title, message, data, read, created_at, updated_at
		FROM notifications
		WHERE id = $1`

	notification := &model.Notification{}
	var data []byte

	err := r.query().QueryRowContext(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Priority,
		&notification.Title,
		&notification.Message,
		&data,
		&notification.Read,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get notification", 500)
	}

	if err := json.Unmarshal(data, &notification.Data); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal notification data", 500)
	}

	return notification, nil
}

func (r *NotificationSQL) GetUserNotifications(ctx context.Context, userID string, unreadOnly bool) ([]*model.Notification, error) {
	query := `
		SELECT 
			id, user_id, type, priority, title, message, data, read, created_at, updated_at
		FROM notifications
		WHERE user_id = $1
		AND ($2 = false OR read = false)
		ORDER BY created_at DESC`

	rows, err := r.query().QueryContext(ctx, query, userID, unreadOnly)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get user notifications", 500)
	}
	defer rows.Close()

	var notifications []*model.Notification
	for rows.Next() {
		notification := &model.Notification{}
		var data []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Priority,
			&notification.Title,
			&notification.Message,
			&data,
			&notification.Read,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to scan notification", 500)
		}

		if err := json.Unmarshal(data, &notification.Data); err != nil {
			return nil, errors.Wrap(err, "Failed to unmarshal notification data", 500)
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func (r *NotificationSQL) MarkNotificationAsRead(ctx context.Context, userID string, notificationID string) error {
	query := `
		UPDATE notifications
		SET read = true,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2`

	result, err := r.query().ExecContext(ctx, query, notificationID, userID)
	if err != nil {
		return errors.Wrap(err, "Failed to mark notification as read", 500)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "Failed to get rows affected", 500)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *NotificationSQL) DeleteNotification(ctx context.Context, id string) error {
	query := "DELETE FROM notifications WHERE id = $1"

	result, err := r.query().ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "Failed to delete notification", 500)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "Failed to get rows affected", 500)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *NotificationSQL) GetNotificationPreferences(ctx context.Context, userID string) (*model.NotificationPreferences, error) {
	query := `
		SELECT 
			id, user_id, email_enabled, push_enabled, in_app_enabled,
			min_priority, recurring_failures, upcoming_recurring,
			created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1`

	prefs := &model.NotificationPreferences{}
	err := r.query().QueryRowContext(ctx, query, userID).Scan(
		&prefs.ID,
		&prefs.UserID,
		&prefs.EmailEnabled,
		&prefs.PushEnabled,
		&prefs.InAppEnabled,
		&prefs.MinPriority,
		&prefs.RecurringFailures,
		&prefs.UpcomingRecurring,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return default preferences
		return &model.NotificationPreferences{
			UserID:             userID,
			EmailEnabled:       true,
			InAppEnabled:       true,
			MinPriority:       model.NotificationPriorityLow,
			RecurringFailures: true,
			UpcomingRecurring: true,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get notification preferences", 500)
	}

	return prefs, nil
}

func (r *NotificationSQL) UpdateNotificationPreferences(ctx context.Context, prefs *model.NotificationPreferences) error {
	query := `
		INSERT INTO notification_preferences (
			user_id, email_enabled, push_enabled, in_app_enabled,
			min_priority, recurring_failures, upcoming_recurring
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		ON CONFLICT (user_id) DO UPDATE
		SET
			email_enabled = $2,
			push_enabled = $3,
			in_app_enabled = $4,
			min_priority = $5,
			recurring_failures = $6,
			upcoming_recurring = $7,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at`

	err := r.query().QueryRowContext(
		ctx,
		query,
		prefs.UserID,
		prefs.EmailEnabled,
		prefs.PushEnabled,
		prefs.InAppEnabled,
		prefs.MinPriority,
		prefs.RecurringFailures,
		prefs.UpcomingRecurring,
	).Scan(&prefs.ID, &prefs.CreatedAt, &prefs.UpdatedAt)

	if err != nil {
		return errors.Wrap(err, "Failed to update notification preferences", 500)
	}
	return nil
}
