package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"github.com/yeboahd24/personal-finance-manager/internal/queue"
	"github.com/yeboahd24/personal-finance-manager/internal/service"
)

type RecurringTransactionWorker struct {
	recurringService *service.RecurringTransactionService
	interval         time.Duration
	retryQueue       *queue.RetryQueue
	stopChan         chan struct{}
	wg               sync.WaitGroup
}

// NewRecurringTransactionWorker creates a new worker for processing recurring transactions
func NewRecurringTransactionWorker(recurringService *service.RecurringTransactionService, notificationService *service.NotificationService, interval time.Duration) *RecurringTransactionWorker {
	if interval < time.Minute {
		interval = time.Minute
	}
	return &RecurringTransactionWorker{
		recurringService: recurringService,
		interval:        interval,
		retryQueue:      queue.NewRetryQueue(3, notificationService), // Max 3 retries
		stopChan:        make(chan struct{}),
	}
}

func (w *RecurringTransactionWorker) Start(ctx context.Context) {
	// Start main processing worker
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		// Run once immediately on startup
		w.processRecurringTransactions(ctx)

		for {
			select {
			case <-ctx.Done():
				log.Println("Context cancelled, stopping recurring transaction worker")
				return
			case <-w.stopChan:
				log.Println("Stop signal received, stopping recurring transaction worker")
				return
			case <-ticker.C:
				w.processRecurringTransactions(ctx)
			}
		}
	}()

	// Start retry worker
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		retryTicker := time.NewTicker(time.Minute) // Check retries every minute
		defer retryTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-w.stopChan:
				return
			case <-retryTicker.C:
				w.processRetries(ctx)
			}
		}
	}()
}

func (w *RecurringTransactionWorker) Stop() {
	close(w.stopChan)
	w.wg.Wait()
}

func (w *RecurringTransactionWorker) processRecurringTransactions(ctx context.Context) {
	// Create a new context with timeout for this processing cycle
	ctx, cancel := context.WithTimeout(ctx, w.interval/2)
	defer cancel()

	if err := w.recurringService.ProcessDueTransactions(ctx); err != nil {
		log.Printf("Error processing recurring transactions: %v", err)
		w.retryQueue.Add(&model.RecurringTransaction{}, err)
	}
}

func (w *RecurringTransactionWorker) processRetries(ctx context.Context) {
	dueItems := w.retryQueue.GetDueItems(time.Now())
	for _, item := range dueItems {
		// Create a new context with timeout for each retry
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)

		err := w.processTransaction(ctx, item.Transaction)
		if err != nil {
			// If still failing, add back to retry queue
			w.retryQueue.Add(item.Transaction, err)
		} else {
			// If successful, remove from retry queue
			w.retryQueue.Remove(item.Transaction.ID)
		}

		cancel()
	}
}

func (w *RecurringTransactionWorker) processTransaction(ctx context.Context, rt *model.RecurringTransaction) error {
	now := time.Now().UTC()

	// Create the actual transaction
	tx := &model.Transaction{
		UserID:      rt.UserID,
		AccountID:   rt.AccountID,
		CategoryID:  &rt.CategoryID,
		Amount:      rt.Amount,
		Date:        now,
		Description: rt.Description,
	}

	if err := w.recurringService.CreateTransactionForRecurring(ctx, rt.UserID, tx); err != nil {
		return err
	}

	// Update the recurring transaction's last run and next run dates
	lastRun := now
	nextRun := rt.CalculateNextRun(now)

	return w.recurringService.UpdateLastRun(ctx, rt.ID, lastRun, nextRun)
}

// GetRetryQueueMetrics returns metrics about the retry queue
func (w *RecurringTransactionWorker) GetRetryQueueMetrics() queue.RetryQueueMetrics {
	return w.retryQueue.GetMetrics()
}
