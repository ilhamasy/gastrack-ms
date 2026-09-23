package service

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
)

func TestExpenseService_GetExpenseAnalytics(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewExpenseService(mock)
	vehicleID := "v-123"

	// Monthly query
	mock.ExpectQuery("SELECT TO_CHAR").
		WithArgs(vehicleID).
		WillReturnRows(pgxmock.NewRows([]string{"month", "total"}).
			AddRow("2026-08", 150000.0).
			AddRow("2026-09", 250000.0))

	// Category query
	mock.ExpectQuery("SELECT i.item_name").
		WithArgs(vehicleID).
		WillReturnRows(pgxmock.NewRows([]string{"category", "total"}).
			AddRow("Oil Change", 100000.0).
			AddRow("Tire Change", 300000.0))

	analytics, err := svc.GetExpenseAnalytics(context.Background(), vehicleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(analytics.MonthlyExpenses) != 2 {
		t.Fatalf("expected 2 monthly expenses, got %d", len(analytics.MonthlyExpenses))
	}
	if len(analytics.CategoryExpenses) != 2 {
		t.Fatalf("expected 2 category expenses, got %d", len(analytics.CategoryExpenses))
	}
}
