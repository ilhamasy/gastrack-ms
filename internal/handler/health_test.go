package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilhamasy/gastrack-ms/internal/handler"
)

func TestHealthHandler_Success_NoDB(t *testing.T) {
	// Test the health handler when DB is nil (it should return status ok, database uninitialized)
	h := handler.NewHealthHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusServiceUnavailable {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusServiceUnavailable)
	}

	var resp handler.HealthResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not parse json response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %v", resp.Status)
	}
	if resp.Database != "uninitialized" {
		t.Errorf("expected database uninitialized, got %v", resp.Database)
	}
}
