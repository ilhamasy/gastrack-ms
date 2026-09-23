package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/middleware"
	"github.com/ilhamasy/gastrack-ms/internal/service"
)

type RecommendationHandler struct {
	recommendationService *service.RecommendationService
	vehicleService        *service.VehicleService
}

func NewRecommendationHandler(recommendationService *service.RecommendationService, vehicleService *service.VehicleService) *RecommendationHandler {
	return &RecommendationHandler{recommendationService: recommendationService, vehicleService: vehicleService}
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
	vID, err := uuid.Parse(vehicleID)
	if err != nil {
		http.Error(w, "invalid vehicle id", http.StatusBadRequest)
		return
	}

	userIDStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	isOwner, err := h.vehicleService.IsVehicleOwner(r.Context(), vID, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !isOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
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
