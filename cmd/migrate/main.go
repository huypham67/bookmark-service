package main

import (
	"flag"

	"github.com/rs/zerolog/log"

	"github.com/huypham67/bookmark-common/pkg/logger"
	"github.com/huypham67/bookmark-common/pkg/sqldb"
)

const migrationPath = "migrations"

func main() {
	direction := flag.String("direction", "up", "migration direction when applying steps: 'up' or 'down'")
	steps := flag.Int("steps", 0, "number of migration steps to apply (0 = all pending, valid with 'up')")
	force := flag.Int("force", -1, "force the schema to a specific version and clear the dirty flag (recovery); -1 disables")
	showVersion := flag.Bool("version", false, "print the current migration version and dirty state, then exit")
	flag.Parse()

	if err := logger.NewClient(""); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize logger")
	}

	dbClient, err := sqldb.NewClient("")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize postgres client")
	}

	switch {
	case *showVersion:
		version, dirty, vErr := sqldb.PostgresMigrationVersion(dbClient, migrationPath)
		if vErr != nil {
			log.Fatal().Err(vErr).Msg("failed to read migration version")
		}
		log.Info().Uint("version", version).Bool("dirty", dirty).Msg("current migration status")

	case *force >= 0:
		if err := sqldb.ForcePostgresDB(dbClient, migrationPath, *force); err != nil {
			log.Fatal().Err(err).Int("version", *force).Msg("failed to force migration version")
		}
		log.Info().Int("version", *force).Msg("migration version forced; dirty flag cleared")

	case *steps == 0:
		if err := sqldb.MigratePostgresDB(dbClient, migrationPath); err != nil {
			log.Fatal().Err(err).Msg("failed to apply migrations")
		}
		log.Info().Msg("all pending migrations applied")

	default:
		if err := sqldb.MigratePostgresDBWithSteps(dbClient, migrationPath, *direction, *steps); err != nil {
			log.Fatal().Err(err).Str("direction", *direction).Int("steps", *steps).Msg("failed to apply migration steps")
		}
		log.Info().Str("direction", *direction).Int("steps", *steps).Msg("migration steps applied")
	}
}
