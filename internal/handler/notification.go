package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yeboahd24/personal-finance-manager/internal/model"
	"github.com/yeboahd24/personal-finance-manager/internal/service"
)

type NotificationHandler struct {
	service *service.NotificationService
}

func NewNotificationHandler(service *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		service: service,
	}
}

// Add your notification endpoints here
func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	unreadOnly := r.URL.Query().Get("unread") == "true"

	notifications, err := h.service.GetUserNotifications(r.Context(), userID, unreadOnly)
	if err != nil {
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}

	respondJSON(w, notifications)
}

// CreateNotification handles the creation of a new notification
func (h *NotificationHandler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	var notificationReq service.NotificationRequest
	if err := decodeJSON(r, &notificationReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID").(string)
	notificationReq.UserID = userID

	// Convert request to model
	notification := &model.Notification{
		UserID:    notificationReq.UserID,
		Title:     notificationReq.Title,
		Message:   notificationReq.Message,
		Type:      model.NotificationType(notificationReq.Type),
		Priority:  model.NotificationPriority(notificationReq.Priority),
		Data:      notificationReq.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := h.service.CreateNotification(r.Context(), notification)
	if err != nil {
		http.Error(w, "Failed to create notification", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// UpdatePreferences handles updating user notification preferences
func (h *NotificationHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	var preferences model.NotificationPreferences
	if err := decodeJSON(r, &preferences); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID").(string)
	err := h.service.UpdateNotificationPreferences(r.Context(), userID, &preferences)
	if err != nil {
		http.Error(w, "Failed to update notification preferences", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// MarkAsRead handles marking a notification as read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	notificationID := r.URL.Query().Get("notification_id")
	
	if userID == "" || notificationID == "" {
		http.Error(w, "Missing user_id or notification_id parameter", http.StatusBadRequest)
		return
	}

	if err := h.service.MarkAsRead(r.Context(), userID, notificationID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to mark notification as read: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ServeHTTP implements the http.Handler interface
func (h *NotificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetNotifications(w, r)
	case http.MethodPost:
		if r.URL.Path == "/api/notifications/preferences" {
			h.UpdatePreferences(w, r)
		} else {
			h.CreateNotification(w, r)
		}
	case http.MethodPatch:
		h.MarkAsRead(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Helper functions
func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
