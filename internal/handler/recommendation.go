package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ilhamasy/gastrack-ms/internal/service"
)

// RecommendationHandler handles recommendation endpoints.
type RecommendationHandler struct {
	recommendationService *service.RecommendationService
	vehicleService        *service.VehicleService
}

// NewRecommendationHandler creates a new RecommendationHandler.
func NewRecommendationHandler(recommendationService *service.RecommendationService, vehicleService *service.VehicleService) *RecommendationHandler {
	return &RecommendationHandler{recommendationService: recommendationService, vehicleService: vehicleService}
}

// GetRecommendations returns maintenance recommendations for a vehicle.
func (h *RecommendationHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vehicleID, err := extractVehicleIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, _, err := requireVehicleOwnership(w, r, h.vehicleService, vehicleID); err != nil {
		return
	}

	recs, err := h.recommendationService.GetRecommendations(r.Context(), vehicleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recs)
}
