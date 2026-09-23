package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/middleware"
	"github.com/ilhamasy/gastrack-ms/internal/service"
	"github.com/pashagolub/pgxmock/v4"
)

func TestOdometerHandler_LogOdometer(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	odoSvc := service.NewOdometerService(mock)
	h := NewOdometerHandler(odoSvc, nil)

	vID := uuid.New()
	uID := uuid.New()

	mock.ExpectBegin()

	// Ownership & current odo check
	mock.ExpectQuery("SELECT current_odometer FROM vehicles").
		WithArgs(vID, uID).
		WillReturnRows(pgxmock.NewRows([]string{"current_odometer"}).AddRow(10000))

	// Insert log
	mock.ExpectExec("INSERT INTO odometer_logs").
		WithArgs(vID, 12000).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Update vehicle odo
	mock.ExpectExec("UPDATE vehicles SET current_odometer").
		WithArgs(12000, vID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectCommit()

	body, _ := json.Marshal(map[string]int{"odometer_value": 12000})

	req := httptest.NewRequest(http.MethodPost, "/api/vehicles/"+vID.String()+"/odometer", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestOdometerHandler_GetOdometerHistory(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	odoSvc := service.NewOdometerService(mock)
	h := NewOdometerHandler(odoSvc, nil)

	vID := uuid.New()
	uID := uuid.New()
	logID := uuid.New()
	now := time.Now()

	// Ownership check
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(vID, uID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	// Get history
	mock.ExpectQuery("SELECT id, vehicle_id, odometer_value, recorded_at FROM odometer_logs").
		WithArgs(vID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "vehicle_id", "odometer_value", "recorded_at"}).
			AddRow(logID, vID, 12000, now))

	req := httptest.NewRequest(http.MethodGet, "/api/vehicles/"+vID.String()+"/odometer", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
