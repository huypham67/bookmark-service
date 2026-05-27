package testutil

import (
	"testing"

	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/pkg/sqldb"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func SetupUserTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	db := sqldb.CreateMockDB(t)

	err := db.AutoMigrate(&model.User{})
	require.NoError(t, err)

	users := NewUserFixtures()

	err = db.Create(&users).Error
	require.NoError(t, err)

	return db
}
