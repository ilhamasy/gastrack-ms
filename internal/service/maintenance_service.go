package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MaintenanceService struct {
	db *pgxpool.Pool
}

func NewMaintenanceService(db *pgxpool.Pool) *MaintenanceService {
	return &MaintenanceService{db: db}
}

// ApplyTemplatesToVehicle finds matching templates and inserts them into vehicle_maintenance.
// It uses an existing transaction since it is called during vehicle creation.
func (s *MaintenanceService) ApplyTemplatesToVehicle(ctx context.Context, tx pgx.Tx, vehicle *Vehicle) error {
	// Query matching templates
	query := `
		SELECT id, name, description, interval_km, interval_months 
		FROM maintenance_templates 
		WHERE make ILIKE $1 AND model ILIKE $2
	`
	rows, err := tx.Query(ctx, query, vehicle.Make, vehicle.Model)
	if err != nil {
		return fmt.Errorf("failed to query maintenance templates: %w", err)
	}
	defer rows.Close()

	type template struct {
		id             string
		name           string
		description    *string
		intervalKm     *int
		intervalMonths *int
	}

	var templates []template
	for rows.Next() {
		var t template
		if err := rows.Scan(&t.id, &t.name, &t.description, &t.intervalKm, &t.intervalMonths); err != nil {
			return fmt.Errorf("failed to scan maintenance template: %w", err)
		}
		templates = append(templates, t)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows iteration error: %w", err)
	}

	// Insert into vehicle_maintenance
	if len(templates) > 0 {
		insertQuery := `
			INSERT INTO vehicle_maintenance 
			(vehicle_id, template_id, name, description, interval_km, interval_months, last_service_km) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		for _, t := range templates {
			// Initially last_service_km could be the current odometer or 0
			// Depending on rules, if vehicle is added, its current odometer is baseline.
			_, err := tx.Exec(ctx, insertQuery,
				vehicle.ID, t.id, t.name, t.description, t.intervalKm, t.intervalMonths, vehicle.CurrentOdometer,
			)
			if err != nil {
				return fmt.Errorf("failed to insert vehicle maintenance item: %w", err)
			}
		}
	}

	return nil
}
