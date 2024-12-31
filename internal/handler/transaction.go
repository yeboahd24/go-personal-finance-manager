package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/errors"
	"github.com/yeboahd24/personal-finance-manager/internal/middleware"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"github.com/yeboahd24/personal-finance-manager/internal/service"
)

type TransactionHandler struct {
	transactionService *service.TransactionService
}

func NewTransactionHandler(transactionService *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, errors.Message(err), errors.StatusCode(err))
		return
	}

	filter := model.TransactionFilter{
		UserID:    userID,
		AccountID: r.URL.Query().Get("account_id"),
	}

	// Parse date filters
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		t, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			http.Error(w, "Invalid start_date format", http.StatusBadRequest)
			return
		}
		filter.StartDate = t
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		t, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			http.Error(w, "Invalid end_date format", http.StatusBadRequest)
			return
		}
		filter.EndDate = t
	}

	// Parse amount filters
	if minAmount := r.URL.Query().Get("min_amount"); minAmount != "" {
		amount, err := strconv.ParseFloat(minAmount, 64)
		if err != nil {
			http.Error(w, "Invalid min_amount format", http.StatusBadRequest)
			return
		}
		filter.MinAmount = &amount
	}

	if maxAmount := r.URL.Query().Get("max_amount"); maxAmount != "" {
		amount, err := strconv.ParseFloat(maxAmount, 64)
		if err != nil {
			http.Error(w, "Invalid max_amount format", http.StatusBadRequest)
			return
		}
		filter.MaxAmount = &amount
	}

	// Parse other filters
	filter.Type = r.URL.Query().Get("type")
	filter.Search = r.URL.Query().Get("search")

	if limit := r.URL.Query().Get("limit"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			http.Error(w, "Invalid limit format", http.StatusBadRequest)
			return
		}
		filter.Limit = l
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			http.Error(w, "Invalid offset format", http.StatusBadRequest)
			return
		}
		filter.Offset = o
	}

	transactions, err := h.transactionService.GetTransactions(r.Context(), userID, filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get transactions: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": transactions,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

type updateTransactionRequest struct {
	CategoryID  *string `json:"category_id"`
	Description string  `json:"description"`
}

func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, errors.Message(err), errors.StatusCode(err))
		return
	}

	transactionID := r.URL.Path[len("/api/transactions/"):]
	if transactionID == "" {
		http.Error(w, "Transaction ID required", http.StatusBadRequest)
		return
	}

	var req updateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx := &model.Transaction{
		ID:          transactionID,
		CategoryID:  req.CategoryID,
		Description: req.Description,
	}

	if err := h.transactionService.UpdateTransaction(r.Context(), userID, tx); err != nil {
		http.Error(w, errors.Message(err), errors.StatusCode(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TransactionHandler) SyncTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, errors.Message(err), errors.StatusCode(err))
		return
	}

	if err := h.transactionService.SyncTransactions(r.Context(), userID); err != nil {
		http.Error(w, errors.Message(err), errors.StatusCode(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TransactionHandler) GetTransactionForm(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetTransactionForm: Starting to handle request")
	
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		log.Printf("GetTransactionForm: Failed to get userID: %v", err)
		http.Error(w, errors.Message(err), errors.StatusCode(err))
		return
	}
	log.Printf("GetTransactionForm: Got userID: %s", userID)

	// Get user's accounts for the form
	accounts, err := h.transactionService.GetUserAccounts(r.Context(), userID)
	if err != nil {
		log.Printf("GetTransactionForm: Failed to get user accounts: %v", err)
		http.Error(w, "Failed to get user accounts", http.StatusInternalServerError)
		return
	}
	log.Printf("GetTransactionForm: Got %d accounts", len(accounts))

	// Get categories for the form
	categories, err := h.transactionService.GetCategories(r.Context(), userID)
	if err != nil {
		log.Printf("GetTransactionForm: Failed to get categories: %v", err)
		http.Error(w, "Failed to get categories", http.StatusInternalServerError)
		return
	}
	log.Printf("GetTransactionForm: Got %d categories", len(categories))

	response := map[string]interface{}{
		"accounts":   accounts,
		"categories": categories,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("GetTransactionForm: Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	log.Printf("GetTransactionForm: Successfully sent response")
}

func (h *TransactionHandler) GetRecentTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user ID: %v", err), http.StatusUnauthorized)
		return
	}

	transactions, err := h.transactionService.GetRecentTransactions(r.Context(), userID, 5)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get recent transactions: %v", err), http.StatusInternalServerError)
		return
	}

	if transactions == nil {
		transactions = []*model.Transaction{} // Ensure we always return an array
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": transactions,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, errors.Message(err), errors.StatusCode(err))
		return
	}

	var tx model.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create the transaction
	if err := h.transactionService.CreateTransaction(r.Context(), userID, &tx); err != nil {
		log.Printf("Error creating transaction: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the created transaction
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Transaction created successfully",
		"data":    tx,
	})
}

// ServeHTTP implements the http.Handler interface
func (h *TransactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/recent"):
		h.GetRecentTransactions(w, r)
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/form"):
		h.GetTransactionForm(w, r)
	case r.Method == http.MethodGet:
		h.GetTransactions(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/sync"):
		h.SyncTransactions(w, r)
	case r.Method == http.MethodPost:
		h.CreateTransaction(w, r)
	case r.Method == http.MethodPut:
		h.UpdateTransaction(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
