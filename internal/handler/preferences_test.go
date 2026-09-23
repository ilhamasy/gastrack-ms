package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilhamasy/gastrack-ms/internal/middleware"
	"github.com/ilhamasy/gastrack-ms/internal/model"
	"github.com/ilhamasy/gastrack-ms/internal/service"
	"github.com/pashagolub/pgxmock/v4"
)

func TestPreferencesHandler_GetPreferences(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := service.NewPreferencesService(mock)
	h := NewPreferencesHandler(svc)

	userID := "user-123"
	mock.ExpectQuery("SELECT email_enabled, push_enabled").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"email_enabled", "push_enabled"}).AddRow(true, true))

	req := httptest.NewRequest(http.MethodGet, "/api/users/preferences", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestPreferencesHandler_PutPreferences(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := service.NewPreferencesService(mock)
	h := NewPreferencesHandler(svc)

	userID := "user-123"
	prefs := model.NotificationPreferences{EmailEnabled: false, PushEnabled: true}
	body, _ := json.Marshal(prefs)

	mock.ExpectExec("INSERT INTO notification_preferences").
		WithArgs(userID, false, true).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	req := httptest.NewRequest(http.MethodPut, "/api/users/preferences", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
