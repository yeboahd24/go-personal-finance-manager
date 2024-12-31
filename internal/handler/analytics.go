package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/yeboahd24/personal-finance-manager/internal/middleware"
	"github.com/yeboahd24/personal-finance-manager/internal/service"
)

type AnalyticsHandler struct {
	analyticsService *service.AnalyticsService
}

func NewAnalyticsHandler(analyticsService *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

func (h *AnalyticsHandler) GetSpendingByCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Failed to get user ID: "+err.Error(), http.StatusUnauthorized)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month"
	}

	spending, err := h.analyticsService.GetSpendingByCategory(r.Context(), userID, period)
	if err != nil {
		http.Error(w, "Failed to get spending by category: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(spending); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *AnalyticsHandler) GetMonthlySpending(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Failed to get user ID: "+err.Error(), http.StatusUnauthorized)
		return
	}

	months := 12
	if m := r.URL.Query().Get("months"); m != "" {
		if val, err := strconv.Atoi(m); err == nil && val > 0 {
			months = val
		}
	}

	spending, err := h.analyticsService.GetMonthlySpending(r.Context(), userID, months)
	if err != nil {
		http.Error(w, "Failed to get monthly spending: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(spending); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *AnalyticsHandler) GetCashFlow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Failed to get user ID: "+err.Error(), http.StatusUnauthorized)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month"
	}

	cashFlow, err := h.analyticsService.GetCashFlow(r.Context(), userID, period)
	if err != nil {
		http.Error(w, "Failed to get cash flow: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cashFlow); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *AnalyticsHandler) GetTopMerchants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Failed to get user ID: "+err.Error(), http.StatusUnauthorized)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month"
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	merchants, err := h.analyticsService.GetTopMerchants(r.Context(), userID, period, limit)
	if err != nil {
		http.Error(w, "Failed to get top merchants: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(merchants); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *AnalyticsHandler) GetFinancialReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Failed to get user ID: "+err.Error(), http.StatusUnauthorized)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month"
	}

	report, err := h.analyticsService.GetFinancialReport(r.Context(), userID, period)
	if err != nil {
		http.Error(w, "Failed to get financial report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *AnalyticsHandler) GetIncomeVsExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		http.Error(w, "Failed to get user ID: "+err.Error(), http.StatusUnauthorized)
		return
	}
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month" // default to current month
	}

	data, err := h.analyticsService.GetIncomeVsExpenses(ctx, userID, period)
	if err != nil {
		http.Error(w, "Failed to get income vs expenses data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if we need to return specific data based on the request path
	path := strings.TrimPrefix(r.URL.Path, "/api/analytics/")
	path = strings.TrimSuffix(path, "/")

	w.Header().Set("Content-Type", "application/json")
	switch path {
	case "income":
		if err := json.NewEncoder(w).Encode(data.TotalIncome); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	case "expenses":
		if err := json.NewEncoder(w).Encode(data.TotalExpenses); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	case "savings":
		savings := data.TotalIncome - data.TotalExpenses
		if err := json.NewEncoder(w).Encode(savings); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *AnalyticsHandler) GetBudgetPerformance(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month" // default to current month
	}

	ctx := r.Context()
	performance, err := h.analyticsService.GetBudgetPerformance(ctx, userID, period)
	if err != nil {
		http.Error(w, "Failed to get budget performance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(performance); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ServeHTTP implements the http.Handler interface
func (h *AnalyticsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/analytics/")
	path = strings.TrimSuffix(path, "/")

	switch path {
	case "spending-by-category":
		h.GetSpendingByCategory(w, r)
	case "monthly-spending":
		h.GetMonthlySpending(w, r)
	case "cash-flow":
		h.GetCashFlow(w, r)
	case "top-merchants":
		h.GetTopMerchants(w, r)
	case "financial-report":
		h.GetFinancialReport(w, r)
	case "income":
		h.GetIncomeVsExpenses(w, r)
	case "expenses":
		h.GetIncomeVsExpenses(w, r)
	case "savings":
		h.GetIncomeVsExpenses(w, r)
	default:
		http.Error(w, "Invalid analytics type", http.StatusBadRequest)
	}
}
