package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ilhamasy/gastrack-ms/internal/model"
)

type ExpenseService struct {
	db *pgxpool.Pool
}

func NewExpenseService(db *pgxpool.Pool) *ExpenseService {
	return &ExpenseService{db: db}
}

func (s *ExpenseService) GetExpenseAnalytics(ctx context.Context, vehicleID string) (*model.ExpenseAnalytics, error) {
	analytics := &model.ExpenseAnalytics{
		MonthlyExpenses:  []model.MonthlyExpense{},
		CategoryExpenses: []model.CategoryExpense{},
	}

	monthlyQuery := `
		SELECT TO_CHAR(service_date, 'YYYY-MM') as month, SUM(total_cost) as total
		FROM service_records
		WHERE vehicle_id = $1 AND service_date >= CURRENT_DATE - INTERVAL '6 months'
		GROUP BY month
		ORDER BY month ASC
	`
	rows, err := s.db.Query(ctx, monthlyQuery, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query monthly expenses: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m model.MonthlyExpense
		if err := rows.Scan(&m.Month, &m.TotalCost); err != nil {
			return nil, fmt.Errorf("failed to scan monthly expense: %w", err)
		}
		analytics.MonthlyExpenses = append(analytics.MonthlyExpenses, m)
	}

	categoryQuery := `
		SELECT i.item_name as category, SUM(i.cost * i.quantity) as total
		FROM service_items i
		JOIN service_records r ON i.service_record_id = r.id
		WHERE r.vehicle_id = $1 AND r.service_date >= CURRENT_DATE - INTERVAL '6 months'
		GROUP BY i.item_name
		ORDER BY total DESC
	`
	catRows, err := s.db.Query(ctx, categoryQuery, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query category expenses: %w", err)
	}
	defer catRows.Close()

	for catRows.Next() {
		var c model.CategoryExpense
		if err := catRows.Scan(&c.Category, &c.TotalCost); err != nil {
			return nil, fmt.Errorf("failed to scan category expense: %w", err)
		}
		analytics.CategoryExpenses = append(analytics.CategoryExpenses, c)
	}

	return analytics, nil
}
