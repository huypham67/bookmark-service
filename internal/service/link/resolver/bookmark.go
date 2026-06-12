package resolver

import "context"

// Bookmark resolves a bookmark's original URL by its short code.
// It is the minimal slice of the bookmark repository the redirect flow needs,
// keeping the link service decoupled from the rest of the bookmark layer.
//
//go:generate mockery --name=Bookmark --output=./mocks --outpkg=mocks --filename=mock_bookmark_resolver.go
type Bookmark interface {
	GetURLByCode(ctx context.Context, code string) (string, error)
}
