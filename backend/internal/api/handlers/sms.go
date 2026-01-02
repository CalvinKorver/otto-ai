package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"carbuyer/internal/api/middleware"
	"carbuyer/internal/db/models"
	"carbuyer/internal/services"
	"carbuyer/internal/twilio"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SMSHandler struct {
	db           *gorm.DB
	smsService   *services.SMSService
	twilioClient *twilio.Client
	authToken    string
}

func NewSMSHandler(db *gorm.DB, smsService *services.SMSService, twilioClient *twilio.Client, authToken string) *SMSHandler {
	return &SMSHandler{
		db:           db,
		smsService:   smsService,
		twilioClient: twilioClient,
		authToken:    authToken,
	}
}

// InboundSMS handles incoming SMS webhooks from Twilio
// POST /api/v1/webhooks/sms/inbound
func (h *SMSHandler) InboundSMS(w http.ResponseWriter, r *http.Request) {
	// #region agent log
	func() {
		logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"webhook-debug","hypothesisId":"Z","location":"sms.go:37","message":"InboundSMS handler called","data":{"method":"%s","path":"%s","remoteAddr":"%s","userAgent":"%s","contentType":"%s"},"timestamp":%d}`, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent(), r.Header.Get("Content-Type"), time.Now().UnixMilli())
		if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintln(f, logData)
			f.Close()
		}
	}()
	// #endregion

	// Parse form data (Twilio sends webhooks as form-encoded)
	if err := r.ParseForm(); err != nil {
		// #region agent log
		func() {
			logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"webhook-debug","hypothesisId":"Z","location":"sms.go:45","message":"Failed to parse form","data":{"error":"%s"},"timestamp":%d}`, err.Error(), time.Now().UnixMilli())
			if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				fmt.Fprintln(f, logData)
				f.Close()
			}
		}()
		// #endregion
		log.Printf("Failed to parse form: %v", err)
		// Return 200 OK even on parse error to prevent Twilio retries
		// Log the error internally but don't expose it to Twilio
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{})
		return
	}

	// Validate webhook signature for security
	signature := r.Header.Get("X-Twilio-Signature")

	// #region agent log
	func() {
		logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"webhook-debug","hypothesisId":"A","location":"sms.go:48","message":"Webhook received, checking signature","data":{"hasSignature":%t,"hasAuthToken":%t,"authTokenLength":%d,"method":"%s","path":"%s","host":"%s","forwardedProto":"%s","forwardedHost":"%s","rawURL":"%s"},"timestamp":%d}`, signature != "", h.authToken != "", len(h.authToken), r.Method, r.URL.Path, r.Host, r.Header.Get("X-Forwarded-Proto"), r.Header.Get("X-Forwarded-Host"), r.URL.String(), time.Now().UnixMilli())
		if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintln(f, logData)
			f.Close()
		}
	}()
	// #endregion

	// For development/testing: allow skipping signature validation if auth token is not set
	// In production, this should always be validated
	if signature != "" && h.authToken != "" {
		// Build URL for signature validation
		// Handle proxy headers (X-Forwarded-Proto, X-Forwarded-Host) for ngrok/local dev
		scheme := "https"
		if r.Header.Get("X-Forwarded-Proto") != "" {
			scheme = r.Header.Get("X-Forwarded-Proto")
		} else if r.TLS == nil {
			scheme = "http"
		}

		host := r.Host
		if r.Header.Get("X-Forwarded-Host") != "" {
			host = r.Header.Get("X-Forwarded-Host")
		}

		// Build URL - Twilio signs using the full URL they call
		// Use Path (not RequestURI) as query params are in the form data, not URL
		webhookURL := scheme + "://" + host + r.URL.Path

		// Convert form values to map for validation
		params := make(map[string]string)
		for k, v := range r.Form {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}

		// #region agent log
		func() {
			paramKeys := make([]string, 0, len(params))
			for k := range params {
				paramKeys = append(paramKeys, k)
			}
			logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"webhook-debug","hypothesisId":"B","location":"sms.go:75","message":"Before signature validation","data":{"webhookURL":"%s","paramCount":%d,"paramKeys":%v,"signatureLength":%d},"timestamp":%d}`, webhookURL, len(params), paramKeys, len(signature), time.Now().UnixMilli())
			if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				fmt.Fprintln(f, logData)
				f.Close()
			}
		}()
		// #endregion

		// Validate signature
		isValid := twilio.ValidateWebhookSignature(webhookURL, params, signature, h.authToken)

		// #region agent log
		func() {
			logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"webhook-debug","hypothesisId":"B","location":"sms.go:87","message":"Signature validation result","data":{"isValid":%t,"webhookURL":"%s"},"timestamp":%d}`, isValid, webhookURL, time.Now().UnixMilli())
			if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				fmt.Fprintln(f, logData)
				f.Close()
			}
		}()
		// #endregion

		if !isValid {
			// #region agent log
			func() {
				logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"webhook-debug","hypothesisId":"C","location":"sms.go:95","message":"Signature validation FAILED","data":{"webhookURL":"%s","signature":"%s","authTokenLength":%d,"paramCount":%d},"timestamp":%d}`, webhookURL, signature, len(h.authToken), len(params), time.Now().UnixMilli())
				if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
					fmt.Fprintln(f, logData)
					f.Close()
				}
			}()
			// #endregion
			log.Printf("Invalid Twilio signature - URL: %s, Signature present: %t, AuthToken present: %t, AuthToken length: %d", webhookURL, signature != "", h.authToken != "", len(h.authToken))
			// Return 401 for invalid signature - this is a security issue
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte{})
			return
		}
		// #region agent log
		func() {
			logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"webhook-debug","hypothesisId":"C","location":"sms.go:105","message":"Signature validation SUCCESS","data":{"webhookURL":"%s"},"timestamp":%d}`, webhookURL, time.Now().UnixMilli())
			if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				fmt.Fprintln(f, logData)
				f.Close()
			}
		}()
		// #endregion
		log.Printf("Twilio signature validated successfully")
	} else if signature == "" {
		log.Printf("Warning: Missing X-Twilio-Signature header (webhook may be insecure)")
	} else if h.authToken == "" {
		log.Printf("Warning: Auth token not configured, skipping signature validation")
	}

	// Extract webhook data
	to := r.FormValue("To")     // The Twilio number that received the SMS
	from := r.FormValue("From") // The dealer's phone number
	body := r.FormValue("Body") // The message content
	messageSID := r.FormValue("MessageSid")

	// #region agent log
	func() {
		logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"webhook-debug","hypothesisId":"D","location":"sms.go:151","message":"Extracted webhook data","data":{"to":"%s","from":"%s","bodyLength":%d,"messageSID":"%s"},"timestamp":%d}`, to, from, len(body), messageSID, time.Now().UnixMilli())
		if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintln(f, logData)
			f.Close()
		}
	}()
	// #endregion

	log.Printf("Inbound SMS - To: %s, From: %s, Body: %s", to, from, body)

	// Lookup user by Twilio phone number
	var user models.User
	if err := h.db.Where("phone_number = ?", to).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// #region agent log
			func() {
				logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"webhook-debug","hypothesisId":"E","location":"sms.go:165","message":"User not found for phone number","data":{"to":"%s"},"timestamp":%d}`, to, time.Now().UnixMilli())
				if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
					fmt.Fprintln(f, logData)
					f.Close()
				}
			}()
			// #endregion
			// User not found - return 200 OK with empty body to prevent retries
			log.Printf("User not found for Twilio number: %s", to)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte{})
			return
		}
		// #region agent log
		func() {
			logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"webhook-debug","hypothesisId":"E","location":"sms.go:177","message":"Database error looking up user","data":{"to":"%s","error":"%s"},"timestamp":%d}`, to, err.Error(), time.Now().UnixMilli())
			if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				fmt.Fprintln(f, logData)
				f.Close()
			}
		}()
		// #endregion
		log.Printf("Database error: %v", err)
		// Return 200 OK even on error to prevent Twilio retries
		// Log the error internally but don't expose it to Twilio
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{})
		return
	}

	// #region agent log
	func() {
		logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"webhook-debug","hypothesisId":"F","location":"sms.go:190","message":"User found, processing SMS","data":{"userID":"%s","to":"%s","from":"%s"},"timestamp":%d}`, user.ID.String(), to, from, time.Now().UnixMilli())
		if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintln(f, logData)
			f.Close()
		}
	}()
	// #endregion

	// Process inbound SMS
	message, err := h.smsService.ProcessInboundSMS(user.ID, from, body, messageSID)
	if err != nil {
		// #region agent log
		func() {
			logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"webhook-debug","hypothesisId":"G","location":"sms.go:200","message":"ProcessInboundSMS failed","data":{"userID":"%s","error":"%s"},"timestamp":%d}`, user.ID.String(), err.Error(), time.Now().UnixMilli())
			if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				fmt.Fprintln(f, logData)
				f.Close()
			}
		}()
		// #endregion
		log.Printf("Failed to process inbound SMS: %v", err)
		// Return 200 OK even on error to prevent Twilio retries
		// Log the error internally but don't expose it to Twilio
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{})
		return
	}

	// #region agent log
	func() {
		logData := fmt.Sprintf(`{"sessionId":"debug-session","runId":"webhook-debug","hypothesisId":"H","location":"sms.go:215","message":"SMS processed successfully, returning 200","data":{"messageID":"%s"},"timestamp":%d}`, message.ID.String(), time.Now().UnixMilli())
		if f, err := os.OpenFile("/Users/calvinkorver/car-buyer/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintln(f, logData)
			f.Close()
		}
	}()
	// #endregion

	// Return 200 OK (critical for Twilio)
	// For SMS webhooks, Twilio expects either TwiML (text/xml) or just 200 OK with no body
	// Returning empty 200 OK is the standard for SMS webhooks
	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}

// SendSMS handles sending SMS replies
// POST /api/v1/messages/{messageId}/sms-reply
func (h *SMSHandler) SendSMS(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
		return
	}

	messageIDStr := chi.URLParam(r, "messageId")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid message ID"})
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Content == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "content is required"})
		return
	}

	// Get the message to find the thread
	var message models.Message
	if err := h.db.Where("id = ? AND user_id = ?", messageID, userID).First(&message).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "message not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "database error"})
		return
	}

	if message.ThreadID == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "message is not assigned to a thread"})
		return
	}

	// Send SMS via service
	if err := h.smsService.SendSMS(userID, *message.ThreadID, req.Content); err != nil {
		w.Header().Set("Content-Type", "application/json")
		errMsg := err.Error()
		if errMsg == "user not found" || errMsg == "thread not found" {
			w.WriteHeader(http.StatusNotFound)
		} else if errMsg == "user does not have a Twilio phone number allocated" {
			w.WriteHeader(http.StatusBadRequest)
		} else if errMsg == "thread does not have a phone number assigned" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "SMS sent successfully"})
}

// GetPhoneNumber returns the user's allocated phone number
// GET /api/v1/sms/phone-number
func (h *SMSHandler) GetPhoneNumber(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
		return
	}

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "user not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "database error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"phoneNumber": user.PhoneNumber,
	})
}
