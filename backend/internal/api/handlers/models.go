package handlers

import (
	"encoding/json"
	"net/http"

	"carbuyer/internal/services"
)

type ModelsHandler struct {
	modelsService *services.ModelsService
}

func NewModelsHandler(modelsService *services.ModelsService) *ModelsHandler {
	return &ModelsHandler{
		modelsService: modelsService,
	}
}

// GetModels returns all vehicle makes and models
func (h *ModelsHandler) GetModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.modelsService.GetModels()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Models data not yet available"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models)
}

