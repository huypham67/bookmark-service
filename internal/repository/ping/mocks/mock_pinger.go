package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type Pinger struct {
	mock.Mock
}

func (m *Pinger) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
