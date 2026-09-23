package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/middleware"
	"github.com/ilhamasy/gastrack-ms/internal/model"
	"github.com/ilhamasy/gastrack-ms/internal/service"
)

type ServiceRecordHandler struct {
	serviceRecordService *service.ServiceRecordService
	vehicleService       *service.VehicleService
}

func NewServiceRecordHandler(srs *service.ServiceRecordService, vs *service.VehicleService) *ServiceRecordHandler {
	return &ServiceRecordHandler{
		serviceRecordService: srs,
		vehicleService:       vs,
	}
}

func (h *ServiceRecordHandler) AddServiceRecord(w http.ResponseWriter, r *http.Request) {
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

func (h *ServiceRecordHandler) GetServiceRecords(w http.ResponseWriter, r *http.Request) {
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
