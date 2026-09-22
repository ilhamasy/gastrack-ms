package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/middleware"
	"github.com/ilhamasy/gastrack-ms/internal/service"
)

type VehicleHandler struct {
	vehicleService *service.VehicleService
}

func NewVehicleHandler(vehicleService *service.VehicleService) *VehicleHandler {
	return &VehicleHandler{vehicleService: vehicleService}
}

type AddVehicleRequest struct {
	Name            string `json:"name"`
	Make            string `json:"make"`
	Model           string `json:"model"`
	Variant         string `json:"variant"`
	Year            int    `json:"year"`
	CurrentOdometer int    `json:"current_odometer"`
}

func (h *VehicleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/vehicles":
		h.GetVehicles(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/api/vehicles":
		h.AddVehicle(w, r)
	case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/vehicles/"):
		h.UpdateVehicle(w, r)
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/vehicles/"):
		h.DeleteVehicle(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *VehicleHandler) GetVehicles(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	vehicles, err := h.vehicleService.GetVehicles(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch vehicles")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vehicles)
}

func (h *VehicleHandler) AddVehicle(w http.ResponseWriter, r *http.Request) {
	userIDStr, _ := r.Context().Value(middleware.UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)

	var req AddVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Make == "" || req.Model == "" || req.Year == 0 {
		writeError(w, http.StatusBadRequest, "make, model, and year are required")
		return
	}

	// Default name to Make + Model if not provided
	if req.Name == "" {
		req.Name = req.Make + " " + req.Model
	}

	vehicle := &service.Vehicle{
		UserID:          userID,
		Name:            req.Name,
		Make:            req.Make,
		Model:           req.Model,
		Variant:         req.Variant,
		Year:            req.Year,
		CurrentOdometer: req.CurrentOdometer,
	}

	if err := h.vehicleService.AddVehicle(r.Context(), vehicle); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add vehicle")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vehicle)
}

func (h *VehicleHandler) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	userIDStr, _ := r.Context().Value(middleware.UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)

	// Extract ID from URL path /api/vehicles/{id}
	pathParts := strings.Split(r.URL.Path, "/")
	idStr := pathParts[len(pathParts)-1]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vehicle id")
		return
	}

	var req AddVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Make == "" || req.Model == "" || req.Year == 0 {
		writeError(w, http.StatusBadRequest, "make, model, and year are required")
		return
	}
	if req.Name == "" {
		req.Name = req.Make + " " + req.Model
	}

	// We might need to fetch the existing vehicle first to preserve is_primary etc.
	// But let's assume UI sends is_primary or we just query it. For MVP, we can just set it to false and let the service handle it, or pass it via request.
	// We'll query first to be safe and merge.
	vehicles, err := h.vehicleService.GetVehicles(r.Context(), userID)
	var existing *service.Vehicle
	for _, v := range vehicles {
		if v.ID == id {
			existing = &v
			break
		}
	}

	if existing == nil {
		writeError(w, http.StatusNotFound, "vehicle not found")
		return
	}

	existing.Name = req.Name
	existing.Make = req.Make
	existing.Model = req.Model
	existing.Variant = req.Variant
	existing.Year = req.Year

	if err := h.vehicleService.UpdateVehicle(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update vehicle")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

func (h *VehicleHandler) DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	userIDStr, _ := r.Context().Value(middleware.UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)

	pathParts := strings.Split(r.URL.Path, "/")
	idStr := pathParts[len(pathParts)-1]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vehicle id")
		return
	}

	if err := h.vehicleService.DeleteVehicle(r.Context(), id, userID); err != nil {
		if err == service.ErrVehicleNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete vehicle")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
