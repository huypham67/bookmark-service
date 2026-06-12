package main

import (
	"flag"

	"github.com/rs/zerolog/log"

	"github.com/huypham67/bookmark-common/pkg/logger"
	"github.com/huypham67/bookmark-common/pkg/sqldb"
)

func main() {
	direction := flag.String("direction", "up", "migration direction: 'up' or 'down'")
	steps := flag.Int("steps", 0, "number of migration steps (0 means all)")
	flag.Parse()

	if err := logger.NewClient(""); err != nil {
		log.Error().Err(err).Msg("failed to initialize logger")
		return
	}

	dbClient, err := sqldb.NewClient("")
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize postgres client")
		return
	}

	var migrateErr error
	switch {
	case *steps == 0:
		migrateErr = sqldb.MigratePostgresDB(dbClient, "migrations")
	default:
		migrateErr = sqldb.MigratePostgresDBWithSteps(dbClient, "migrations", *direction, *steps)
	}

	if migrateErr != nil {
		log.Error().Err(migrateErr).Msg("failed to run database migrations")
		return
	}

	log.Info().
		Str("direction", *direction).
		Int("steps", *steps).
		Msg("database migrations completed successfully")
}
