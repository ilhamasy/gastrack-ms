package model

import (
	"time"

	"github.com/google/uuid"
)

type ServiceRecord struct {
	ID           uuid.UUID     `json:"id"`
	VehicleID    uuid.UUID     `json:"vehicle_id"`
	ServiceDate  time.Time     `json:"service_date"`
	OdometerKm   int           `json:"odometer_km"`
	WorkshopName *string       `json:"workshop_name"`
	TotalCost    float64       `json:"total_cost"`
	Notes        *string       `json:"notes"`
	Items        []ServiceItem `json:"items,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type ServiceItem struct {
	ID              uuid.UUID  `json:"id"`
	ServiceRecordID uuid.UUID  `json:"service_record_id"`
	MaintenanceID   *uuid.UUID `json:"maintenance_id"`
	ItemName        string     `json:"item_name"`
	Brand           *string    `json:"brand"`
	Product         *string    `json:"product"`
	PartNumber      *string    `json:"part_number"`
	Quantity        float64    `json:"quantity"`
	Cost            float64    `json:"cost"`
	Notes           *string    `json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
}
