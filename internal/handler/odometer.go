package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/middleware"
	"github.com/ilhamasy/gastrack-ms/internal/service"
)

type OdometerHandler struct {
	odometerService *service.OdometerService
}

func NewOdometerHandler(odometerService *service.OdometerService) *OdometerHandler {
	return &OdometerHandler{odometerService: odometerService}
}

type LogOdometerRequest struct {
	OdometerValue int `json:"odometer_value"`
}

func (h *OdometerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/vehicles/"), "/")
	if len(parts) != 2 || parts[1] != "odometer" {
		http.NotFound(w, r)
		return
	}

	idStr := parts[0]
	vehicleID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	userIDStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid user ID in context")
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.handleLogOdometer(w, r, vehicleID, userID)
	case http.MethodGet:
		h.handleGetOdometerHistory(w, r, vehicleID, userID)
	default:
		w.Header().Set("Allow", "POST, GET")
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *OdometerHandler) handleLogOdometer(w http.ResponseWriter, r *http.Request, vehicleID, userID uuid.UUID) {
	var req LogOdometerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.OdometerValue < 0 {
		writeError(w, http.StatusBadRequest, "Odometer value cannot be negative")
		return
	}

	err := h.odometerService.LogOdometer(r.Context(), vehicleID, userID, req.OdometerValue)
	if err != nil {
		if err == service.ErrVehicleNotFound {
			writeError(w, http.StatusNotFound, "Vehicle not found")
			return
		}
		if err == service.ErrInvalidOdometer {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to log odometer")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": map[string]interface{}{
			"success": true,
		},
	})
}

func (h *OdometerHandler) handleGetOdometerHistory(w http.ResponseWriter, r *http.Request, vehicleID, userID uuid.UUID) {
	history, err := h.odometerService.GetOdometerHistory(r.Context(), vehicleID, userID)
	if err != nil {
		if err == service.ErrVehicleNotFound {
			writeError(w, http.StatusNotFound, "Vehicle not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to retrieve odometer history")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": history,
	})
}
