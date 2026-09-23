package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ilhamasy/gastrack-ms/internal/service"
)

type RecommendationHandler struct {
	recommendationService *service.RecommendationService
}

func NewRecommendationHandler(recommendationService *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recommendationService: recommendationService}
}

func (h *RecommendationHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/vehicles/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "vehicle id is required", http.StatusBadRequest)
		return
	}
	vehicleID := parts[0]

	recs, err := h.recommendationService.GetRecommendations(r.Context(), vehicleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recs)
}
