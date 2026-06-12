package fixtures

import (
	"testing"

	"github.com/huypham67/bookmark-common/pkg/sqldb"
	"gorm.io/gorm"
)

type TestDatabase interface {
	SetupDB(db *gorm.DB)
	MigrateDB() error
	SeedData() error
	GetDB() *gorm.DB
}

type baseTestDB struct {
	db *gorm.DB
}

func (b *baseTestDB) SetupDB(db *gorm.DB) {
	b.db = db
}

func (b *baseTestDB) GetDB() *gorm.DB {
	return b.db
}

// NewTestDB creates a new test database, runs migrations, and seeds data for testing.
func NewTestDB(t *testing.T, testdb TestDatabase) *gorm.DB {
	testdb.SetupDB(sqldb.NewMock(t))

	err := testdb.MigrateDB()

	if err != nil {
		t.Fatal("failed to migrate database:", err)
	}

	err = testdb.SeedData()

	if err != nil {
		t.Fatal("failed to generate data:", err)
	}

	return testdb.GetDB()
}
