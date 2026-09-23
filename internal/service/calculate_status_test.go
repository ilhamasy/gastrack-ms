package service

import (
	"testing"
	"time"

	"github.com/ilhamasy/gastrack-ms/internal/model"
)

func TestCalculateMaintenanceStatus_KM(t *testing.T) {
	intervalKm := 2000
	lastServiceKm := 10000

	tests := []struct {
		name            string
		currentOdometer int
		expectedStatus  string
		expectedRemKm   int
	}{
		{
			name:            "Normal",
			currentOdometer: 11000,
			expectedStatus:  "NORMAL",
			expectedRemKm:   1000, // 12000 - 11000
		},
		{
			name:            "Upcoming",
			currentOdometer: 11600,
			expectedStatus:  "UPCOMING",
			expectedRemKm:   400, // 12000 - 11600
		},
		{
			name:            "Critical",
			currentOdometer: 11950, // 50 KM remaining
			expectedStatus:  "CRITICAL",
			expectedRemKm:   50, // 12000 - 11950
		},
		{
			name:            "Due",
			currentOdometer: 12000,
			expectedStatus:  "DUE",
			expectedRemKm:   0,
		},
		{
			name:            "Overdue",
			currentOdometer: 12100,
			expectedStatus:  "OVERDUE",
			expectedRemKm:   -100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &model.VehicleMaintenance{
				IntervalKm:    &intervalKm,
				LastServiceKm: &lastServiceKm,
			}

			CalculateMaintenanceStatus(item, tt.currentOdometer)

			if item.Status != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, item.Status)
			}
			if item.RemainingKm == nil || *item.RemainingKm != tt.expectedRemKm {
				t.Errorf("expected remaining km %d, got %v", tt.expectedRemKm, item.RemainingKm)
			}
		})
	}
}

func TestCalculateMaintenanceStatus_Time(t *testing.T) {
	intervalMonths := 6
	lastServiceDate := time.Now().AddDate(0, -6, 7) // Last service 6 months minus 7 days ago -> due in 7 days

	item := &model.VehicleMaintenance{
		IntervalMonths:  &intervalMonths,
		LastServiceDate: &lastServiceDate,
	}

	CalculateMaintenanceStatus(item, 10000)

	if item.Status != "CRITICAL" {
		t.Errorf("expected status CRITICAL, got %q", item.Status)
	}
	if item.RemainingDays == nil || *item.RemainingDays < 6 || *item.RemainingDays > 7 {
		var got int
		if item.RemainingDays != nil {
			got = *item.RemainingDays
		}
		t.Errorf("expected remaining days around 7, got %d", got)
	}
}
