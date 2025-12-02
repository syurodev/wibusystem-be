package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	"system/configs"
	"system/internal/platform/database"
	"system/internal/platform/logger"
)

func main() {
	// 1. Load configuration từ .env file
	cfg, err := configs.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize logger
	appLogger, err := logger.InitLogger(false)
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer logger.SyncLogger()

	appLogger.Info("Starting ClickHouse initialization...")

	// 3. Connect to ClickHouse
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ch, err := database.NewClickHouseClient(ctx, &cfg.ClickHouse, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to ClickHouse", zap.Error(err))
	}
	defer ch.Close()

	appLogger.Info("Connected to ClickHouse successfully")

	// 4. Read migration file
	migrationPath := "migrations/clickhouse/000001_create_view_events.sql"
	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		appLogger.Fatal("Failed to read migration file",
			zap.String("path", migrationPath),
			zap.Error(err))
	}

	appLogger.Info("Read migration file", zap.String("path", migrationPath))

	// 5. Execute migration
	appLogger.Info("Executing migration...")
	
	// Split statements by semicolon
	// Note: This is a simple split and might break if semicolons are inside strings/comments
	// For production, use a proper SQL parser or migration tool
	statements := strings.Split(string(sqlBytes), ";")

	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		appLogger.Info("Executing statement", zap.Int("index", i))
		if err := ch.Conn.Exec(ctx, stmt); err != nil {
			appLogger.Fatal("Failed to execute statement", 
				zap.Int("index", i),
				zap.String("statement", stmt),
				zap.Error(err))
		}
	}

	appLogger.Info("Migration executed successfully")

	// 6. Verify tables created
	var count uint64
	row := ch.Conn.QueryRow(ctx, "SELECT count() FROM system.tables WHERE database = ? AND name = 'view_events'", cfg.ClickHouse.Database)
	if err := row.Scan(&count); err != nil {
		appLogger.Fatal("Failed to verify view_events table", zap.Error(err))
	}

	if count == 0 {
		appLogger.Fatal("view_events table was not created")
	}

	appLogger.Info("Verified view_events table exists")

	// 7. Verify materialized view
	row = ch.Conn.QueryRow(ctx, "SELECT count() FROM system.tables WHERE database = ? AND name = 'view_events_daily'", cfg.ClickHouse.Database)
	if err := row.Scan(&count); err != nil {
		appLogger.Fatal("Failed to verify view_events_daily table", zap.Error(err))
	}

	if count == 0 {
		appLogger.Fatal("view_events_daily materialized view was not created")
	}

	appLogger.Info("Verified view_events_daily materialized view exists")

	// 8. Success!
	appLogger.Info("✅ ClickHouse initialization completed successfully!")
	appLogger.Info("Tables created:",
		zap.String("main_table", "view_events"),
		zap.String("materialized_view", "view_events_daily"))
}
