package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/model"
	"github.com/pashagolub/pgxmock/v4"
)

func TestServiceRecordService_AddServiceRecord(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewServiceRecordService(mock)
	vehicleID := uuid.New().String()
	notes := "Regular maintenance"
	now := time.Now()

	record := model.ServiceRecord{
		ServiceDate: time.Now(),
		OdometerKm:  15000,
		Notes:       &notes,
		TotalCost:   150000,
		Items: []model.ServiceItem{
			{ItemName: "Engine Oil", Cost: 150000, Quantity: 1},
		},
	}

	mock.ExpectBegin()

	// Insert record (7 args: recordID, vehicleID, record.ServiceDate, record.OdometerKm, record.WorkshopName, record.TotalCost, record.Notes)
	mock.ExpectQuery("INSERT INTO service_records").
		WithArgs(pgxmock.AnyArg(), vehicleID, record.ServiceDate, record.OdometerKm, record.WorkshopName, record.TotalCost, record.Notes).
		WillReturnRows(pgxmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))

	// Insert item (10 args)
	mock.ExpectQuery("INSERT INTO service_items").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), "Engine Oil", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), 1.0, 150000.0, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"created_at"}).AddRow(now))

	// Update vehicle odometer
	mock.ExpectExec("UPDATE vehicles").
		WithArgs(record.OdometerKm, vehicleID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectCommit()

	result, err := svc.AddServiceRecord(context.Background(), vehicleID, record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCost != 150000 {
		t.Errorf("expected TotalCost 150000, got %v", result.TotalCost)
	}
}

func TestServiceRecordService_GetServiceRecords(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewServiceRecordService(mock)
	vehicleID := uuid.New().String()
	recID := uuid.New()
	itemID := uuid.New()
	now := time.Now()
	notes := "Regular maintenance"

	// Select records
	mock.ExpectQuery("SELECT r.id, r.vehicle_id").
		WithArgs(vehicleID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "vehicle_id", "service_date", "odometer_km", "workshop_name", "total_cost", "notes", "created_at", "updated_at",
			"item_id", "maintenance_id", "item_name", "brand", "product", "part_number", "quantity", "cost", "item_notes", "item_created_at",
		}).AddRow(
			recID, uuid.MustParse(vehicleID), now, 15000, nil, 150000.0, &notes, now, now,
			&itemID, nil, "Engine Oil", nil, nil, nil, 1.0, 150000.0, nil, now,
		))

	records, err := svc.GetServiceRecords(context.Background(), vehicleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if len(records[0].Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(records[0].Items))
	}
}
