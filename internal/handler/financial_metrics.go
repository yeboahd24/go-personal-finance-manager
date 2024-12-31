package handler

import (
	"encoding/json"
	"net/http"

	"github.com/yeboahd24/personal-finance-manager/internal/middleware"
	"github.com/yeboahd24/personal-finance-manager/internal/service"
)

// FinancialMetricsHandler handles financial metric-related HTTP requests
type FinancialMetricsHandler struct {
	metricsService *service.MetricsService
}

// NewFinancialMetricsHandler creates a new financial metrics handler
func NewFinancialMetricsHandler(metricsService *service.MetricsService) *FinancialMetricsHandler {
	return &FinancialMetricsHandler{
		metricsService: metricsService,
	}
}

// GetNetWorth handles net worth calculation requests
func (h *FinancialMetricsHandler) GetNetWorth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	netWorth, err := h.metricsService.GetNetWorth(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{
		"net_worth": netWorth,
	})
}

// GetSavingsRate handles savings rate calculation requests
func (h *FinancialMetricsHandler) GetSavingsRate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	savingsRate, err := h.metricsService.GetSavingsRate(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{
		"savings_rate": savingsRate,
	})
}

// GetDebtToIncomeRatio handles debt-to-income ratio calculation requests
func (h *FinancialMetricsHandler) GetDebtToIncomeRatio(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ratio, err := h.metricsService.GetDebtToIncomeRatio(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{
		"debt_to_income_ratio": ratio,
	})
}

// GetEmergencyFundCoverage handles emergency fund coverage calculation requests
func (h *FinancialMetricsHandler) GetEmergencyFundCoverage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	coverage, err := h.metricsService.GetEmergencyFundCoverage(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{
		"emergency_fund_coverage": coverage,
	})
}

// ServeHTTP implements the http.Handler interface
func (h *FinancialMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		switch r.URL.Query().Get("type") {
		case "net_worth":
			h.GetNetWorth(w, r)
		case "savings_rate":
			h.GetSavingsRate(w, r)
		case "debt_to_income":
			h.GetDebtToIncomeRatio(w, r)
		case "emergency_fund":
			h.GetEmergencyFundCoverage(w, r)
		default:
			http.Error(w, "Invalid metrics type", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
