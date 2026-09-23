package service

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
)

func TestNotificationService_CheckAndSendMaintenanceReminders(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	prefSvc := NewPreferencesService(mock)
	svc := NewNotificationService(mock, prefSvc)

	userID := "user-123"
	vehicleID := "v-456"

	// Mock preferences lookup
	mock.ExpectQuery("SELECT email_enabled, push_enabled FROM notification_preferences").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"email_enabled", "push_enabled"}).AddRow(true, true))

	err = svc.CheckAndSendMaintenanceReminders(context.Background(), userID, vehicleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
