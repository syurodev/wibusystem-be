//go:build ignore

// Script to manually trigger rank snapshot creation for testing.
// Usage: go run scripts/trigger_snapshot.go [period] [entityType]
// Examples:
//   go run scripts/trigger_snapshot.go week novel
//   go run scripts/trigger_snapshot.go month creator

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"system/configs"
	"system/internal/modules/analytics"
	"system/internal/platform/database"

	"go.uber.org/zap"
)

func main() {
	// Load config
	cfg, err := configs.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Connect to ClickHouse
	ctx := context.Background()
	ch, err := database.NewClickHouseClient(ctx, &cfg.ClickHouse, logger)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer ch.Close()

	// Create analytics repository
	repo := analytics.NewViewAnalyticsClickHouseRepository(ch)

	// Get arguments
	period := "week"
	entityType := "novel"
	limit := 500

	if len(os.Args) > 1 {
		period = os.Args[1]
	}
	if len(os.Args) > 2 {
		entityType = os.Args[2]
	}

	fmt.Printf("Creating rank snapshot: period=%s, entityType=%s, limit=%d\n", period, entityType, limit)

	snapshotCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	err = repo.CreateRankSnapshot(snapshotCtx, time.Now(), period, entityType, limit)
	if err != nil {
		log.Fatalf("Failed to create rank snapshot: %v", err)
	}

	fmt.Println("✅ Rank snapshot created successfully!")
}
