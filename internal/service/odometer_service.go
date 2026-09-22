package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ilhamasy/gastrack-ms/internal/model"
)

var (
	ErrInvalidOdometer = errors.New("new odometer value cannot be lower than current")
)

type OdometerService struct {
	db *pgxpool.Pool
}

func NewOdometerService(db *pgxpool.Pool) *OdometerService {
	return &OdometerService{db: db}
}

func (s *OdometerService) LogOdometer(ctx context.Context, vehicleID uuid.UUID, userID uuid.UUID, value int) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Verify vehicle ownership and get current odometer
	var currentOdometer int
	err = tx.QueryRow(ctx, "SELECT current_odometer FROM vehicles WHERE id = $1 AND user_id = $2 FOR UPDATE", vehicleID, userID).Scan(&currentOdometer)
	if err != nil {
		// Could be pgx.ErrNoRows or something else, return vehicle not found
		return ErrVehicleNotFound
	}

	if value < currentOdometer {
		return ErrInvalidOdometer
	}

	// Insert into odometer_logs
	_, err = tx.Exec(ctx, "INSERT INTO odometer_logs (vehicle_id, odometer_value) VALUES ($1, $2)", vehicleID, value)
	if err != nil {
		return fmt.Errorf("failed to insert odometer log: %w", err)
	}

	// Update current_odometer in vehicles table
	_, err = tx.Exec(ctx, "UPDATE vehicles SET current_odometer = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", value, vehicleID)
	if err != nil {
		return fmt.Errorf("failed to update vehicle current_odometer: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *OdometerService) GetOdometerHistory(ctx context.Context, vehicleID uuid.UUID, userID uuid.UUID) ([]model.OdometerLog, error) {
	// First check if the vehicle belongs to the user
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM vehicles WHERE id = $1 AND user_id = $2)", vehicleID, userID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ownership: %w", err)
	}
	if !exists {
		return nil, ErrVehicleNotFound
	}

	rows, err := s.db.Query(ctx, "SELECT id, vehicle_id, odometer_value, recorded_at FROM odometer_logs WHERE vehicle_id = $1 ORDER BY recorded_at DESC", vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query odometer history: %w", err)
	}
	defer rows.Close()

	var history []model.OdometerLog
	for rows.Next() {
		var log model.OdometerLog
		if err := rows.Scan(&log.ID, &log.VehicleID, &log.OdometerValue, &log.RecordedAt); err != nil {
			return nil, fmt.Errorf("failed to scan odometer log: %w", err)
		}
		history = append(history, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	if history == nil {
		history = []model.OdometerLog{}
	}

	return history, nil
}
