package service

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/ilhamasy/gastrack-ms/internal/model"
	"github.com/pashagolub/pgxmock/v4"
)

func TestPreferencesService_GetPreferences_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewPreferencesService(mock)
	userID := "user-123"

	mock.ExpectQuery("SELECT email_enabled, push_enabled FROM notification_preferences").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"email_enabled", "push_enabled"}).AddRow(true, false))

	prefs, err := svc.GetPreferences(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !prefs.EmailEnabled || prefs.PushEnabled {
		t.Errorf("unexpected preferences state: %+v", prefs)
	}
}

func TestPreferencesService_GetPreferences_Default(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewPreferencesService(mock)
	userID := "user-123"

	mock.ExpectQuery("SELECT email_enabled, push_enabled FROM notification_preferences").
		WithArgs(userID).
		WillReturnError(pgx.ErrNoRows)

	mock.ExpectExec("INSERT INTO notification_preferences").
		WithArgs(userID).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	prefs, err := svc.GetPreferences(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !prefs.EmailEnabled || !prefs.PushEnabled {
		t.Errorf("expected default true/true, got %+v", prefs)
	}
}

func TestPreferencesService_UpdatePreferences(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewPreferencesService(mock)
	userID := "user-123"
	prefs := &model.NotificationPreferences{
		EmailEnabled: false,
		PushEnabled:  true,
	}

	mock.ExpectExec("INSERT INTO notification_preferences").
		WithArgs(userID, prefs.EmailEnabled, prefs.PushEnabled).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = svc.UpdatePreferences(context.Background(), userID, prefs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
