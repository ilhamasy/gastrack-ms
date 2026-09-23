package model

import (
	"time"

	"github.com/google/uuid"
)

type MaintenanceTemplate struct {
	ID             uuid.UUID  `json:"id"`
	Make           string     `json:"make"`
	Model          string     `json:"model"`
	Variant        *string    `json:"variant"`
	YearStart      *int       `json:"year_start"`
	YearEnd        *int       `json:"year_end"`
	Name           string     `json:"name"`
	Description    *string    `json:"description"`
	IntervalKm     *int       `json:"interval_km"`
	IntervalMonths *int       `json:"interval_months"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type VehicleMaintenance struct {
	ID              uuid.UUID  `json:"id"`
	VehicleID       uuid.UUID  `json:"vehicle_id"`
	TemplateID      *uuid.UUID `json:"template_id"`
	Name            string     `json:"name"`
	Description     *string    `json:"description"`
	IntervalKm      *int       `json:"interval_km"`
	IntervalMonths  *int       `json:"interval_months"`
	LastServiceKm   *int       `json:"last_service_km"`
	LastServiceDate *time.Time `json:"last_service_date"`
	Source          string     `json:"source"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Calculated status fields (not persisted)
	Status          string     `json:"status,omitempty"`
	Priority        int        `json:"priority,omitempty"`
	RemainingKm     *int       `json:"remaining_km,omitempty"`
	NextServiceKm   *int       `json:"next_service_km,omitempty"`
	RemainingDays   *int       `json:"remaining_days,omitempty"`
	NextServiceDate *time.Time `json:"next_service_date,omitempty"`
}
