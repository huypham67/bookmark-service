package redis

import "github.com/stretchr/testify/mock"

type MockPinger struct {
	mock.Mock
}

func (m *MockPinger) Ping() error {
	args := m.Called()
	return args.Error(0)
}
