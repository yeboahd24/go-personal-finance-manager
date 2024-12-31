package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/yeboahd24/personal-finance-manager/internal/middleware"
	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"github.com/yeboahd24/personal-finance-manager/internal/service"
)

type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

func (h *AccountHandler) CreateLinkToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	linkToken, err := h.accountService.CreateLinkToken(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"link_token": linkToken})
}

type linkAccountRequest struct {
	PublicToken string `json:"public_token"`
}

func (h *AccountHandler) LinkAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req linkAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.accountService.LinkAccount(r.Context(), userID, req.PublicToken); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accounts, err := h.accountService.GetAccounts(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get accounts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log the accounts being returned
	fmt.Printf("Returning accounts for user %s: %+v\n", userID, accounts)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"accounts": accounts,
	})
}

func (h *AccountHandler) SyncAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.accountService.SyncAccounts(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type createAccountRequest struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Received request to create account. Method: %s, Content-Type: %s", r.Method, r.Header.Get("Content-Type"))

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		log.Printf("Failed to get userID from context: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("Creating account for user: %s", userID)

	var req createAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Creating account for user %s with data: %+v", userID, req)

	account := &model.Account{
		UserID:   userID,
		Name:     req.Name,
		Type:     req.Type,
		Balance:  req.Balance,
		Currency: req.Currency,
	}

	if err := h.accountService.CreateAccount(r.Context(), account); err != nil {
		log.Printf("Failed to create account in database: %v", err)
		http.Error(w, "Failed to create account: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully created account: %+v", account)

	// Get all accounts for the user
	accounts, err := h.accountService.GetAccounts(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get accounts after creation: %v", err)
		http.Error(w, "Account created but failed to retrieve accounts", http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieved %d accounts for user %s", len(accounts), userID)

	// Set headers and status code before writing any data
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Changed to StatusOK since we're returning accounts
	
	response := map[string]interface{}{
		"message": "Account created successfully",
		"account": account,
		"accounts": accounts,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *AccountHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accountID := strings.TrimPrefix(r.URL.Path, "/api/accounts/")
	if accountID == "" {
		http.Error(w, "Account ID is required", http.StatusBadRequest)
		return
	}

	var req createAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	account := &model.Account{
		ID:       accountID,
		UserID:   userID,
		Name:     req.Name,
		Type:     req.Type,
		Balance:  req.Balance,
		Currency: req.Currency,
	}

	if err := h.accountService.UpdateAccount(r.Context(), account); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accountID := strings.TrimPrefix(r.URL.Path, "/api/accounts/")
	if accountID == "" {
		http.Error(w, "Account ID is required", http.StatusBadRequest)
		return
	}

	account, err := h.accountService.GetAccountByID(r.Context(), accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Verify the account belongs to the user
	if account.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Log the account retrieval
	fmt.Printf("Account retrieved for user %s: %+v\n", userID, account)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

func (h *AccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	
	switch {
	case path == "/api/accounts" && r.Method == http.MethodGet:
		h.GetAccounts(w, r)
	case path == "/api/accounts" && r.Method == http.MethodPost:
		h.CreateAccount(w, r)
	case strings.HasPrefix(path, "/api/accounts/") && r.Method == http.MethodGet:
		h.GetAccount(w, r)
	case strings.HasPrefix(path, "/api/accounts/") && r.Method == http.MethodPut:
		h.UpdateAccount(w, r)
	case path == "/api/accounts/link-token":
		h.CreateLinkToken(w, r)
	case path == "/api/accounts/link":
		h.LinkAccount(w, r)
	case path == "/api/accounts/sync":
		h.SyncAccounts(w, r)
	default:
		http.NotFound(w, r)
	}
}
