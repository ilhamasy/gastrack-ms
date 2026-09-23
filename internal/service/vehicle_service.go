package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrVehicleNotFound = errors.New("vehicle not found")
)

type Vehicle struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	Name            string    `json:"name"`
	Make            string    `json:"make"`
	Model           string    `json:"model"`
	Variant         string    `json:"variant"`
	Year            int       `json:"year"`
	IsPrimary       bool      `json:"is_primary"`
	CurrentOdometer int       `json:"current_odometer"`
}

type VehicleService struct {
	db                 *pgxpool.Pool
	maintenanceService *MaintenanceService
}

func NewVehicleService(db *pgxpool.Pool, maintenanceService *MaintenanceService) *VehicleService {
	return &VehicleService{db: db, maintenanceService: maintenanceService}
}

func (s *VehicleService) GetVehicles(ctx context.Context, userID uuid.UUID) ([]Vehicle, error) {
	rows, err := s.db.Query(ctx, 
		"SELECT id, user_id, name, make, model, variant, year, is_primary, current_odometer FROM vehicles WHERE user_id = $1 ORDER BY created_at DESC", 
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicles: %w", err)
	}
	defer rows.Close()

	var vehicles []Vehicle
	for rows.Next() {
		var v Vehicle
		if err := rows.Scan(&v.ID, &v.UserID, &v.Name, &v.Make, &v.Model, &v.Variant, &v.Year, &v.IsPrimary, &v.CurrentOdometer); err != nil {
			return nil, fmt.Errorf("failed to scan vehicle: %w", err)
		}
		vehicles = append(vehicles, v)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if vehicles == nil {
		vehicles = []Vehicle{} // Ensure we don't return null in JSON
	}

	return vehicles, nil
}

func (s *VehicleService) IsVehicleOwner(ctx context.Context, vehicleID uuid.UUID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM vehicles WHERE id = $1 AND user_id = $2)", vehicleID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check vehicle ownership: %w", err)
	}
	return exists, nil
}

func (s *VehicleService) AddVehicle(ctx context.Context, v *Vehicle) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if this is the user's first vehicle
	var count int
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM vehicles WHERE user_id = $1", v.UserID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count vehicles: %w", err)
	}
	
	if count == 0 {
		v.IsPrimary = true
	}

	err = tx.QueryRow(ctx, 
		`INSERT INTO vehicles (user_id, name, make, model, variant, year, is_primary, current_odometer) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		v.UserID, v.Name, v.Make, v.Model, v.Variant, v.Year, v.IsPrimary, v.CurrentOdometer,
	).Scan(&v.ID)

	if err != nil {
		return fmt.Errorf("failed to insert vehicle: %w", err)
	}

	if s.maintenanceService != nil {
		if err := s.maintenanceService.ApplyTemplatesToVehicle(ctx, tx, v); err != nil {
			return fmt.Errorf("failed to apply maintenance templates: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *VehicleService) UpdateVehicle(ctx context.Context, v *Vehicle) error {
	cmdTag, err := s.db.Exec(ctx, 
		`UPDATE vehicles 
		 SET name = $1, make = $2, model = $3, variant = $4, year = $5, is_primary = $6, updated_at = CURRENT_TIMESTAMP
		 WHERE id = $7 AND user_id = $8`,
		v.Name, v.Make, v.Model, v.Variant, v.Year, v.IsPrimary, v.ID, v.UserID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update vehicle: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrVehicleNotFound
	}

	return nil
}

func (s *VehicleService) DeleteVehicle(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Let's check if it was primary. If yes, and there are other vehicles, we might need to assign a new primary. 
	// For now, MVP just deletes it.

	cmdTag, err := s.db.Exec(ctx, "DELETE FROM vehicles WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete vehicle: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrVehicleNotFound
	}

	return nil
}

func (s *VehicleService) SetPrimaryVehicle(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Unset primary for all vehicles of this user
	_, err = tx.Exec(ctx, "UPDATE vehicles SET is_primary = false WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to unset primary vehicles: %w", err)
	}

	// Set the selected vehicle as primary
	cmdTag, err := tx.Exec(ctx, "UPDATE vehicles SET is_primary = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return fmt.Errorf("failed to set primary vehicle: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrVehicleNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
