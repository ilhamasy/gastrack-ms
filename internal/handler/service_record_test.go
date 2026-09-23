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

func TestServiceRecordHandler_AddServiceRecord(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	srSvc := service.NewServiceRecordService(mock)
	vSvc := service.NewVehicleService(mock, nil)
	h := NewServiceRecordHandler(srSvc, vSvc)

	vID := uuid.New()
	uID := uuid.New()
	vIDStr := vID.String()

	// Ownership check
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(vID, uID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	notes := "Regular oil change"
	record := model.ServiceRecord{
		ServiceDate: time.Now(),
		OdometerKm:  15000,
		Notes:       &notes,
		TotalCost:   150000,
		Items: []model.ServiceItem{
			{ItemName: "Engine Oil", Cost: 150000, Quantity: 1},
		},
	}
	body, _ := json.Marshal(record)

	recID := uuid.New()
	itemID := uuid.New()
	now := time.Now()

	mock.ExpectBegin()

	// Insert record
	mock.ExpectQuery("INSERT INTO service_records").
		WithArgs(pgxmock.AnyArg(), vIDStr, pgxmock.AnyArg(), record.OdometerKm, record.WorkshopName, record.TotalCost, record.Notes).
		WillReturnRows(pgxmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))

	// Insert item
	mock.ExpectQuery("INSERT INTO service_items").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), "Engine Oil", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), 1.0, 150000.0, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"created_at"}).AddRow(now))

	// Update vehicle odometer
	mock.ExpectExec("UPDATE vehicles SET current_odometer").
		WithArgs(record.OdometerKm, vIDStr).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodPost, "/api/vehicles/"+vIDStr+"/service-records", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	_ = recID
	_ = itemID
}

func TestServiceRecordHandler_GetServiceRecords(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	srSvc := service.NewServiceRecordService(mock)
	vSvc := service.NewVehicleService(mock, nil)
	h := NewServiceRecordHandler(srSvc, vSvc)

	vID := uuid.New()
	uID := uuid.New()
	vIDStr := vID.String()

	// Ownership check
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(vID, uID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	now := time.Now()
	recID := uuid.New()
	itemID := uuid.New()
	notes := "Regular oil change"

	mock.ExpectQuery("SELECT r.id, r.vehicle_id").
		WithArgs(vIDStr).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "vehicle_id", "service_date", "odometer_km", "workshop_name", "total_cost", "notes", "created_at", "updated_at",
			"item_id", "maintenance_id", "item_name", "brand", "product", "part_number", "quantity", "cost", "item_notes", "item_created_at",
		}).AddRow(
			recID, vID, now, 15000, nil, 150000.0, &notes, now, now,
			&itemID, nil, "Engine Oil", nil, nil, nil, 1.0, 150000.0, nil, now,
		))

	req := httptest.NewRequest(http.MethodGet, "/api/vehicles/"+vIDStr+"/service-records", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
