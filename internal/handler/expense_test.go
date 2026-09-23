package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/middleware"
	"github.com/ilhamasy/gastrack-ms/internal/service"
	"github.com/pashagolub/pgxmock/v4"
)

func TestExpenseHandler_GetExpenseAnalytics(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	expSvc := service.NewExpenseService(mock)
	vSvc := service.NewVehicleService(mock, nil)
	h := NewExpenseHandler(expSvc, vSvc)

	vID := uuid.New()
	uID := uuid.New()

	// Ownership check
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(vID, uID).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	// Analytics queries
	mock.ExpectQuery("SELECT TO_CHAR").
		WithArgs(vID.String()).
		WillReturnRows(pgxmock.NewRows([]string{"month", "total"}).AddRow("2026-09", 100000.0))

	mock.ExpectQuery("SELECT i.item_name").
		WithArgs(vID.String()).
		WillReturnRows(pgxmock.NewRows([]string{"category", "total"}).AddRow("Oil", 100000.0))

	req := httptest.NewRequest(http.MethodGet, "/api/vehicles/"+vID.String()+"/analytics/expenses", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uID.String())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.GetExpenseAnalytics(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
