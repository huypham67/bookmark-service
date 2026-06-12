package ping

import (
	"context"
)

// Pinger defines the contract for a service that can check the health of a data store (e.g., Redis).
type Pinger interface {
	Ping(ctx context.Context) error
}
