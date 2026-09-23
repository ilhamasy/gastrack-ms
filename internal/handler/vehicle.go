package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/service"
)

// VehicleHandler handles vehicle CRUD operations.
type VehicleHandler struct {
	vehicleService *service.VehicleService
}

// NewVehicleHandler creates a new VehicleHandler.
func NewVehicleHandler(vehicleService *service.VehicleService) *VehicleHandler {
	return &VehicleHandler{vehicleService: vehicleService}
}

// AddVehicleRequest represents the request body for creating or updating a vehicle.
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
	case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/primary"):
		h.SetPrimaryVehicle(w, r)
	case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/vehicles/"):
		h.UpdateVehicle(w, r)
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/vehicles/"):
		h.DeleteVehicle(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetVehicles returns all vehicles for the authenticated user.
func (h *VehicleHandler) GetVehicles(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
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

// AddVehicle creates a new vehicle for the authenticated user.
func (h *VehicleHandler) AddVehicle(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
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

// UpdateVehicle updates an existing vehicle.
func (h *VehicleHandler) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

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

	existing, err := h.vehicleService.GetVehicleByID(r.Context(), id, userID)
	if err != nil {
		if err == service.ErrVehicleNotFound {
			writeError(w, http.StatusNotFound, "vehicle not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch vehicle")
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

// DeleteVehicle removes a vehicle.
func (h *VehicleHandler) DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

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

// SetPrimaryVehicle marks a vehicle as the user's primary vehicle.
func (h *VehicleHandler) SetPrimaryVehicle(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 2 {
		writeError(w, http.StatusBadRequest, "invalid vehicle id")
		return
	}
	idStr := pathParts[len(pathParts)-2]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vehicle id")
		return
	}

	if err := h.vehicleService.SetPrimaryVehicle(r.Context(), id, userID); err != nil {
		if err == service.ErrVehicleNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to set primary vehicle")
		return
	}

	w.WriteHeader(http.StatusOK)
}
