package bookmark

import (
	"testing"

	"github.com/huypham67/bookmark-service/internal/test/fixtures"
	"gorm.io/gorm"
)

func newTestRepository(t *testing.T) (Repository, *gorm.DB) {
	t.Helper()

	testDB := fixtures.NewTestDB(t, &fixtures.BookmarkTestDB{})
	repo := NewRepository(testDB)

	return repo, testDB
}
