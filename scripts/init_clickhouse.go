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

	// 4. Read migration files
	migrationDir := "migrations/clickhouse"
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		appLogger.Fatal("Failed to read migration directory", zap.Error(err))
	}

	// Filter for .sql files
	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	appLogger.Info("Found migration files", zap.Int("count", len(migrationFiles)))

	// 5. Execute migrations
	for _, fileName := range migrationFiles {
		migrationPath := migrationDir + "/" + fileName
		appLogger.Info("Processing migration file", zap.String("path", migrationPath))

		sqlBytes, err := os.ReadFile(migrationPath)
		if err != nil {
			appLogger.Fatal("Failed to read migration file",
				zap.String("path", migrationPath),
				zap.Error(err))
		}

		// Split statements by semicolon
		// Note: This is a simple split and might break if semicolons are inside strings/comments
		// For production, use a proper SQL parser or migration tool
		statements := strings.Split(string(sqlBytes), ";")

		for i, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			// Detailed logging for debugging
			// appLogger.Info("Executing statement", zap.Int("index", i))
			if err := ch.Conn.Exec(ctx, stmt); err != nil {
				appLogger.Fatal("Failed to execute statement",
					zap.String("file", fileName),
					zap.Int("index", i),
					zap.String("statement", stmt),
					zap.Error(err))
			}
		}
		appLogger.Info("Successfully executed migration file", zap.String("file", fileName))
	}

	appLogger.Info("All migrations executed successfully")

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
