package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/middleware"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"github.com/yeboahd24/personal-finance-manager/internal/service"
)

type BudgetHandler struct {
	budgetService *service.BudgetService
}

func NewBudgetHandler(budgetService *service.BudgetService) *BudgetHandler {
	return &BudgetHandler{
		budgetService: budgetService,
	}
}

func (h *BudgetHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var budget model.Budget
	if err := json.NewDecoder(r.Body).Decode(&budget); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.budgetService.CreateBudget(r.Context(), userID, &budget); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budget)
}

func (h *BudgetHandler) GetBudgets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	filter := model.BudgetFilter{}

	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		filter.CategoryID = categoryID
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		t, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			http.Error(w, "Invalid start_date format", http.StatusBadRequest)
			return
		}
		filter.PeriodStart = t
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		t, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			http.Error(w, "Invalid end_date format", http.StatusBadRequest)
			return
		}
		filter.PeriodEnd = t
	}

	budgets, err := h.budgetService.GetBudgets(r.Context(), userID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budgets)
}

func (h *BudgetHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/budgets/")
	if id == "" {
		http.Error(w, "Budget ID required", http.StatusBadRequest)
		return
	}

	budget, err := h.budgetService.GetBudgetByID(r.Context(), userID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budget)
}

func (h *BudgetHandler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/budgets/")
	if id == "" {
		http.Error(w, "Budget ID required", http.StatusBadRequest)
		return
	}

	var budget model.Budget
	if err := json.NewDecoder(r.Body).Decode(&budget); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	budget.ID = id
	if err := h.budgetService.UpdateBudget(r.Context(), userID, &budget); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budget)
}

func (h *BudgetHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/budgets/")
	if id == "" {
		http.Error(w, "Budget ID required", http.StatusBadRequest)
		return
	}

	if err := h.budgetService.DeleteBudget(r.Context(), userID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BudgetHandler) GetBudgetSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month" // default to current month
	}

	summary, err := h.budgetService.GetBudgetSummary(r.Context(), userID, period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// ServeHTTP implements the http.Handler interface
func (h *BudgetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetBudgets(w, r)
	case http.MethodPost:
		h.CreateBudget(w, r)
	case http.MethodPut:
		h.UpdateBudget(w, r)
	case http.MethodDelete:
		h.DeleteBudget(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
