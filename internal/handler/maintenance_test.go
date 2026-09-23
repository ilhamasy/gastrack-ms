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
	"github.com/ilhamasy/gastrack-ms/internal/model"
	"github.com/ilhamasy/gastrack-ms/internal/service"
	"github.com/pashagolub/pgxmock/v4"
)

func TestMaintenanceHandler_GetVehicleMaintenance(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mSvc := service.NewMaintenanceService(mock)
	vSvc := service.NewVehicleService(mock, nil)
	h := NewMaintenanceHandler(mSvc, vSvc)

	vID := uuid.New()
	uID := uuid.New()
	vIDStr := vID.String()

	// Ownership check
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(vID, uID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	// Get current odometer
	mock.ExpectQuery("SELECT current_odometer FROM vehicles").
		WithArgs(vIDStr).
		WillReturnRows(pgxmock.NewRows([]string{"current_odometer"}).AddRow(15000))

	// Maintenance items
	now := time.Now()
	mID := uuid.New()
	desc := "Engine oil"
	intervalKm := 10000
	intervalMonths := 6
	lastKm := 5000
	lastDate := time.Now().AddDate(0, -3, 0)

	mock.ExpectQuery("SELECT id, vehicle_id, template_id, name, description").
		WithArgs(vIDStr).
		WillReturnRows(pgxmock.NewRows([]string{"id", "vehicle_id", "template_id", "name", "description", "interval_km", "interval_months", "last_service_km", "last_service_date", "source", "created_at", "updated_at"}).
			AddRow(mID, vIDStr, nil, "Oil Change", &desc, &intervalKm, &intervalMonths, &lastKm, &lastDate, "USER_CREATED", now, now))

	req := httptest.NewRequest(http.MethodGet, "/api/vehicles/"+vIDStr+"/maintenance", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestMaintenanceHandler_AddCustomMaintenance(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mSvc := service.NewMaintenanceService(mock)
	vSvc := service.NewVehicleService(mock, nil)
	h := NewMaintenanceHandler(mSvc, vSvc)

	vID := uuid.New()
	uID := uuid.New()
	vIDStr := vID.String()

	// Ownership check
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(vID, uID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	desc := "Check spark plugs"
	intervalKm := 20000
	item := model.VehicleMaintenance{
		Name:        "Spark Plug Check",
		Description: &desc,
		IntervalKm:  &intervalKm,
	}
	body, _ := json.Marshal(item)

	mID := uuid.New()
	now := time.Now()
	mock.ExpectQuery("INSERT INTO vehicle_maintenance").
		WithArgs(vID, item.Name, item.Description, item.IntervalKm, item.IntervalMonths, item.LastServiceKm, item.LastServiceDate).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at", "source"}).
			AddRow(mID, now, now, "USER_CREATED"))

	req := httptest.NewRequest(http.MethodPost, "/api/vehicles/"+vIDStr+"/maintenance", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestMaintenanceHandler_UpdateMaintenance(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mSvc := service.NewMaintenanceService(mock)
	vSvc := service.NewVehicleService(mock, nil)
	h := NewMaintenanceHandler(mSvc, vSvc)

	vID := uuid.New()
	uID := uuid.New()
	mID := uuid.New()
	vIDStr := vID.String()
	mIDStr := mID.String()

	// Ownership check
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(vID, uID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	desc := "Updated coolant"
	intervalKm := 25000
	item := model.VehicleMaintenance{
		Name:        "Coolant Flush",
		Description: &desc,
		IntervalKm:  &intervalKm,
	}
	body, _ := json.Marshal(item)

	mock.ExpectQuery("SELECT source FROM vehicle_maintenance").
		WithArgs(mIDStr, vIDStr).
		WillReturnRows(pgxmock.NewRows([]string{"source"}).AddRow("TEMPLATE"))

	mock.ExpectExec("UPDATE vehicle_maintenance SET").
		WithArgs(item.Name, item.Description, item.IntervalKm, item.IntervalMonths, item.LastServiceKm, item.LastServiceDate, "USER_CUSTOMIZED", mIDStr, vIDStr).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	req := httptest.NewRequest(http.MethodPut, "/api/vehicles/"+vIDStr+"/maintenance/"+mIDStr, bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}
