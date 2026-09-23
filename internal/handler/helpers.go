package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/middleware"
	"github.com/ilhamasy/gastrack-ms/internal/service"
)

// extractUserID extracts and parses the authenticated user ID from the request context.
func extractUserID(r *http.Request) (uuid.UUID, error) {
	userIDStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("unauthorized: missing user ID in context")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("unauthorized: invalid user ID format")
	}
	return userID, nil
}

// extractVehicleIDFromPath extracts the vehicle ID string from a URL path like /api/vehicles/{id}/...
func extractVehicleIDFromPath(path string) (string, error) {
	parts := strings.Split(strings.TrimPrefix(path, "/api/vehicles/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", fmt.Errorf("vehicle id is required")
	}
	return parts[0], nil
}

// requireVehicleOwnership verifies the authenticated user owns the specified vehicle.
// Returns nil on success, or writes an HTTP error response and returns an error.
func requireVehicleOwnership(w http.ResponseWriter, r *http.Request, vs *service.VehicleService, vehicleIDStr string) (uuid.UUID, uuid.UUID, error) {
	vID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		http.Error(w, "invalid vehicle id", http.StatusBadRequest)
		return uuid.Nil, uuid.Nil, err
	}

	userID, err := extractUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return uuid.Nil, uuid.Nil, err
	}

	isOwner, err := vs.IsVehicleOwner(r.Context(), vID, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return uuid.Nil, uuid.Nil, err
	}
	if !isOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return uuid.Nil, uuid.Nil, fmt.Errorf("forbidden")
	}

	return vID, userID, nil
}
