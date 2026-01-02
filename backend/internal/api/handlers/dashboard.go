package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"carbuyer/internal/api/middleware"
	"carbuyer/internal/db"
	"carbuyer/internal/db/models"
	"carbuyer/internal/services"

	"github.com/google/uuid"
)

type DashboardHandler struct {
	threadService  *services.ThreadService
	messageService *services.MessageService
	db             *db.Database
}

func NewDashboardHandler(threadService *services.ThreadService, messageService *services.MessageService, database *db.Database) *DashboardHandler {
	return &DashboardHandler{
		threadService:  threadService,
		messageService: messageService,
		db:             database,
	}
}

// DashboardResponse represents the consolidated dashboard data
type DashboardResponse struct {
	Threads []ThreadResponse `json:"threads"`
	Offers  []OfferResponse  `json:"offers"`
}


// GetDashboard retrieves all dashboard data (threads, inbox messages, and offers) in a single request
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
		return
	}

	// Use goroutines to fetch all data in parallel
	var threadsWithCounts []services.ThreadWithCounts
	var offers []models.TrackedOffer
	var threadsErr, offersErr error

	var wg sync.WaitGroup
	wg.Add(2)

	// Fetch threads
	go func() {
		defer wg.Done()
		threadsWithCounts, threadsErr = h.threadService.GetUserThreads(userID)
	}()

	// Fetch offers
	go func() {
		defer wg.Done()
		offers, offersErr = h.fetchOffers(userID)
	}()

	wg.Wait()

	// Check for errors
	if threadsErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to fetch threads"})
		return
	}

	if offersErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to fetch offers"})
		return
	}

	// Build response
	response := DashboardResponse{
		Threads: make([]ThreadResponse, len(threadsWithCounts)),
		Offers:  make([]OfferResponse, len(offers)),
	}

	// Convert threads
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

	// Convert offers
	for i, offer := range offers {
		offerResp := OfferResponse{
			ID:        offer.ID.String(),
			ThreadID:  offer.ThreadID.String(),
			OfferText: offer.OfferText,
			TrackedAt: offer.TrackedAt.Format("2006-01-02T15:04:05Z"),
		}

		if offer.MessageID != nil {
			msgID := offer.MessageID.String()
			offerResp.MessageID = &msgID
		}

		// Include thread details if available
		if offer.Thread != nil {
			offerResp.SellerName = &offer.Thread.SellerName
			threadType := string(offer.Thread.SellerType)
			offerResp.ThreadType = &threadType
		}

		response.Offers[i] = offerResp
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}


// fetchOffers is a helper to fetch offers
func (h *DashboardHandler) fetchOffers(userID uuid.UUID) ([]models.TrackedOffer, error) {
	var offers []models.TrackedOffer

	// Get all offers for user's threads, ordered by most recent first
	err := h.db.DB.
		Joins("JOIN threads ON threads.id = tracked_offers.thread_id").
		Where("threads.user_id = ?", userID).
		Preload("Thread").
		Order("tracked_offers.tracked_at DESC").
		Find(&offers).Error

	if err != nil {
		return nil, err
	}

	return offers, nil
}


