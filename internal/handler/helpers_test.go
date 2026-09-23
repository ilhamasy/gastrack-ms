package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/middleware"
	"github.com/ilhamasy/gastrack-ms/internal/service"
	"github.com/pashagolub/pgxmock/v4"
)

func TestExtractUserID_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, err := extractUserID(req)
	if err == nil {
		t.Error("expected error for missing user_id in context")
	}
}

func TestExtractUserID_InvalidUUID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "not-a-uuid")
	req = req.WithContext(ctx)

	_, err := extractUserID(req)
	if err == nil {
		t.Error("expected error for invalid user_id UUID string")
	}
}

func TestExtractVehicleIDFromPath_Empty(t *testing.T) {
	_, err := extractVehicleIDFromPath("/api/vehicles/")
	if err == nil {
		t.Error("expected error for empty vehicle id path")
	}
}

func TestRequireVehicleOwnership_InvalidVehicleID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/vehicles/invalid-uuid/items", nil)
	rec := httptest.NewRecorder()

	_, _, err := requireVehicleOwnership(rec, req, nil, "invalid-uuid")
	if err == nil {
		t.Error("expected error")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestRequireVehicleOwnership_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/vehicles/550e8400-e29b-41d4-a716-446655440000/items", nil)
	rec := httptest.NewRecorder()

	_, _, err := requireVehicleOwnership(rec, req, nil, "550e8400-e29b-41d4-a716-446655440000")
	if err == nil {
		t.Error("expected error")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireVehicleOwnership_Forbidden(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	vs := service.NewVehicleService(mock, nil)

	uID := uuid.New()
	vID := uuid.New()

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(vID, uID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))

	req := httptest.NewRequest(http.MethodGet, "/api/vehicles/"+vID.String()+"/items", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	_, _, err = requireVehicleOwnership(rec, req, vs, vID.String())
	if err == nil {
		t.Error("expected error for non-owner")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
