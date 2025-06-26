package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Init() error {
	fmt.Println("🔌 Attempting to connect to database...")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL not set")
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse DB config: %w", err)
	}

	// ✅ Disable statement caching for Supabase
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	DB, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	fmt.Println("✅ Connected to database.")
	return nil
}
