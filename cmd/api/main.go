package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"strings"

	"github.com/ilhamasy/gastrack-ms/internal/config"
	"github.com/ilhamasy/gastrack-ms/internal/database"
	"github.com/ilhamasy/gastrack-ms/internal/handler"
	"github.com/ilhamasy/gastrack-ms/internal/middleware"
	"github.com/ilhamasy/gastrack-ms/internal/service"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Println("Configuration loaded")

	// 2. Connect to the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connection established")

	// 3. Run migrations
	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// 4. Setup services and handlers
	authService := service.NewAuthService(db.Pool, cfg.JWTSecret)
	maintenanceService := service.NewMaintenanceService(db.Pool)
	vehicleService := service.NewVehicleService(db.Pool, maintenanceService)
	
	mux := http.NewServeMux()
	
	healthHandler := handler.NewHealthHandler(db)
	mux.Handle("/health", healthHandler)

	authHandler := handler.NewAuthHandler(authService)
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)

	odometerService := service.NewOdometerService(db.Pool)
	odometerHandler := handler.NewOdometerHandler(odometerService)

	maintenanceHandler := handler.NewMaintenanceHandler(maintenanceService, vehicleService)
	serviceRecordService := service.NewServiceRecordService(db.Pool)
	serviceRecordHandler := handler.NewServiceRecordHandler(serviceRecordService)

	vehicleHandler := handler.NewVehicleHandler(vehicleService)
	authMiddleware := middleware.Auth([]byte(cfg.JWTSecret))
	
	expenseService := service.NewExpenseService(db.Pool)
	expenseHandler := handler.NewExpenseHandler(expenseService)

	recommendationService := service.NewRecommendationService(db.Pool)
	recommendationHandler := handler.NewRecommendationHandler(recommendationService)

	mux.Handle("/api/vehicles", authMiddleware(vehicleHandler))
	mux.Handle("/api/vehicles/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if strings.Contains(r.URL.Path, "/odometer") {
			odometerHandler.ServeHTTP(w, r)
			return
		}
		if strings.Contains(r.URL.Path, "/maintenance") {
			maintenanceHandler.ServeHTTP(w, r)
			return
		}
		if strings.Contains(r.URL.Path, "/service-records") {
			serviceRecordHandler.ServeHTTP(w, r)
			return
		}
		if strings.Contains(r.URL.Path, "/expenses") {
			expenseHandler.GetExpenseAnalytics(w, r)
			return
		}
		if strings.Contains(r.URL.Path, "/recommendations") {
			recommendationHandler.GetRecommendations(w, r)
			return
		}
		vehicleHandler.ServeHTTP(w, r)
	})))

	// 5. Setup HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	// 6. Graceful shutdown
	go func() {
		log.Printf("Server starting on port %s...", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
