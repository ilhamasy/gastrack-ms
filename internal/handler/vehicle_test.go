package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/middleware"
	"github.com/ilhamasy/gastrack-ms/internal/service"
	"github.com/pashagolub/pgxmock/v4"
)

func TestVehicleHandler_GetVehicles_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	vs := service.NewVehicleService(mock, nil)
	h := NewVehicleHandler(vs)

	uID := uuid.New()
	vID := uuid.New()

	mock.ExpectQuery("SELECT id, user_id, name, make, model, variant, year, is_primary, current_odometer FROM vehicles").
		WithArgs(uID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "make", "model", "variant", "year", "is_primary", "current_odometer"}).
			AddRow(vID, uID, "Civic", "Honda", "Civic", "RS", 2022, true, 10000))

	req := httptest.NewRequest(http.MethodGet, "/api/vehicles", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestVehicleHandler_AddVehicle_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	vs := service.NewVehicleService(mock, nil)
	h := NewVehicleHandler(vs)

	uID := uuid.New()
	vID := uuid.New()

	v := service.Vehicle{
		Name:            "Civic",
		Make:            "Honda",
		Model:           "Civic",
		Year:            2022,
		CurrentOdometer: 10000,
	}
	body, _ := json.Marshal(v)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT").
		WithArgs(uID).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("INSERT INTO vehicles").
		WithArgs(uID, v.Name, v.Make, v.Model, v.Variant, v.Year, true, v.CurrentOdometer).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(vID))

	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodPost, "/api/vehicles", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestVehicleHandler_UpdateVehicle_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	vs := service.NewVehicleService(mock, nil)
	h := NewVehicleHandler(vs)

	uID := uuid.New()
	vID := uuid.New()

	v := service.Vehicle{
		Name:  "Updated Civic",
		Make:  "Honda",
		Model: "Civic",
		Year:  2022,
	}
	body, _ := json.Marshal(v)

	mock.ExpectQuery("SELECT id, user_id, name").
		WithArgs(vID, uID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "make", "model", "variant", "year", "is_primary", "current_odometer"}).
			AddRow(vID, uID, "Civic", "Honda", "Civic", "RS", 2022, true, 10000))

	mock.ExpectExec("UPDATE vehicles").
		WithArgs(v.Name, v.Make, v.Model, v.Variant, v.Year, true, vID, uID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	req := httptest.NewRequest(http.MethodPut, "/api/vehicles/"+vID.String(), bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestVehicleHandler_DeleteVehicle_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	vs := service.NewVehicleService(mock, nil)
	h := NewVehicleHandler(vs)

	uID := uuid.New()
	vID := uuid.New()

	mock.ExpectExec("DELETE FROM vehicles").
		WithArgs(vID, uID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	req := httptest.NewRequest(http.MethodDelete, "/api/vehicles/"+vID.String(), nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestVehicleHandler_SetPrimaryVehicle_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	vs := service.NewVehicleService(mock, nil)
	h := NewVehicleHandler(vs)

	uID := uuid.New()
	vID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE vehicles SET is_primary = false").
		WithArgs(uID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectExec("UPDATE vehicles SET is_primary = true").
		WithArgs(vID, uID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodPut, "/api/vehicles/"+vID.String()+"/primary", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
