package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ilhamasy/gastrack-ms/internal/database"
	"github.com/ilhamasy/gastrack-ms/internal/model"
)

type ServiceRecordService struct {
	db database.DBPool
}

func NewServiceRecordService(db database.DBPool) *ServiceRecordService {
	return &ServiceRecordService{db: db}
}

func (s *ServiceRecordService) AddServiceRecord(ctx context.Context, vehicleID string, record model.ServiceRecord) (*model.ServiceRecord, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Calculate total cost if not provided accurately (or trust the input, let's just use input for now or sum up items)
	var totalCost float64
	for _, item := range record.Items {
		totalCost += item.Cost * item.Quantity
	}
	if record.TotalCost == 0 && totalCost > 0 {
		record.TotalCost = totalCost
	}

	recordID := uuid.New()
	query := `
		INSERT INTO service_records (id, vehicle_id, service_date, odometer_km, workshop_name, total_cost, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`
	err = tx.QueryRow(ctx, query,
		recordID, vehicleID, record.ServiceDate, record.OdometerKm, record.WorkshopName, record.TotalCost, record.Notes,
	).Scan(&record.CreatedAt, &record.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert service record: %w", err)
	}
	record.ID = recordID
	record.VehicleID = uuid.MustParse(vehicleID)

	for i, item := range record.Items {
		itemID := uuid.New()
		itemQuery := `
			INSERT INTO service_items (id, service_record_id, maintenance_id, item_name, brand, product, part_number, quantity, cost, notes)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING created_at
		`
		err = tx.QueryRow(ctx, itemQuery,
			itemID, recordID, item.MaintenanceID, item.ItemName, item.Brand, item.Product, item.PartNumber, item.Quantity, item.Cost, item.Notes,
		).Scan(&item.CreatedAt)
		
		if err != nil {
			return nil, fmt.Errorf("failed to insert service item: %w", err)
		}
		
		record.Items[i].ID = itemID
		record.Items[i].ServiceRecordID = recordID
		record.Items[i].CreatedAt = item.CreatedAt

		// Update vehicle_maintenance if linked
		if item.MaintenanceID != nil {
			updateMaintQuery := `
				UPDATE vehicle_maintenance
				SET last_service_km = $1, last_service_date = $2, updated_at = CURRENT_TIMESTAMP
				WHERE id = $3
			`
			_, err = tx.Exec(ctx, updateMaintQuery, record.OdometerKm, record.ServiceDate, item.MaintenanceID)
			if err != nil {
				return nil, fmt.Errorf("failed to update vehicle maintenance: %w", err)
			}
		}
	}

	// Update vehicle current_odometer if this service's odometer is higher
	updateOdoQuery := `
		UPDATE vehicles
		SET current_odometer = GREATEST(current_odometer, $1), updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err = tx.Exec(ctx, updateOdoQuery, record.OdometerKm, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to update vehicle odometer: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &record, nil
}

func (s *ServiceRecordService) GetServiceRecords(ctx context.Context, vehicleID string) ([]model.ServiceRecord, error) {
	query := `
		SELECT 
			r.id, r.vehicle_id, r.service_date, r.odometer_km, r.workshop_name, r.total_cost, r.notes, r.created_at, r.updated_at,
			i.id, i.maintenance_id, i.item_name, i.brand, i.product, i.part_number, i.quantity, i.cost, i.notes, i.created_at
		FROM service_records r
		LEFT JOIN service_items i ON r.id = i.service_record_id
		WHERE r.vehicle_id = $1
		ORDER BY r.service_date DESC, r.created_at DESC
	`

	rows, err := s.db.Query(ctx, query, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query service records: %w", err)
	}
	defer rows.Close()

	recordMap := make(map[uuid.UUID]*model.ServiceRecord)
	var orderedIDs []uuid.UUID

	for rows.Next() {
		var r model.ServiceRecord
		var i model.ServiceItem
		var itemID *uuid.UUID

		err := rows.Scan(
			&r.ID, &r.VehicleID, &r.ServiceDate, &r.OdometerKm, &r.WorkshopName, &r.TotalCost, &r.Notes, &r.CreatedAt, &r.UpdatedAt,
			&itemID, &i.MaintenanceID, &i.ItemName, &i.Brand, &i.Product, &i.PartNumber, &i.Quantity, &i.Cost, &i.Notes, &i.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if _, exists := recordMap[r.ID]; !exists {
			recordMap[r.ID] = &model.ServiceRecord{
				ID:           r.ID,
				VehicleID:    r.VehicleID,
				ServiceDate:  r.ServiceDate,
				OdometerKm:   r.OdometerKm,
				WorkshopName: r.WorkshopName,
				TotalCost:    r.TotalCost,
				Notes:        r.Notes,
				Items:        []model.ServiceItem{},
				CreatedAt:    r.CreatedAt,
				UpdatedAt:    r.UpdatedAt,
			}
			orderedIDs = append(orderedIDs, r.ID)
		}

		if itemID != nil {
			i.ID = *itemID
			i.ServiceRecordID = r.ID
			recordMap[r.ID].Items = append(recordMap[r.ID].Items, i)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	var records []model.ServiceRecord
	for _, id := range orderedIDs {
		records = append(records, *recordMap[id])
	}

	return records, nil
}
