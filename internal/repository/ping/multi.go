package ping

import (
	"context"
)

type multiPinger struct {
	pingers []Pinger
}

// NewMulti creates a Pinger that checks every provided dependency in order and
// fails fast on the first unhealthy one. bookmark-service uses this to verify
// both the database and Redis in a single health check.
func NewMulti(pingers ...Pinger) Pinger {
	return &multiPinger{
		pingers: pingers,
	}
}

// Ping pings each underlying dependency and returns the first error encountered.
func (m *multiPinger) Ping(ctx context.Context) error {
	for _, p := range m.pingers {
		if err := p.Ping(ctx); err != nil {
			return err
		}
	}

	return nil
}
