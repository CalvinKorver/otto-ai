package services

import (
	"errors"
	"fmt"
	"time"

	"carbuyer/internal/db/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ThreadService struct {
	db *gorm.DB
}

func NewThreadService(db *gorm.DB) *ThreadService {
	return &ThreadService{
		db: db,
	}
}

// CreateThread creates a new thread for a user
func (s *ThreadService) CreateThread(userID uuid.UUID, sellerName string, sellerType models.SellerType) (*models.Thread, error) {
	// Validate input
	if sellerName == "" {
		return nil, errors.New("seller name is required")
	}

	if sellerType != models.SellerTypePrivate && sellerType != models.SellerTypeDealership && sellerType != models.SellerTypeOther {
		return nil, errors.New("invalid seller type")
	}

	// Create thread
	thread := &models.Thread{
		UserID:     userID,
		SellerName: sellerName,
		SellerType: sellerType,
	}

	if err := s.db.Create(thread).Error; err != nil {
		return nil, fmt.Errorf("failed to create thread: %w", err)
	}

	return thread, nil
}

// ThreadWithCounts represents a thread with calculated counts and preview
type ThreadWithCounts struct {
	models.Thread
	MessageCount      int64  `json:"messageCount"`
	UnreadCount       int64  `json:"unreadCount"`
	LastMessagePreview string `json:"lastMessagePreview"`
	DisplayName       string `json:"displayName"`
}

// GetUserThreads retrieves all threads for a user with calculated counts and previews
func (s *ThreadService) GetUserThreads(userID uuid.UUID) ([]ThreadWithCounts, error) {
	var threads []models.Thread
	if err := s.db.Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("COALESCE(last_message_at, created_at) DESC").
		Find(&threads).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve threads: %w", err)
	}

	result := make([]ThreadWithCounts, len(threads))
	for i, thread := range threads {
		// Calculate message count
		var messageCount int64
		if err := s.db.Model(&models.Message{}).
			Where("thread_id = ? AND deleted_at IS NULL", thread.ID).
			Count(&messageCount).Error; err != nil {
			return nil, fmt.Errorf("failed to count messages: %w", err)
		}

		// Calculate unread count
		var unreadCount int64
		query := s.db.Model(&models.Message{}).
			Where("thread_id = ? AND deleted_at IS NULL", thread.ID)
		
		if thread.LastReadAt != nil {
			query = query.Where("timestamp > ?", thread.LastReadAt)
		}
		
		if err := query.Count(&unreadCount).Error; err != nil {
			return nil, fmt.Errorf("failed to count unread messages: %w", err)
		}

		// Get last message preview
		var lastMessage models.Message
		lastMessagePreview := ""
		if err := s.db.Where("thread_id = ? AND deleted_at IS NULL", thread.ID).
			Order("timestamp DESC").
			First(&lastMessage).Error; err == nil {
			// Truncate to ~50 characters
			preview := lastMessage.Content
			if len(preview) > 50 {
				preview = preview[:47] + "..."
			}
			lastMessagePreview = preview
		}

		// Calculate display name
		displayName := thread.SellerName
		if thread.SellerName == thread.Phone || thread.SellerName == "" {
			displayName = thread.Phone
		}

		result[i] = ThreadWithCounts{
			Thread:            thread,
			MessageCount:      messageCount,
			UnreadCount:       unreadCount,
			LastMessagePreview: lastMessagePreview,
			DisplayName:       displayName,
		}
	}

	return result, nil
}

// GetThreadByID retrieves a specific thread
func (s *ThreadService) GetThreadByID(threadID, userID uuid.UUID) (*models.Thread, error) {
	var thread models.Thread
	if err := s.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", threadID, userID).First(&thread).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("thread not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &thread, nil
}

// DeleteThread deletes a thread (soft delete in future)
func (s *ThreadService) DeleteThread(threadID, userID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", threadID, userID).Delete(&models.Thread{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete thread: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("thread not found")
	}

	return nil
}

// ArchiveThread soft deletes a thread by setting deleted_at
func (s *ThreadService) ArchiveThread(threadID, userID uuid.UUID) error {
	// Verify the thread exists, belongs to the user, and is not already archived
	var thread models.Thread
	if err := s.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", threadID, userID).First(&thread).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("thread not found")
		}
		return fmt.Errorf("failed to verify thread: %w", err)
	}

	// Soft delete the thread by setting deleted_at
	now := time.Now()
	thread.DeletedAt = &now
	if err := s.db.Save(&thread).Error; err != nil {
		return fmt.Errorf("failed to archive thread: %w", err)
	}

	return nil
}

// MarkThreadAsRead marks a thread as read by setting last_read_at to now
func (s *ThreadService) MarkThreadAsRead(threadID, userID uuid.UUID) error {
	// Verify the thread exists and belongs to the user
	var thread models.Thread
	if err := s.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", threadID, userID).First(&thread).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("thread not found")
		}
		return fmt.Errorf("failed to verify thread: %w", err)
	}

	// Set last_read_at to now
	now := time.Now()
	if err := s.db.Model(&models.Thread{}).Where("id = ?", threadID).Update("last_read_at", now).Error; err != nil {
		return fmt.Errorf("failed to mark thread as read: %w", err)
	}

	return nil
}
