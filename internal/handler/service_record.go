package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ilhamasy/gastrack-ms/internal/model"
	"github.com/ilhamasy/gastrack-ms/internal/service"
)

// ServiceRecordHandler handles service record CRUD operations.
type ServiceRecordHandler struct {
	serviceRecordService *service.ServiceRecordService
	vehicleService       *service.VehicleService
}

// NewServiceRecordHandler creates a new ServiceRecordHandler.
func NewServiceRecordHandler(srs *service.ServiceRecordService, vs *service.VehicleService) *ServiceRecordHandler {
	return &ServiceRecordHandler{
		serviceRecordService: srs,
		vehicleService:       vs,
	}
}

// AddServiceRecord creates a new service record for a vehicle.
func (h *ServiceRecordHandler) AddServiceRecord(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := extractVehicleIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, _, err := requireVehicleOwnership(w, r, h.vehicleService, vehicleID); err != nil {
		return
	}

	var req model.ServiceRecord
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ServiceDate.IsZero() {
		http.Error(w, "service_date is required", http.StatusBadRequest)
		return
	}
	if req.OdometerKm <= 0 {
		http.Error(w, "odometer_km must be positive", http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "at least one service item is required", http.StatusBadRequest)
		return
	}

	record, err := h.serviceRecordService.AddServiceRecord(r.Context(), vehicleID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

// GetServiceRecords returns all service records for a vehicle.
func (h *ServiceRecordHandler) GetServiceRecords(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := extractVehicleIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, _, err := requireVehicleOwnership(w, r, h.vehicleService, vehicleID); err != nil {
		return
	}

	records, err := h.serviceRecordService.GetServiceRecords(r.Context(), vehicleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if records == nil {
		records = []model.ServiceRecord{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

// ServeHTTP routes requests to the appropriate handler method.
func (h *ServiceRecordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetServiceRecords(w, r)
	case http.MethodPost:
		h.AddServiceRecord(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
