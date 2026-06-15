package fixtures

import (
	"embed"
	"testing"
)

//go:embed testdata/*.csv
var csvFS embed.FS

// Shared CSV fixture names. Reuse these across service, handler and integration
// tests so the same input drives every layer (and there is one place to edit).
const (
	CSVBookmarksValid   = "bookmarks_valid.csv"   // header + 2 valid rows
	CSVBookmarksInvalid = "bookmarks_invalid.csv" // header + 1 row with a bad URL
	CSVBookmarksEmpty   = "bookmarks_empty.csv"   // header only, no data rows
)

// ReadCSV returns the bytes of a shared CSV fixture from testdata/.
func ReadCSV(t *testing.T, name string) []byte {
	t.Helper()

	data, err := csvFS.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read csv fixture %q: %v", name, err)
	}
	return data
}
