package service

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationService struct {
	db                 *pgxpool.Pool
	preferencesService *PreferencesService
}

func NewNotificationService(db *pgxpool.Pool, prefsService *PreferencesService) *NotificationService {
	return &NotificationService{db: db, preferencesService: prefsService}
}

// CheckAndSendMaintenanceReminders checks if a user should be notified about maintenance.
// Typically this would be called by a cron job or after an odometer update.
func (s *NotificationService) CheckAndSendMaintenanceReminders(ctx context.Context, userID string, vehicleID string) error {
	prefs, err := s.preferencesService.GetPreferences(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get preferences: %w", err)
	}

	if !prefs.PushEnabled && !prefs.EmailEnabled {
		return nil // User opted out
	}

	// Example logic: Just log a mock notification for now.
	// In a real app, this would query vehicle_maintenance for due items and send FCM/Email.
	
	if prefs.PushEnabled {
		log.Printf("[PUSH NOTIFICATION] Sending maintenance reminder to User: %s for Vehicle: %s", userID, vehicleID)
	}
	
	if prefs.EmailEnabled {
		log.Printf("[EMAIL NOTIFICATION] Sending maintenance reminder to User: %s for Vehicle: %s", userID, vehicleID)
	}

	return nil
}
