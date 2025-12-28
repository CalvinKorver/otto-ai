package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// ModelsService handles fetching and caching vehicle makes/models from auto.dev API
type ModelsService struct {
	mu          sync.RWMutex
	cachedData  map[string][]string
	lastFetched time.Time
	apiURL      string
}

// NewModelsService creates a new models service
func NewModelsService() *ModelsService {
	return &ModelsService{
		cachedData: make(map[string][]string),
		apiURL:     "https://auto.dev/api/models",
	}
}

// FetchModels fetches models from the auto.dev API
func (s *ModelsService) FetchModels() error {
	log.Println("Fetching vehicle models from auto.dev API...")

	resp, err := http.Get(s.apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var data map[string][]string
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Update cache with write lock
	s.mu.Lock()
	s.cachedData = data
	s.lastFetched = time.Now()
	s.mu.Unlock()

	log.Printf("Successfully fetched and cached vehicle models (last fetched: %s)", s.lastFetched.Format(time.RFC3339))
	return nil
}

// GetModels returns the cached models data
func (s *ModelsService) GetModels() (map[string][]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.cachedData) == 0 {
		return nil, fmt.Errorf("models data not yet loaded")
	}

	// Create a copy to avoid race conditions
	result := make(map[string][]string)
	for k, v := range s.cachedData {
		result[k] = make([]string, len(v))
		copy(result[k], v)
	}

	return result, nil
}

// GetMakes returns a list of all makes
func (s *ModelsService) GetMakes() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.cachedData) == 0 {
		return nil, fmt.Errorf("models data not yet loaded")
	}

	makes := make([]string, 0, len(s.cachedData))
	for make := range s.cachedData {
		makes = append(makes, make)
	}

	return makes, nil
}

// GetModelsForMake returns the models for a specific make
func (s *ModelsService) GetModelsForMake(make string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.cachedData) == 0 {
		return nil, fmt.Errorf("models data not yet loaded")
	}

	models, exists := s.cachedData[make]
	if !exists {
		return nil, fmt.Errorf("make not found: %s", make)
	}

	// Return a copy
	result := make([]string, len(models))
	copy(result, models)
	return result, nil
}

// StartDailyFetch starts a background goroutine that fetches models once per day
func (s *ModelsService) StartDailyFetch() {
	// Fetch immediately on startup
	if err := s.FetchModels(); err != nil {
		log.Printf("Warning: Failed to fetch models on startup: %v", err)
	}

	// Then fetch once per day
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			// Recover from panics to keep the goroutine running
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Panic in daily models fetch: %v", r)
					}
				}()

				if err := s.FetchModels(); err != nil {
					log.Printf("Error fetching models in daily update: %v", err)
					// Continue serving cached data if available
				}
			}()
		}
	}()
}

