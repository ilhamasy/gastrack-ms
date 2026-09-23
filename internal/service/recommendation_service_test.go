package service

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
)

func TestRecommendationService_GetRecommendations(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewRecommendationService(mock)
	vehicleID := "v-123"

	// 1. Maintenance query
	lastDate := time.Now().AddDate(0, -10, 0)
	lastKm := 5000
	intervalKm := 10000
	intervalMonths := 6
	mock.ExpectQuery("SELECT id, name, last_service_km, interval_km").
		WithArgs(vehicleID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "last_service_km", "interval_km", "last_service_date", "interval_months"}).
			AddRow("m-1", "Spark Plug", &lastKm, &intervalKm, &lastDate, &intervalMonths))

	// Current odometer query
	now := time.Now()
	mock.ExpectQuery("SELECT current_odometer, updated_at FROM vehicles").
		WithArgs(vehicleID).
		WillReturnRows(pgxmock.NewRows([]string{"current_odometer", "updated_at"}).AddRow(20000, now))

	// 2. Service frequency query
	mock.ExpectQuery("SELECT COUNT").
		WithArgs(vehicleID).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))

	recs, err := svc.GetRecommendations(context.Background(), vehicleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) == 0 {
		t.Error("expected at least 1 recommendation")
	}
}
