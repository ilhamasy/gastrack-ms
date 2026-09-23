package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
)

func TestVehicleService_GetVehicles_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewVehicleService(mock, nil)
	userID := uuid.New()
	vehicleID := uuid.New()

	mock.ExpectQuery("SELECT id, user_id, name, make, model, variant, year, is_primary, current_odometer FROM vehicles").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "make", "model", "variant", "year", "is_primary", "current_odometer"}).
			AddRow(vehicleID, userID, "My Car", "Honda", "Civic", "RS", 2020, true, 50000))

	vehicles, err := svc.GetVehicles(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vehicles) != 1 {
		t.Fatalf("expected 1 vehicle, got %d", len(vehicles))
	}
	if vehicles[0].Name != "My Car" {
		t.Errorf("expected 'My Car', got '%s'", vehicles[0].Name)
	}
}

func TestVehicleService_GetVehicles_Empty(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewVehicleService(mock, nil)
	userID := uuid.New()

	mock.ExpectQuery("SELECT").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "make", "model", "variant", "year", "is_primary", "current_odometer"}))

	vehicles, err := svc.GetVehicles(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vehicles == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(vehicles) != 0 {
		t.Errorf("expected 0 vehicles, got %d", len(vehicles))
	}
}

func TestVehicleService_IsVehicleOwner_True(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewVehicleService(mock, nil)
	vehicleID := uuid.New()
	userID := uuid.New()

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(vehicleID, userID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	isOwner, err := svc.IsVehicleOwner(context.Background(), vehicleID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isOwner {
		t.Error("expected true, got false")
	}
}

func TestVehicleService_IsVehicleOwner_False(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewVehicleService(mock, nil)
	vehicleID := uuid.New()
	userID := uuid.New()

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(vehicleID, userID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))

	isOwner, err := svc.IsVehicleOwner(context.Background(), vehicleID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isOwner {
		t.Error("expected false, got true")
	}
}

func TestVehicleService_GetVehicleByID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewVehicleService(mock, nil)
	vehicleID := uuid.New()
	userID := uuid.New()

	mock.ExpectQuery("SELECT id, user_id, name").
		WithArgs(vehicleID, userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "make", "model", "variant", "year", "is_primary", "current_odometer"}).
			AddRow(vehicleID, userID, "My Car", "Honda", "Civic", "RS", 2020, true, 50000))

	v, err := svc.GetVehicleByID(context.Background(), vehicleID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Name != "My Car" {
		t.Errorf("expected 'My Car', got '%s'", v.Name)
	}
}

func TestVehicleService_GetVehicleByID_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewVehicleService(mock, nil)
	vehicleID := uuid.New()
	userID := uuid.New()

	mock.ExpectQuery("SELECT id, user_id, name").
		WithArgs(vehicleID, userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "make", "model", "variant", "year", "is_primary", "current_odometer"}))

	_, err = svc.GetVehicleByID(context.Background(), vehicleID, userID)
	if err != ErrVehicleNotFound {
		t.Errorf("expected ErrVehicleNotFound, got %v", err)
	}
}

func TestVehicleService_DeleteVehicle_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewVehicleService(mock, nil)
	vehicleID := uuid.New()
	userID := uuid.New()

	mock.ExpectExec("DELETE FROM vehicles").
		WithArgs(vehicleID, userID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = svc.DeleteVehicle(context.Background(), vehicleID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVehicleService_DeleteVehicle_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewVehicleService(mock, nil)
	vehicleID := uuid.New()
	userID := uuid.New()

	mock.ExpectExec("DELETE FROM vehicles").
		WithArgs(vehicleID, userID).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err = svc.DeleteVehicle(context.Background(), vehicleID, userID)
	if err != ErrVehicleNotFound {
		t.Errorf("expected ErrVehicleNotFound, got %v", err)
	}
}

func TestVehicleService_UpdateVehicle_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewVehicleService(mock, nil)
	v := &Vehicle{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Name:   "Updated Car",
		Make:   "Toyota",
		Model:  "Corolla",
		Year:   2021,
	}

	mock.ExpectExec("UPDATE vehicles").
		WithArgs(v.Name, v.Make, v.Model, v.Variant, v.Year, v.IsPrimary, v.ID, v.UserID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = svc.UpdateVehicle(context.Background(), v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVehicleService_UpdateVehicle_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewVehicleService(mock, nil)
	v := &Vehicle{ID: uuid.New(), UserID: uuid.New()}

	mock.ExpectExec("UPDATE vehicles").
		WithArgs(v.Name, v.Make, v.Model, v.Variant, v.Year, v.IsPrimary, v.ID, v.UserID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err = svc.UpdateVehicle(context.Background(), v)
	if err != ErrVehicleNotFound {
		t.Errorf("expected ErrVehicleNotFound, got %v", err)
	}
}

func TestVehicleService_AddVehicle(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewVehicleService(mock, nil)
	v := &Vehicle{
		UserID:          uuid.New(),
		Name:            "Civic",
		Make:            "Honda",
		Model:           "Civic",
		Year:            2022,
		CurrentOdometer: 10000,
	}
	vID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT").
		WithArgs(v.UserID).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("INSERT INTO vehicles").
		WithArgs(v.UserID, v.Name, v.Make, v.Model, v.Variant, v.Year, true, v.CurrentOdometer).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(vID))

	mock.ExpectCommit()

	err = svc.AddVehicle(context.Background(), v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != vID {
		t.Errorf("expected vehicle ID %v, got %v", vID, v.ID)
	}
}

func TestVehicleService_SetPrimaryVehicle(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewVehicleService(mock, nil)
	vID := uuid.New()
	uID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE vehicles SET is_primary = false").
		WithArgs(uID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectExec("UPDATE vehicles SET is_primary = true").
		WithArgs(vID, uID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectCommit()

	err = svc.SetPrimaryVehicle(context.Background(), vID, uID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
