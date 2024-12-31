package metrics

import (
	"sync/atomic"
	"time"
)

// RecurringTransactionMetrics tracks metrics for recurring transactions processing
type RecurringTransactionMetrics struct {
	ProcessedCount     uint64
	FailedCount       uint64
	LastProcessedTime atomic.Value // stores time.Time
	LastErrorTime     atomic.Value // stores time.Time
	ProcessingTime    atomic.Value // stores time.Duration
}

func NewRecurringTransactionMetrics() *RecurringTransactionMetrics {
	m := &RecurringTransactionMetrics{}
	m.LastProcessedTime.Store(time.Time{})
	m.LastErrorTime.Store(time.Time{})
	m.ProcessingTime.Store(time.Duration(0))
	return m
}

func (m *RecurringTransactionMetrics) IncrementProcessed() {
	atomic.AddUint64(&m.ProcessedCount, 1)
	m.LastProcessedTime.Store(time.Now())
}

func (m *RecurringTransactionMetrics) IncrementFailed() {
	atomic.AddUint64(&m.FailedCount, 1)
	m.LastErrorTime.Store(time.Now())
}

func (m *RecurringTransactionMetrics) SetProcessingTime(d time.Duration) {
	m.ProcessingTime.Store(d)
}

func (m *RecurringTransactionMetrics) GetProcessedCount() uint64 {
	return atomic.LoadUint64(&m.ProcessedCount)
}

func (m *RecurringTransactionMetrics) GetFailedCount() uint64 {
	return atomic.LoadUint64(&m.FailedCount)
}

func (m *RecurringTransactionMetrics) GetLastProcessedTime() time.Time {
	return m.LastProcessedTime.Load().(time.Time)
}

func (m *RecurringTransactionMetrics) GetLastErrorTime() time.Time {
	return m.LastErrorTime.Load().(time.Time)
}

func (m *RecurringTransactionMetrics) GetProcessingTime() time.Duration {
	return m.ProcessingTime.Load().(time.Duration)
}
