package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/service"
	"github.com/yeboahd24/personal-finance-manager/internal/worker"
)

type RecurringTransactionMetricsHandler struct {
	recurringService *service.RecurringTransactionService
	recurringWorker  *worker.RecurringTransactionWorker
}

func NewRecurringTransactionMetricsHandler(recurringService *service.RecurringTransactionService, recurringWorker *worker.RecurringTransactionWorker) *RecurringTransactionMetricsHandler {
	return &RecurringTransactionMetricsHandler{
		recurringService: recurringService,
		recurringWorker:  recurringWorker,
	}
}

type RecurringTransactionMetricsResponse struct {
	ProcessedCount    uint64  `json:"processed_count"`
	FailedCount       uint64  `json:"failed_count"`
	LastProcessedTime string  `json:"last_processed_time"`
	LastErrorTime     string  `json:"last_error_time"`
	ProcessingTimeMs  float64 `json:"processing_time_ms"`
	RetryQueue        struct {
		CurrentSize     int   `json:"current_size"`
		TotalRetries    int64 `json:"total_retries"`
		SuccessfulRetry int64 `json:"successful_retry"`
		PermanentFails  int64 `json:"permanent_fails"`
	} `json:"retry_queue"`
}

func (h *RecurringTransactionMetricsHandler) GetRecurringTransactionMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := h.recurringService.GetMetrics()
	retryMetrics := h.recurringWorker.GetRetryQueueMetrics()

	response := RecurringTransactionMetricsResponse{
		ProcessedCount:   metrics.GetProcessedCount(),
		FailedCount:      metrics.GetFailedCount(),
		ProcessingTimeMs: float64(metrics.GetProcessingTime().Milliseconds()),
	}

	lastProcessed := metrics.GetLastProcessedTime()
	if !lastProcessed.IsZero() {
		response.LastProcessedTime = lastProcessed.Format(time.RFC3339)
	}

	lastError := metrics.GetLastErrorTime()
	if !lastError.IsZero() {
		response.LastErrorTime = lastError.Format(time.RFC3339)
	}

	response.RetryQueue.CurrentSize = retryMetrics.CurrentSize
	response.RetryQueue.TotalRetries = retryMetrics.TotalRetries
	response.RetryQueue.SuccessfulRetry = retryMetrics.SuccessCount
	response.RetryQueue.PermanentFails = retryMetrics.PermanentFails

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ServeHTTP implements the http.Handler interface
func (h *RecurringTransactionMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetRecurringTransactionMetrics(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
