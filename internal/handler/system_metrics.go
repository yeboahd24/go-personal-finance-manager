package handler

import (
	"encoding/json"
	"net/http"

	"github.com/yeboahd24/personal-finance-manager/internal/middleware"
	"github.com/yeboahd24/personal-finance-manager/internal/service"
)

type SystemMetricsHandler struct {
	metricsService *service.MetricsService
}

func NewSystemMetricsHandler(metricsService *service.MetricsService) *SystemMetricsHandler {
	return &SystemMetricsHandler{
		metricsService: metricsService,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *SystemMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get metrics type from query parameter
	metricsType := r.URL.Query().Get("type")
	if metricsType == "" {
		http.Error(w, "Metrics type is required", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var result interface{}
	var metricErr error

	switch metricsType {
	case "net_worth":
		result, metricErr = h.metricsService.GetNetWorth(r.Context(), userID)
	case "savings_rate":
		result, metricErr = h.metricsService.GetSavingsRate(r.Context(), userID)
	case "debt_to_income":
		result, metricErr = h.metricsService.GetDebtToIncomeRatio(r.Context(), userID)
	case "emergency_fund":
		result, metricErr = h.metricsService.GetEmergencyFundCoverage(r.Context(), userID)
	default:
		http.Error(w, "Invalid metrics type", http.StatusBadRequest)
		return
	}

	if metricErr != nil {
		http.Error(w, metricErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
