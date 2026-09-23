package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ilhamasy/gastrack-ms/internal/database"
	"github.com/ilhamasy/gastrack-ms/internal/model"
)

type MaintenanceService struct {
	db database.DBPool
}

func NewMaintenanceService(db database.DBPool) *MaintenanceService {
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
			(vehicle_id, template_id, name, description, interval_km, interval_months, last_service_km, source) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'TEMPLATE')
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

func (s *MaintenanceService) GetVehicleMaintenance(ctx context.Context, vehicleID string) ([]model.VehicleMaintenance, error) {
	// First get current odometer
	var currentOdometer int
	err := s.db.QueryRow(ctx, "SELECT current_odometer FROM vehicles WHERE id = $1", vehicleID).Scan(&currentOdometer)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle current odometer: %w", err)
	}

	query := `
		SELECT id, vehicle_id, template_id, name, description, interval_km, interval_months, 
		       last_service_km, last_service_date, source, created_at, updated_at
		FROM vehicle_maintenance
		WHERE vehicle_id = $1
		ORDER BY created_at ASC
	`
	rows, err := s.db.Query(ctx, query, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query vehicle maintenance: %w", err)
	}
	defer rows.Close()

	var items []model.VehicleMaintenance
	for rows.Next() {
		var item model.VehicleMaintenance
		if err := rows.Scan(
			&item.ID, &item.VehicleID, &item.TemplateID, &item.Name, &item.Description,
			&item.IntervalKm, &item.IntervalMonths, &item.LastServiceKm, &item.LastServiceDate,
			&item.Source, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan vehicle maintenance item: %w", err)
		}
		
		CalculateMaintenanceStatus(&item, currentOdometer)
		
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Sort items by priority: Overdue (1) first, Normal (5) last.
	sort.Slice(items, func(i, j int) bool {
		return items[i].Priority < items[j].Priority
	})

	if items == nil {
		items = []model.VehicleMaintenance{}
	}

	return items, nil
}

func (s *MaintenanceService) AddCustomMaintenanceItem(ctx context.Context, item *model.VehicleMaintenance) error {
	query := `
		INSERT INTO vehicle_maintenance 
		(vehicle_id, name, description, interval_km, interval_months, last_service_km, last_service_date, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'USER_CREATED')
		RETURNING id, created_at, updated_at, source
	`
	err := s.db.QueryRow(ctx, query,
		item.VehicleID, item.Name, item.Description, item.IntervalKm, item.IntervalMonths,
		item.LastServiceKm, item.LastServiceDate,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt, &item.Source)

	if err != nil {
		return fmt.Errorf("failed to add custom maintenance item: %w", err)
	}
	return nil
}

func (s *MaintenanceService) UpdateMaintenanceItem(ctx context.Context, vehicleID string, itemID string, updates *model.VehicleMaintenance) error {
	// First get current item to check source
	var currentSource string
	err := s.db.QueryRow(ctx, "SELECT source FROM vehicle_maintenance WHERE id = $1 AND vehicle_id = $2", itemID, vehicleID).Scan(&currentSource)
	if err != nil {
		return fmt.Errorf("failed to get maintenance item: %w", err)
	}

	newSource := currentSource
	if currentSource == "TEMPLATE" {
		newSource = "USER_CUSTOMIZED"
	}

	query := `
		UPDATE vehicle_maintenance 
		SET name = $1, description = $2, interval_km = $3, interval_months = $4, 
		    last_service_km = $5, last_service_date = $6, source = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8 AND vehicle_id = $9
	`
	_, err = s.db.Exec(ctx, query,
		updates.Name, updates.Description, updates.IntervalKm, updates.IntervalMonths,
		updates.LastServiceKm, updates.LastServiceDate, newSource, itemID, vehicleID,
	)
	if err != nil {
		return fmt.Errorf("failed to update maintenance item: %w", err)
	}
	return nil
}

func CalculateMaintenanceStatus(item *model.VehicleMaintenance, currentOdometer int) {
	priority := 5
	status := "NORMAL"

	// KM Based Calculation
	if item.IntervalKm != nil && *item.IntervalKm > 0 {
		var nextKm int
		if item.LastServiceKm != nil {
			nextKm = *item.LastServiceKm + *item.IntervalKm
		} else {
			// If no last service, interval starts from 0 effectively, but really it means it should have been done at IntervalKm.
			nextKm = *item.IntervalKm
		}
		item.NextServiceKm = &nextKm
		remKm := nextKm - currentOdometer
		item.RemainingKm = &remKm

		if remKm < 0 {
			priority = 1
			status = "OVERDUE"
		} else if remKm == 0 {
			if priority > 2 {
				priority = 2
				status = "DUE"
			}
		} else if remKm <= 100 {
			if priority > 3 {
				priority = 3
				status = "CRITICAL"
			}
		} else if remKm <= 500 {
			if priority > 4 {
				priority = 4
				status = "UPCOMING"
			}
		}
	}

	// Time Based Calculation
	if item.IntervalMonths != nil && *item.IntervalMonths > 0 {
		if item.LastServiceDate != nil {
			nextDate := item.LastServiceDate.AddDate(0, *item.IntervalMonths, 0)
			item.NextServiceDate = &nextDate
			remDays := int(time.Until(nextDate).Hours() / 24)
			item.RemainingDays = &remDays

			if remDays < 0 {
				priority = 1
				status = "OVERDUE"
			} else if remDays == 0 {
				if priority > 2 {
					priority = 2
					status = "DUE"
				}
			} else if remDays <= 7 {
				if priority > 3 {
					priority = 3
					status = "CRITICAL"
				}
			} else if remDays <= 30 {
				if priority > 4 {
					priority = 4
					status = "UPCOMING"
				}
			}
		}
	}

	item.Status = status
	item.Priority = priority
}
