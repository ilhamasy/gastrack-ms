package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ilhamasy/gastrack-ms/internal/model"
)

type RecommendationService struct {
	db *pgxpool.Pool
}

func NewRecommendationService(db *pgxpool.Pool) *RecommendationService {
	return &RecommendationService{db: db}
}

func (s *RecommendationService) GetRecommendations(ctx context.Context, vehicleID string) ([]model.Recommendation, error) {
	recs := []model.Recommendation{}

	// 1. Check for overdue high priority maintenance
	mQuery := `
		SELECT id, name, last_service_km, interval_km, last_service_date, interval_months
		FROM vehicle_maintenance
		WHERE vehicle_id = $1
	`
	rows, err := s.db.Query(ctx, mQuery, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query maintenance: %w", err)
	}
	defer rows.Close()

	// Get current odometer to calculate remaining km
	var currentOdometer int
	var vehicleUpdatedAt time.Time
	err = s.db.QueryRow(ctx, "SELECT current_odometer, updated_at FROM vehicles WHERE id = $1", vehicleID).Scan(&currentOdometer, &vehicleUpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to query vehicle: %w", err)
	}

	overdueCount := 0
	for rows.Next() {
		var id, name string
		var lastSrvKm, intKm, intMonths *int
		var lastSrvDate *time.Time
		if err := rows.Scan(&id, &name, &lastSrvKm, &intKm, &lastSrvDate, &intMonths); err != nil {
			return nil, err
		}

		due := false
		if lastSrvKm != nil && intKm != nil {
			if currentOdometer-*lastSrvKm >= *intKm {
				due = true
			}
		}
		if lastSrvDate != nil && intMonths != nil {
			nextDate := lastSrvDate.AddDate(0, *intMonths, 0)
			if time.Now().After(nextDate) {
				due = true
			}
		}

		if due {
			overdueCount++
		}
	}

	if overdueCount > 0 {
		recs = append(recs, model.Recommendation{
			ID:          "rec_maintenance_due",
			Title:       "Maintenance Due",
			Description: fmt.Sprintf("You have %d maintenance item(s) due or overdue.", overdueCount),
			ActionLabel: "View Maintenance",
			ActionURL:   "/maintenance",
			Priority:    1,
		})
	}

	// 2. Check if odometer needs update
	if time.Since(vehicleUpdatedAt).Hours() > 24*14 { // 14 days
		recs = append(recs, model.Recommendation{
			ID:          "rec_odometer_update",
			Title:       "Update Odometer",
			Description: "It's been over 2 weeks since you last updated your odometer.",
			ActionLabel: "Update Now",
			ActionURL:   "/dashboard",
			Priority:    2,
		})
	}

	// 3. General tip
	recs = append(recs, model.Recommendation{
		ID:          "rec_general_tip",
		Title:       "Eco-Driving Tip",
		Description: "Keep your tires properly inflated to improve fuel efficiency and extend tire life.",
		Priority:    3,
	})

	return recs, nil
}
