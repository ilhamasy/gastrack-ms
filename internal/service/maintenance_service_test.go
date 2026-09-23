package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/model"
	"github.com/pashagolub/pgxmock/v4"
)

func TestMaintenanceService_ApplyTemplatesToVehicle(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewMaintenanceService(mock)
	v := &Vehicle{
		ID:              uuid.New(),
		Make:            "Honda",
		Model:           "Civic",
		UserID:          uuid.New(),
		CurrentOdometer: 5000,
	}

	mock.ExpectBegin()
	tx, err := mock.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	desc := "Engine Oil Change"
	intervalKm := 10000
	intervalMonths := 6

	mock.ExpectQuery("SELECT id, name, description, interval_km, interval_months FROM maintenance_templates").
		WithArgs(v.Make, v.Model).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "description", "interval_km", "interval_months"}).
			AddRow("tpl-1", "Oil Change", &desc, &intervalKm, &intervalMonths))

	mock.ExpectExec("INSERT INTO vehicle_maintenance").
		WithArgs(v.ID, "tpl-1", "Oil Change", &desc, &intervalKm, &intervalMonths, v.CurrentOdometer).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = svc.ApplyTemplatesToVehicle(context.Background(), tx, v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMaintenanceService_GetVehicleMaintenance(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewMaintenanceService(mock)
	vID := uuid.New().String()
	mID := uuid.New()
	now := time.Now()

	// Query vehicle current odometer
	mock.ExpectQuery("SELECT current_odometer FROM vehicles").
		WithArgs(vID).
		WillReturnRows(pgxmock.NewRows([]string{"current_odometer"}).AddRow(15000))

	lastDate := time.Now().AddDate(0, -5, 0)
	lastKm := 5000
	intervalKm := 10000
	intervalMonths := 6
	desc := "Check brakes"

	mock.ExpectQuery("SELECT id, vehicle_id, template_id, name, description").
		WithArgs(vID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "vehicle_id", "template_id", "name", "description", "interval_km", "interval_months", "last_service_km", "last_service_date", "source", "created_at", "updated_at"}).
			AddRow(mID, vID, nil, "Brake Check", &desc, &intervalKm, &intervalMonths, &lastKm, &lastDate, "USER_CREATED", now, now))

	items, err := svc.GetVehicleMaintenance(context.Background(), vID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 maintenance item, got %d", len(items))
	}
	if items[0].Name != "Brake Check" {
		t.Errorf("expected 'Brake Check', got '%s'", items[0].Name)
	}
}

func TestMaintenanceService_AddCustomMaintenanceItem(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewMaintenanceService(mock)
	mID := uuid.New()
	vID := uuid.New()
	desc := "Coolant flush"
	intervalKm := 20000
	intervalMonths := 12
	lastKm := 5000
	lastDate := time.Now()
	now := time.Now()

	item := &model.VehicleMaintenance{
		VehicleID:       vID,
		Name:            "Coolant Flush",
		Description:     &desc,
		IntervalKm:      &intervalKm,
		IntervalMonths:  &intervalMonths,
		LastServiceKm:   &lastKm,
		LastServiceDate: &lastDate,
	}

	mock.ExpectQuery("INSERT INTO vehicle_maintenance").
		WithArgs(item.VehicleID, item.Name, item.Description, item.IntervalKm, item.IntervalMonths, item.LastServiceKm, item.LastServiceDate).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at", "source"}).
			AddRow(mID, now, now, "USER_CREATED"))

	err = svc.AddCustomMaintenanceItem(context.Background(), item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != mID {
		t.Errorf("expected item ID %v, got %v", mID, item.ID)
	}
}

func TestMaintenanceService_UpdateMaintenanceItem(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewMaintenanceService(mock)
	mID := uuid.New()
	desc := "Updated desc"
	intervalKm := 15000
	intervalMonths := 9
	lastKm := 8000
	lastDate := time.Now()

	item := &model.VehicleMaintenance{
		ID:              mID,
		Name:            "Updated Oil Change",
		Description:     &desc,
		IntervalKm:      &intervalKm,
		IntervalMonths:  &intervalMonths,
		LastServiceKm:   &lastKm,
		LastServiceDate: &lastDate,
	}

	vID := uuid.New().String()
	mIDStr := mID.String()

	mock.ExpectQuery("SELECT source FROM vehicle_maintenance").
		WithArgs(mIDStr, vID).
		WillReturnRows(pgxmock.NewRows([]string{"source"}).AddRow("TEMPLATE"))

	mock.ExpectExec("UPDATE vehicle_maintenance SET").
		WithArgs(item.Name, item.Description, item.IntervalKm, item.IntervalMonths, item.LastServiceKm, item.LastServiceDate, "USER_CUSTOMIZED", mIDStr, vID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = svc.UpdateMaintenanceItem(context.Background(), vID, mIDStr, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
