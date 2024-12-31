package service

import (
	"context"
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/errors"
	"github.com/yeboahd24/personal-finance-manager/internal/metrics"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"github.com/yeboahd24/personal-finance-manager/internal/repository"
)

type RecurringTransactionService struct {
	repo           repository.Repository
	transactionSvc *TransactionService
	metrics        *metrics.RecurringTransactionMetrics
}

func NewRecurringTransactionService(repo repository.Repository, transactionSvc *TransactionService) *RecurringTransactionService {
	return &RecurringTransactionService{
		repo:           repo,
		transactionSvc: transactionSvc,
		metrics:        metrics.NewRecurringTransactionMetrics(),
	}
}

func (s *RecurringTransactionService) CreateRecurringTransaction(ctx context.Context, userID string, tx *model.RecurringTransaction) error {
	if tx.Amount == 0 {
		return errors.New("Amount is required", 400)
	}

	if tx.Description == "" {
		return errors.New("Description is required", 400)
	}

	if tx.AccountID == "" {
		return errors.New("Account ID is required", 400)
	}

	if tx.CategoryID == "" {
		return errors.New("Category ID is required", 400)
	}

	if tx.Interval == "" {
		return errors.New("Interval is required", 400)
	}

	// Validate interval-specific fields
	switch tx.Interval {
	case model.RecurrenceDaily:
		// No additional validation needed
	case model.RecurrenceWeekly:
		if tx.DayOfWeek != nil && (*tx.DayOfWeek < 0 || *tx.DayOfWeek > 6) {
			return errors.New("Day of week must be between 0 (Sunday) and 6 (Saturday)", 400)
		}
	case model.RecurrenceMonthly:
		if tx.DayOfMonth != nil && (*tx.DayOfMonth < 1 || *tx.DayOfMonth > 31) {
			return errors.New("Day of month must be between 1 and 31", 400)
		}
	case model.RecurrenceYearly:
		// No additional validation needed
	default:
		return errors.New("Invalid interval", 400)
	}

	if tx.StartDate.IsZero() {
		tx.StartDate = time.Now().UTC()
	}

	if tx.EndDate != nil && tx.EndDate.Before(tx.StartDate) {
		return errors.New("End date must be after start date", 400)
	}

	// Verify account exists and belongs to user
	account, err := s.repo.GetAccountByID(ctx, tx.AccountID)
	if err != nil {
		if err == errors.ErrNotFound {
			return errors.New("Account not found", 400)
		}
		return err
	}
	if account.UserID != userID {
		return errors.New("Account not found", 400)
	}

	// Verify category exists
	if _, err := s.repo.GetCategoryByID(ctx, tx.CategoryID); err != nil {
		if err == errors.ErrNotFound {
			return errors.New("Category not found", 400)
		}
		return err
	}

	tx.UserID = userID
	tx.Active = true
	tx.NextRun = tx.CalculateNextRun(time.Now().UTC())

	return s.repo.CreateRecurringTransaction(ctx, tx)
}

func (s *RecurringTransactionService) GetRecurringTransactionByID(ctx context.Context, userID, id string) (*model.RecurringTransaction, error) {
	tx, err := s.repo.GetRecurringTransactionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if tx.UserID != userID {
		return nil, errors.ErrNotFound
	}

	return tx, nil
}

func (s *RecurringTransactionService) GetRecurringTransactions(ctx context.Context, userID string, filter model.RecurringTransactionFilter) ([]*model.RecurringTransaction, error) {
	filter.UserID = userID
	return s.repo.GetRecurringTransactions(ctx, filter)
}

func (s *RecurringTransactionService) UpdateRecurringTransaction(ctx context.Context, userID string, tx *model.RecurringTransaction) error {
	existing, err := s.repo.GetRecurringTransactionByID(ctx, tx.ID)
	if err != nil {
		return err
	}

	if existing.UserID != userID {
		return errors.ErrNotFound
	}

	// Don't allow changing user_id
	tx.UserID = existing.UserID

	// Validate account if changed
	if tx.AccountID != existing.AccountID {
		account, err := s.repo.GetAccountByID(ctx, tx.AccountID)
		if err != nil {
			if err == errors.ErrNotFound {
				return errors.New("Account not found", 400)
			}
			return err
		}
		if account.UserID != userID {
			return errors.New("Account not found", 400)
		}
	}

	// Validate category if changed
	if tx.CategoryID != existing.CategoryID {
		if _, err := s.repo.GetCategoryByID(ctx, tx.CategoryID); err != nil {
			if err == errors.ErrNotFound {
				return errors.New("Category not found", 400)
			}
			return err
		}
	}

	// Recalculate next run if schedule changed
	if tx.Interval != existing.Interval ||
		tx.DayOfMonth != existing.DayOfMonth ||
		tx.DayOfWeek != existing.DayOfWeek ||
		!tx.StartDate.Equal(existing.StartDate) {
		tx.NextRun = tx.CalculateNextRun(time.Now().UTC())
	}

	return s.repo.UpdateRecurringTransaction(ctx, tx)
}

func (s *RecurringTransactionService) DeleteRecurringTransaction(ctx context.Context, userID, id string) error {
	existing, err := s.repo.GetRecurringTransactionByID(ctx, id)
	if err != nil {
		return err
	}

	if existing.UserID != userID {
		return errors.ErrNotFound
	}

	return s.repo.DeleteRecurringTransaction(ctx, id)
}

func (s *RecurringTransactionService) ProcessDueTransactions(ctx context.Context) error {
	start := time.Now()
	defer func() {
		s.metrics.SetProcessingTime(time.Since(start))
	}()

	now := time.Now().UTC()
	transactions, err := s.repo.GetDueRecurringTransactions(ctx, now)
	if err != nil {
		s.metrics.IncrementFailed()
		return err
	}

	for _, rt := range transactions {
		// Create the actual transaction
		tx := &model.Transaction{
			UserID:      rt.UserID,
			AccountID:   rt.AccountID,
			CategoryID:  &rt.CategoryID,
			Amount:      rt.Amount,
			Date:        now,
			Description: rt.Description,
		}

		if err := s.transactionSvc.CreateTransaction(ctx, rt.UserID, tx); err != nil {
			// Log error but continue processing other transactions
			s.metrics.IncrementFailed()
			continue
		}

		// Update the recurring transaction's last run and next run dates
		lastRun := now
		nextRun := rt.CalculateNextRun(now)

		if err := s.repo.UpdateLastRun(ctx, rt.ID, lastRun, nextRun); err != nil {
			// Log error but continue processing other transactions
			s.metrics.IncrementFailed()
			continue
		}

		s.metrics.IncrementProcessed()
	}

	return nil
}

// UpdateLastRun updates the last run and next run times for a recurring transaction
func (s *RecurringTransactionService) UpdateLastRun(ctx context.Context, id string, lastRun, nextRun time.Time) error {
	return s.repo.UpdateLastRun(ctx, id, lastRun, nextRun)
}

// CreateTransactionForRecurring creates a new transaction from a recurring transaction context
func (s *RecurringTransactionService) CreateTransactionForRecurring(ctx context.Context, userID string, tx *model.Transaction) error {
	return s.transactionSvc.CreateTransaction(ctx, userID, tx)
}

// GetMetrics returns the current metrics for recurring transaction processing
func (s *RecurringTransactionService) GetMetrics() *metrics.RecurringTransactionMetrics {
	return s.metrics
}
