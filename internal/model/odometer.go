package model

import (
	"time"

	"github.com/google/uuid"
)

type OdometerLog struct {
	ID           uuid.UUID `json:"id"`
	VehicleID    uuid.UUID `json:"vehicle_id"`
	OdometerValue int      `json:"odometer_value"`
	RecordedAt   time.Time `json:"recorded_at"`
}
