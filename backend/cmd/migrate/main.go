package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/store-platform/store/internal/platform/config"
)

func main() {
	direction := flag.String("direction", "up", "up or down")
	steps := flag.Int("steps", 0, "number of steps (0 = all)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = filepath.Join("migrations")
	}
	sourceURL := fmt.Sprintf("file://%s", migrationsPath)
	dbURL := cfg.DatabaseURL
	if strings.HasPrefix(dbURL, "postgres://") {
		dbURL = "pgx5://" + strings.TrimPrefix(dbURL, "postgres://")
	}
	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	ctx := context.Background()
	_ = ctx

	var migErr error
	switch *direction {
	case "up":
		if *steps > 0 {
			migErr = m.Steps(*steps)
		} else {
			migErr = m.Up()
		}
	case "down":
		if *steps > 0 {
			migErr = m.Steps(-*steps)
		} else {
			migErr = m.Down()
		}
	default:
		log.Fatalf("unknown direction %s", *direction)
	}
	if migErr != nil && migErr != migrate.ErrNoChange {
		log.Fatal(migErr)
	}
	log.Println("migrations applied")
}
