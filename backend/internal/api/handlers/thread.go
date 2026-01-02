package handlers

import (
	"encoding/json"
	"net/http"

	"carbuyer/internal/api/middleware"
	"carbuyer/internal/db/models"
	"carbuyer/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ThreadHandler struct {
	threadService *services.ThreadService
}

func NewThreadHandler(threadService *services.ThreadService) *ThreadHandler {
	return &ThreadHandler{
		threadService: threadService,
	}
}

// CreateThreadRequest represents the request to create a thread
type CreateThreadRequest struct {
	SellerName string `json:"sellerName"`
	SellerType string `json:"sellerType"`
}

// ThreadResponse represents a thread in API responses
type ThreadResponse struct {
	ID                string  `json:"id"`
	SellerName        string  `json:"sellerName"`
	SellerType        string  `json:"sellerType"`
	Phone             string  `json:"phone,omitempty"`
	CreatedAt         string  `json:"createdAt"`
	LastMessageAt     *string `json:"lastMessageAt,omitempty"`
	MessageCount      int64   `json:"messageCount"`
	UnreadCount       int64   `json:"unreadCount"`
	LastMessagePreview string `json:"lastMessagePreview,omitempty"`
	DisplayName       string  `json:"displayName"`
}

// CreateThread creates a new thread
func (h *ThreadHandler) CreateThread(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
		return
	}

	var req CreateThreadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid request body"})
		return
	}

	// Convert seller type string to enum
	sellerType := models.SellerType(req.SellerType)

	thread, err := h.threadService.CreateThread(userID, req.SellerName, sellerType)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	// Calculate display name for response
	displayName := thread.SellerName
	if thread.SellerName == thread.Phone || thread.SellerName == "" {
		displayName = thread.Phone
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ThreadResponse{
		ID:          thread.ID.String(),
		SellerName:  thread.SellerName,
		SellerType:  string(thread.SellerType),
		Phone:       thread.Phone,
		CreatedAt:   thread.CreatedAt.Format("2006-01-02T15:04:05Z"),
		DisplayName: displayName,
		// MessageCount and UnreadCount will be 0 for new thread
		MessageCount: 0,
		UnreadCount:  0,
	})
}

// GetThreads retrieves all threads for the authenticated user
func (h *ThreadHandler) GetThreads(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
		return
	}

	threadsWithCounts, err := h.threadService.GetUserThreads(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to retrieve threads"})
		return
	}

	var response struct {
		Threads []ThreadResponse `json:"threads"`
	}

	response.Threads = make([]ThreadResponse, len(threadsWithCounts))
	for i, threadWithCounts := range threadsWithCounts {
		threadResp := ThreadResponse{
			ID:                threadWithCounts.ID.String(),
			SellerName:        threadWithCounts.SellerName,
			SellerType:        string(threadWithCounts.SellerType),
			Phone:             threadWithCounts.Phone,
			CreatedAt:         threadWithCounts.CreatedAt.Format("2006-01-02T15:04:05Z"),
			MessageCount:      threadWithCounts.MessageCount,
			UnreadCount:       threadWithCounts.UnreadCount,
			LastMessagePreview: threadWithCounts.LastMessagePreview,
			DisplayName:       threadWithCounts.DisplayName,
		}

		if threadWithCounts.LastMessageAt != nil {
			lastMsg := threadWithCounts.LastMessageAt.Format("2006-01-02T15:04:05Z")
			threadResp.LastMessageAt = &lastMsg
		}

		response.Threads[i] = threadResp
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetThread retrieves a specific thread
func (h *ThreadHandler) GetThread(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
		return
	}

	threadIDStr := chi.URLParam(r, "id")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid thread ID"})
		return
	}

	thread, err := h.threadService.GetThreadByID(threadID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if err.Error() == "thread not found" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	// Calculate display name
	displayName := thread.SellerName
	if thread.SellerName == thread.Phone || thread.SellerName == "" {
		displayName = thread.Phone
	}

	resp := ThreadResponse{
		ID:          thread.ID.String(),
		SellerName:  thread.SellerName,
		SellerType:  string(thread.SellerType),
		Phone:       thread.Phone,
		CreatedAt:   thread.CreatedAt.Format("2006-01-02T15:04:05Z"),
		DisplayName: displayName,
		// Note: MessageCount and UnreadCount would need to be calculated here if needed
		// For now, leaving them as 0 for single thread endpoint
		MessageCount: 0,
		UnreadCount:  0,
	}

	if thread.LastMessageAt != nil {
		lastMsg := thread.LastMessageAt.Format("2006-01-02T15:04:05Z")
		resp.LastMessageAt = &lastMsg
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// ArchiveThread archives (soft deletes) a thread
func (h *ThreadHandler) ArchiveThread(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
		return
	}

	threadIDStr := chi.URLParam(r, "id")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid thread ID"})
		return
	}

	err = h.threadService.ArchiveThread(threadID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if err.Error() == "thread not found" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "thread archived successfully"})
}

// MarkThreadAsRead marks a thread as read
func (h *ThreadHandler) MarkThreadAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
		return
	}

	threadIDStr := chi.URLParam(r, "id")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid thread ID"})
		return
	}

	err = h.threadService.MarkThreadAsRead(threadID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if err.Error() == "thread not found" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "thread marked as read"})
}
