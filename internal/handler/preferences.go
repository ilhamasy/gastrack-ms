package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ilhamasy/gastrack-ms/internal/model"
	"github.com/ilhamasy/gastrack-ms/internal/service"
)

type PreferencesHandler struct {
	preferencesService *service.PreferencesService
}

func NewPreferencesHandler(preferencesService *service.PreferencesService) *PreferencesHandler {
	return &PreferencesHandler{preferencesService: preferencesService}
}

func (h *PreferencesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	userID := r.Context().Value("user_id").(string)

	switch r.Method {
	case http.MethodGet:
		prefs, err := h.preferencesService.GetPreferences(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prefs)
	case http.MethodPut:
		var prefs model.NotificationPreferences
		if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if err := h.preferencesService.UpdatePreferences(r.Context(), userID, &prefs); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prefs)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
