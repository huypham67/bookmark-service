package ping

import (
	"context"

	"gorm.io/gorm"
)

type sqlDBPinger struct {
	db *gorm.DB
}

// NewSQLDB creates a new SQLDBPinger with the provided GORM database client.
func NewSQLDB(db *gorm.DB) Pinger {
	return &sqlDBPinger{
		db: db,
	}
}

// Ping checks the health of the database connection by issuing a PingContext and returning any error encountered.
func (p *sqlDBPinger) Ping(ctx context.Context) error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.PingContext(ctx)
}
