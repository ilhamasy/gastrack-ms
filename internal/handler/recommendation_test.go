package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/middleware"
	"github.com/ilhamasy/gastrack-ms/internal/service"
	"github.com/pashagolub/pgxmock/v4"
)

func TestRecommendationHandler_GetRecommendations(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	recSvc := service.NewRecommendationService(mock)
	vSvc := service.NewVehicleService(mock, nil)
	h := NewRecommendationHandler(recSvc, vSvc)

	vID := uuid.New()
	uID := uuid.New()
	vIDStr := vID.String()

	// Ownership check
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(vID, uID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	// Recommendation queries
	lastDate := time.Now().AddDate(0, -10, 0)
	lastKm := 5000
	intervalKm := 10000
	intervalMonths := 6
	mock.ExpectQuery("SELECT id, name, last_service_km, interval_km").
		WithArgs(vIDStr).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "last_service_km", "interval_km", "last_service_date", "interval_months"}).
			AddRow("m-1", "Spark Plug", &lastKm, &intervalKm, &lastDate, &intervalMonths))

	now := time.Now()
	mock.ExpectQuery("SELECT current_odometer, updated_at FROM vehicles").
		WithArgs(vIDStr).
		WillReturnRows(pgxmock.NewRows([]string{"current_odometer", "updated_at"}).AddRow(20000, now))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs(vIDStr).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))

	req := httptest.NewRequest(http.MethodGet, "/api/vehicles/"+vIDStr+"/recommendations", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.GetRecommendations(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
