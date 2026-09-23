package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/model"
	"github.com/ilhamasy/gastrack-ms/internal/service"
)

type MaintenanceHandler struct {
	maintenanceService *service.MaintenanceService
	vehicleService     *service.VehicleService
}

func NewMaintenanceHandler(ms *service.MaintenanceService, vs *service.VehicleService) *MaintenanceHandler {
	return &MaintenanceHandler{
		maintenanceService: ms,
		vehicleService:     vs,
	}
}

func (h *MaintenanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// r.URL.Path typically: /api/vehicles/{id}/maintenance or /api/vehicles/{id}/maintenance/{maintenance_id}
	// Let's rely on Go 1.22 routing by not using ServeHTTP, but since we are called from the closure:
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/vehicles/"), "/")
	if len(parts) < 2 || parts[1] != "maintenance" {
		http.NotFound(w, r)
		return
	}

	// We need to inject the PathValues so the existing functions work. 
	// Since we are using Go 1.22 PathValue in GetVehicleMaintenance, but ServeHTTP strips it,
	// actually Go 1.22 routing natively matches paths if we define them globally.
	// Wait, the closure in main.go intercepts the requests. Let's manually parse or set path values.
	// Actually, the closure intercept is for Go 1.21 style. If I just route it manually here:
	vehicleID := parts[0]
	r.SetPathValue("id", vehicleID)

	if len(parts) == 2 {
		if r.Method == http.MethodGet {
			h.GetVehicleMaintenance(w, r)
		} else if r.Method == http.MethodPost {
			h.AddCustomMaintenance(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if len(parts) == 3 {
		r.SetPathValue("maintenance_id", parts[2])
		if r.Method == http.MethodPut {
			h.UpdateMaintenance(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	http.NotFound(w, r)
}

// GetVehicleMaintenance handles GET /api/vehicles/{id}/maintenance
func (h *MaintenanceHandler) GetVehicleMaintenance(w http.ResponseWriter, r *http.Request) {
	vehicleID := r.PathValue("id")
	if _, err := uuid.Parse(vehicleID); err != nil {
		http.Error(w, "invalid vehicle id", http.StatusBadRequest)
		return
	}

	// Wait, we should probably check if the user owns the vehicle
	// But since auth isn't fully locking down routes yet (GT-021 is for isolation), we just query it.
	items, err := h.maintenanceService.GetVehicleMaintenance(r.Context(), vehicleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// AddCustomMaintenance handles POST /api/vehicles/{id}/maintenance
func (h *MaintenanceHandler) AddCustomMaintenance(w http.ResponseWriter, r *http.Request) {
	vehicleID := r.PathValue("id")
	vID, err := uuid.Parse(vehicleID)
	if err != nil {
		http.Error(w, "invalid vehicle id", http.StatusBadRequest)
		return
	}

	var req model.VehicleMaintenance
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.VehicleID = vID

	if err := h.maintenanceService.AddCustomMaintenanceItem(r.Context(), &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

// UpdateMaintenance handles PUT /api/vehicles/{id}/maintenance/{maintenance_id}
func (h *MaintenanceHandler) UpdateMaintenance(w http.ResponseWriter, r *http.Request) {
	vehicleID := r.PathValue("id")
	maintenanceID := r.PathValue("maintenance_id")

	if _, err := uuid.Parse(vehicleID); err != nil {
		http.Error(w, "invalid vehicle id", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(maintenanceID); err != nil {
		http.Error(w, "invalid maintenance id", http.StatusBadRequest)
		return
	}

	var req model.VehicleMaintenance
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.maintenanceService.UpdateMaintenanceItem(r.Context(), vehicleID, maintenanceID, &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch the updated item to return
	items, _ := h.maintenanceService.GetVehicleMaintenance(r.Context(), vehicleID)
	var updatedItem model.VehicleMaintenance
	for _, item := range items {
		if item.ID.String() == maintenanceID {
			updatedItem = item
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedItem)
}
