package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ilhamasy/gastrack-ms/internal/service"
)

// ExpenseHandler handles expense analytics endpoints.
type ExpenseHandler struct {
	expenseService *service.ExpenseService
	vehicleService *service.VehicleService
}

// NewExpenseHandler creates a new ExpenseHandler.
func NewExpenseHandler(expenseService *service.ExpenseService, vehicleService *service.VehicleService) *ExpenseHandler {
	return &ExpenseHandler{expenseService: expenseService, vehicleService: vehicleService}
}

// GetExpenseAnalytics returns expense analytics for a vehicle.
func (h *ExpenseHandler) GetExpenseAnalytics(w http.ResponseWriter, r *http.Request) {
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

	analytics, err := h.expenseService.GetExpenseAnalytics(r.Context(), vehicleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}
