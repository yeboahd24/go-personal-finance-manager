package queue

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"github.com/yeboahd24/personal-finance-manager/internal/service"
)

type RetryItem struct {
	Transaction     *model.RecurringTransaction
	RetryCount     int
	NextRetryTime  time.Time
	LastError      error
	LastRetryTime  time.Time
}

type RetryQueue struct {
	items            map[string]*RetryItem // key is transaction ID
	maxRetry         int
	mu               sync.RWMutex
	metrics          *RetryQueueMetrics
	notificationSvc  *service.NotificationService
}

type RetryQueueMetrics struct {
	CurrentSize    int
	TotalRetries   int64
	SuccessCount   int64
	PermanentFails int64
}

func NewRetryQueue(maxRetry int, notificationSvc *service.NotificationService) *RetryQueue {
	if maxRetry <= 0 {
		maxRetry = 3
	}
	return &RetryQueue{
		items:           make(map[string]*RetryItem),
		maxRetry:        maxRetry,
		metrics:         &RetryQueueMetrics{},
		notificationSvc: notificationSvc,
	}
}

func (q *RetryQueue) Add(tx *model.RecurringTransaction, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, exists := q.items[tx.ID]
	if !exists {
		item = &RetryItem{
			Transaction: tx,
			RetryCount: 0,
		}
		q.items[tx.ID] = item
		q.metrics.CurrentSize++
	}

	item.LastError = err
	item.LastRetryTime = time.Now()
	item.RetryCount++
	item.NextRetryTime = q.calculateNextRetryTime(item.RetryCount)
	q.metrics.TotalRetries++

	// Notify about the failure
	ctx := context.Background()
	if item.RetryCount > q.maxRetry {
		// Permanent failure notification
		if q.notificationSvc != nil {
			if err := q.notificationSvc.NotifyPermanentFailure(ctx, tx.UserID, tx, err); err != nil {
				// Log the error but don't fail the operation
				log.Printf("Failed to send permanent failure notification: %v", err)
			}
		}
	} else {
		// Regular failure notification
		if q.notificationSvc != nil {
			if err := q.notificationSvc.NotifyRecurringTransactionFailure(ctx, tx.UserID, tx, err, item.RetryCount); err != nil {
				log.Printf("Failed to send failure notification: %v", err)
			}
		}
	}
}

func (q *RetryQueue) Remove(txID string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.items[txID]; exists {
		delete(q.items, txID)
		q.metrics.CurrentSize--
		q.metrics.SuccessCount++
	}
}

func (q *RetryQueue) GetDueItems(now time.Time) []*RetryItem {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var dueItems []*RetryItem
	for _, item := range q.items {
		if item.RetryCount > q.maxRetry {
			// Mark as permanently failed
			q.metrics.PermanentFails++
			delete(q.items, item.Transaction.ID)
			q.metrics.CurrentSize--
			continue
		}

		if now.After(item.NextRetryTime) {
			dueItems = append(dueItems, item)
		}
	}
	return dueItems
}

func (q *RetryQueue) GetMetrics() RetryQueueMetrics {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return *q.metrics
}

func (q *RetryQueue) calculateNextRetryTime(retryCount int) time.Time {
	// Exponential backoff: 1min, 2min, 4min, 8min, etc.
	delay := time.Duration(1<<uint(retryCount-1)) * time.Minute
	if delay > 1*time.Hour {
		delay = 1 * time.Hour // Cap at 1 hour
	}
	return time.Now().Add(delay)
}

func (q *RetryQueue) GetItemStatus(txID string) (*RetryItem, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	item, exists := q.items[txID]
	return item, exists
}
