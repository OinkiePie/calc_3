package jwt_manager

import "github.com/stretchr/testify/mock"

type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) Generate(userID int64) (string, string, int64, error) {
	args := m.Called(userID)
	return args.String(0), args.String(1), args.Get(2).(int64), args.Error(3)
}

func (m *MockJWTManager) Validate(tokenString string) (Claims, error) {
	args := m.Called(tokenString)
	return args.Get(0).(Claims), args.Error(1)
}
