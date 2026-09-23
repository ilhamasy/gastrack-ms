package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/ilhamasy/gastrack-ms/internal/database"
	"github.com/ilhamasy/gastrack-ms/internal/model"
)

type PreferencesService struct {
	db database.DBPool
}

func NewPreferencesService(db database.DBPool) *PreferencesService {
	return &PreferencesService{db: db}
}

func (s *PreferencesService) GetPreferences(ctx context.Context, userID string) (*model.NotificationPreferences, error) {
	var prefs model.NotificationPreferences
	err := s.db.QueryRow(ctx, "SELECT email_enabled, push_enabled FROM notification_preferences WHERE user_id = $1", userID).Scan(&prefs.EmailEnabled, &prefs.PushEnabled)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			// Create default preferences
			_, err = s.db.Exec(ctx, "INSERT INTO notification_preferences (user_id, email_enabled, push_enabled) VALUES ($1, true, true)", userID)
			if err != nil {
				return nil, fmt.Errorf("failed to create default preferences: %w", err)
			}
			return &model.NotificationPreferences{EmailEnabled: true, PushEnabled: true}, nil
		}
		return nil, fmt.Errorf("failed to fetch preferences: %w", err)
	}

	return &prefs, nil
}

func (s *PreferencesService) UpdatePreferences(ctx context.Context, userID string, prefs *model.NotificationPreferences) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO notification_preferences (user_id, email_enabled, push_enabled) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) 
		DO UPDATE SET email_enabled = EXCLUDED.email_enabled, push_enabled = EXCLUDED.push_enabled, updated_at = CURRENT_TIMESTAMP
	`, userID, prefs.EmailEnabled, prefs.PushEnabled)
	
	if err != nil {
		return fmt.Errorf("failed to update preferences: %w", err)
	}
	return nil
}
