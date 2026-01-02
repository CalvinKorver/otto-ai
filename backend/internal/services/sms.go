package services

import (
	"fmt"
	"os"
	"time"

	"carbuyer/internal/db/models"
	"carbuyer/internal/twilio"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SMSService struct {
	db            *gorm.DB
	twilioClient  *twilio.Client
}

func NewSMSService(db *gorm.DB, twilioClient *twilio.Client) *SMSService {
	return &SMSService{
		db:           db,
		twilioClient: twilioClient,
	}
}

// ProcessInboundSMS creates a message from an incoming SMS and auto-assigns it to a thread
func (s *SMSService) ProcessInboundSMS(userID uuid.UUID, from, body, messageSID string) (*models.Message, error) {
	// Get phone message type
	var phoneType models.MessageType
	if err := s.db.Where("type = ?", models.MessageTypePhone).First(&phoneType).Error; err != nil {
		return nil, fmt.Errorf("failed to get phone message type: %w", err)
	}

	// Check for duplicate based on external_message_id (Twilio message SID)
	if messageSID != "" {
		var existingMessage models.Message
		if err := s.db.Where("external_message_id = ?", messageSID).First(&existingMessage).Error; err == nil {
			// Message already exists, return it
			return &existingMessage, nil
		}
	}

	// Check if thread exists for this phone number
	var thread models.Thread
	err := s.db.Where("phone = ? AND user_id = ? AND deleted_at IS NULL", from, userID).First(&thread).Error
	
	var threadID *uuid.UUID
	if err == nil {
		// Thread exists, use it
		threadID = &thread.ID
	} else if err == gorm.ErrRecordNotFound {
		// Thread doesn't exist, create new one with phone number as name
		newThread := &models.Thread{
			UserID:     userID,
			SellerName: from, // Use phone number as name for unassigned threads
			SellerType: models.SellerTypeOther,
			Phone:      from,
			LastReadAt: nil, // New thread, all messages unread
		}
		if err := s.db.Create(newThread).Error; err != nil {
			return nil, fmt.Errorf("failed to create thread: %w", err)
		}
		threadID = &newThread.ID
	} else {
		return nil, fmt.Errorf("failed to check for existing thread: %w", err)
	}

	// Create message assigned to thread
	now := time.Now()
	message := &models.Message{
		UserID:            userID,
		ThreadID:          threadID,
		MessageTypeID:     &phoneType.ID,
		Sender:            models.SenderTypeSeller,
		Content:           body,
		Timestamp:         now,
		SenderPhone:       from,
		ExternalMessageID: messageSID,
	}

	// Save message and update thread's last_message_at in a transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(message).Error; err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}

		// Update thread's last_message_at
		if err := tx.Model(&models.Thread{}).Where("id = ?", *threadID).Update("last_message_at", now).Error; err != nil {
			return fmt.Errorf("failed to update thread last_message_at: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return message, nil
}

// SendSMS sends an SMS reply via Twilio
func (s *SMSService) SendSMS(userID uuid.UUID, threadID uuid.UUID, content string) error {
	// Get user's Twilio phone number
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.PhoneNumber == "" {
		return fmt.Errorf("user does not have a Twilio phone number allocated")
	}

	// Get thread to find dealer phone number
	var thread models.Thread
	if err := s.db.Where("id = ? AND user_id = ?", threadID, userID).First(&thread).Error; err != nil {
		return fmt.Errorf("thread not found: %w", err)
	}

	if thread.Phone == "" {
		return fmt.Errorf("thread does not have a phone number assigned")
	}

	// Send SMS via Twilio
	client := s.twilioClient.GetClient()
	messagingServiceSID := s.twilioClient.GetMessagingServiceSID()
	
	_, err := twilio.SendSMS(
		client,
		user.PhoneNumber, // from
		thread.Phone,     // to (dealer)
		content,
		messagingServiceSID,
	)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	// Get phone message type
	var phoneType models.MessageType
	if err := s.db.Where("type = ?", models.MessageTypePhone).First(&phoneType).Error; err != nil {
		return fmt.Errorf("failed to get phone message type: %w", err)
	}

	// Save outbound message to database
	message := &models.Message{
		UserID:        userID,
		ThreadID:      &threadID,
		MessageTypeID: &phoneType.ID,
		Sender:        models.SenderTypeAgent,
		Content:       content,
		Timestamp:      time.Now(),
		SenderPhone:   thread.Phone,
		SentViaSMS:    true,
	}

	if err := s.db.Create(message).Error; err != nil {
		return fmt.Errorf("failed to save outbound message: %w", err)
	}

	// Update thread last message time
	now := time.Now()
	if err := s.db.Model(&models.Thread{}).Where("id = ?", threadID).Update("last_message_at", now).Error; err != nil {
		return fmt.Errorf("failed to update thread: %w", err)
	}

	return nil
}

// AllocatePhoneNumber allocates a Twilio phone number for a user
func (s *SMSService) AllocatePhoneNumber(userID uuid.UUID) error {
	// #region agent log
	func() {
		logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"C","location":"sms.go:136","message":"AllocatePhoneNumber called","data":{"userID":"%s","twilioClientIsNil":%t},"timestamp":%d}`, userID.String(), s.twilioClient == nil, time.Now().UnixMilli())
		if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintln(f, logData)
			f.Close()
		}
	}()
	// #endregion

	// Check if user already has a phone number
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		// #region agent log
		func() {
			logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"D","location":"sms.go:140","message":"User lookup failed","data":{"userID":"%s","error":"%s"},"timestamp":%d}`, userID.String(), err.Error(), time.Now().UnixMilli())
			if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				fmt.Fprintln(f, logData)
				f.Close()
			}
		}()
		// #endregion
		return fmt.Errorf("user not found: %w", err)
	}

	// #region agent log
	func() {
		logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"D","location":"sms.go:150","message":"User found, checking existing phone","data":{"userID":"%s","hasPhoneNumber":%t,"phoneNumber":"%s"},"timestamp":%d}`, userID.String(), user.PhoneNumber != "", user.PhoneNumber, time.Now().UnixMilli())
		if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintln(f, logData)
			f.Close()
		}
	}()
	// #endregion

	if user.PhoneNumber != "" {
		// User already has a phone number
		return nil
	}

	// Try to get area code from user's zip code if available
	areaCode := ""
	if user.ZipCode != "" {
		// For now, we'll use a default area code or try to extract from zip
		// In a real implementation, you might want to map zip codes to area codes
		// For simplicity, we'll search for numbers in a common area code
		areaCode = "415" // Default to San Francisco area code
	} else {
		areaCode = "415" // Default area code
	}

	// #region agent log
	func() {
		logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"E","location":"sms.go:160","message":"Before Twilio API calls","data":{"userID":"%s","areaCode":"%s","clientIsNil":%t,"messagingServiceSID":"%s"},"timestamp":%d}`, userID.String(), areaCode, s.twilioClient == nil, s.twilioClient.GetMessagingServiceSID(), time.Now().UnixMilli())
		if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintln(f, logData)
			f.Close()
		}
	}()
	// #endregion

	// Search for available numbers
	client := s.twilioClient.GetClient()
	availableNumbers, err := twilio.SearchAvailableNumbers(client, areaCode, 1)
	if err != nil {
		// #region agent log
		func() {
			logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"E","location":"sms.go:165","message":"SearchAvailableNumbers failed","data":{"userID":"%s","areaCode":"%s","error":"%s"},"timestamp":%d}`, userID.String(), areaCode, err.Error(), time.Now().UnixMilli())
			if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				fmt.Fprintln(f, logData)
				f.Close()
			}
		}()
		// #endregion
		return fmt.Errorf("failed to search available numbers: %w", err)
	}

	// #region agent log
	func() {
		logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"E","location":"sms.go:175","message":"SearchAvailableNumbers succeeded","data":{"userID":"%s","availableCount":%d},"timestamp":%d}`, userID.String(), len(availableNumbers), time.Now().UnixMilli())
		if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintln(f, logData)
			f.Close()
		}
	}()
	// #endregion

	if len(availableNumbers) == 0 {
		return fmt.Errorf("no available phone numbers found for area code %s", areaCode)
	}

	// Purchase the first available number
	phoneNumber := *availableNumbers[0].PhoneNumber
	messagingServiceSID := s.twilioClient.GetMessagingServiceSID()
	
	// #region agent log
	func() {
		logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"E","location":"sms.go:183","message":"Before PurchaseNumber","data":{"userID":"%s","phoneNumber":"%s","messagingServiceSID":"%s"},"timestamp":%d}`, userID.String(), phoneNumber, messagingServiceSID, time.Now().UnixMilli())
		if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintln(f, logData)
			f.Close()
		}
	}()
	// #endregion

	twilioSID, err := twilio.PurchaseNumber(client, phoneNumber, messagingServiceSID)
	if err != nil {
		// #region agent log
		func() {
			logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"E","location":"sms.go:190","message":"PurchaseNumber failed","data":{"userID":"%s","phoneNumber":"%s","error":"%s"},"timestamp":%d}`, userID.String(), phoneNumber, err.Error(), time.Now().UnixMilli())
			if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				fmt.Fprintln(f, logData)
				f.Close()
			}
		}()
		// #endregion
		return fmt.Errorf("failed to purchase number: %w", err)
	}

	// #region agent log
	func() {
		logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"E","location":"sms.go:198","message":"PurchaseNumber succeeded","data":{"userID":"%s","phoneNumber":"%s","twilioSID":"%s"},"timestamp":%d}`, userID.String(), phoneNumber, twilioSID, time.Now().UnixMilli())
		if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintln(f, logData)
			f.Close()
		}
	}()
	// #endregion

	// #region agent log
	func() {
		logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"F","location":"sms.go:200","message":"Before database save","data":{"userID":"%s","phoneNumber":"%s","twilioSID":"%s"},"timestamp":%d}`, userID.String(), phoneNumber, twilioSID, time.Now().UnixMilli())
		if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintln(f, logData)
			f.Close()
		}
	}()
	// #endregion

	// Update user with phone number and Twilio SID
	user.PhoneNumber = phoneNumber
	user.TwilioSID = twilioSID
	if err := s.db.Save(&user).Error; err != nil {
		// #region agent log
		func() {
			logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"F","location":"sms.go:207","message":"Database save failed","data":{"userID":"%s","phoneNumber":"%s","error":"%s"},"timestamp":%d}`, userID.String(), phoneNumber, err.Error(), time.Now().UnixMilli())
			if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				fmt.Fprintln(f, logData)
				f.Close()
			}
		}()
		// #endregion
		return fmt.Errorf("failed to save phone number to user: %w", err)
	}

	// #region agent log
	func() {
		logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"F","location":"sms.go:217","message":"AllocatePhoneNumber completed successfully","data":{"userID":"%s","phoneNumber":"%s","twilioSID":"%s"},"timestamp":%d}`, userID.String(), phoneNumber, twilioSID, time.Now().UnixMilli())
		if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintln(f, logData)
			f.Close()
		}
	}()
	// #endregion

	return nil
}

