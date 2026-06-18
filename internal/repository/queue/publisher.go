package queue

import (
	"context"
)

// Publisher pushes already-serialized job payloads onto a named queue.
//
//go:generate mockery --name=Publisher --output=./mocks --outpkg=mocks --filename=mock_publisher.go
type Publisher interface {
	Enqueue(ctx context.Context, queue string, payloads ...[]byte) error
}
