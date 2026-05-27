package sqldb

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDBClient initializes and returns a new database client based on the provided configuration.
func NewDBClient(envPrefix string) (*gorm.DB, error) {
	cfg, err := LoadDBConfig(envPrefix)
	if err != nil {
		return nil, err
	}
	dsn := getDSN(cfg)

	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func getDSN(cfg *DBConfig) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
		cfg.TimeZone,
	)
}
