package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ilhamasy/gastrack-ms/internal/service"
)

type ExpenseHandler struct {
	expenseService *service.ExpenseService
}

func NewExpenseHandler(expenseService *service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseService: expenseService}
}

func (h *ExpenseHandler) GetExpenseAnalytics(w http.ResponseWriter, r *http.Request) {
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

	analytics, err := h.expenseService.GetExpenseAnalytics(r.Context(), vehicleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}
