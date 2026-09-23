package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
)

func TestOdometerService_LogOdometer_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewOdometerService(mock)
	vID := uuid.New()
	uID := uuid.New()

	mock.ExpectBegin()

	// Check current odometer
	mock.ExpectQuery("SELECT current_odometer FROM vehicles").
		WithArgs(vID, uID).
		WillReturnRows(pgxmock.NewRows([]string{"current_odometer"}).AddRow(10000))

	// Insert odometer log
	mock.ExpectExec("INSERT INTO odometer_logs").
		WithArgs(vID, 12000).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Update vehicle odometer
	mock.ExpectExec("UPDATE vehicles SET current_odometer").
		WithArgs(12000, vID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectCommit()

	err = svc.LogOdometer(context.Background(), vID, uID, 12000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOdometerService_LogOdometer_InvalidOdometer(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewOdometerService(mock)
	vID := uuid.New()
	uID := uuid.New()

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT current_odometer FROM vehicles").
		WithArgs(vID, uID).
		WillReturnRows(pgxmock.NewRows([]string{"current_odometer"}).AddRow(10000))

	mock.ExpectRollback()

	err = svc.LogOdometer(context.Background(), vID, uID, 8000)
	if err != ErrInvalidOdometer {
		t.Errorf("expected ErrInvalidOdometer, got %v", err)
	}
}

func TestOdometerService_GetOdometerHistory(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewOdometerService(mock)
	vID := uuid.New()
	uID := uuid.New()
	logID := uuid.New()
	now := time.Now()

	// Ownership check
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(vID, uID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	// History query
	mock.ExpectQuery("SELECT id, vehicle_id, odometer_value, recorded_at FROM odometer_logs").
		WithArgs(vID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "vehicle_id", "odometer_value", "recorded_at"}).
			AddRow(logID, vID, 10000, now))

	logs, err := svc.GetOdometerHistory(context.Background(), vID, uID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].OdometerValue != 10000 {
		t.Errorf("expected 10000, got %d", logs[0].OdometerValue)
	}
}
