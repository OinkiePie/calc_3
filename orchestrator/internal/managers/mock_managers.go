package managers

import (
	"context"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/stretchr/testify/mock"
)

type MockUserManager struct {
	mock.Mock
}

func (m *MockUserManager) Register(ctx context.Context, login, password string) (int64, error, int) {
	args := m.Called(ctx, login, password)
	return args.Get(0).(int64), args.Error(1), args.Get(2).(int)
}

func (m *MockUserManager) Login(ctx context.Context, login, password string) (string, int64, error, int) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Get(1).(int64), args.Error(2), args.Int(3)
}

func (m *MockUserManager) Logout(ctx context.Context, jti string) (error, int) {
	args := m.Called(ctx, jti)
	return args.Error(0), args.Int(1)
}

func (m *MockUserManager) Delete(ctx context.Context, login, password string) (int64, error, int) {
	args := m.Called(ctx, login, password)
	return args.Get(0).(int64), args.Error(1), args.Int(2)
}

func (m *MockUserManager) SessionExists(ctx context.Context, jti string) (error, bool) {
	args := m.Called(ctx, jti)
	return args.Error(0), args.Bool(1)
}

type MockExpressionManager struct {
	mock.Mock
}

func (m *MockExpressionManager) AddExpression(ctx context.Context, expressionString string, claims int64) (int64, error, int) {
	args := m.Called(ctx, expressionString, claims)
	return args.Get(0).(int64), args.Error(1), args.Int(2)
}

func (m *MockExpressionManager) ReadExpressions(ctx context.Context, id int64) ([]*models.Expression, error, int) {
	args := m.Called(ctx, id)
	return args.Get(0).([]*models.Expression), args.Error(1), args.Int(2)
}

func (m *MockExpressionManager) ReadExpression(ctx context.Context, id int64) (*models.Expression, error, int) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Expression), args.Error(1), args.Int(2)
}

func (m *MockExpressionManager) ReadTask(ctx context.Context) (*models.Task, error, int) {
	args := m.Called(ctx)
	return args.Get(0).(*models.Task), args.Error(1), args.Int(2)
}

func (m *MockExpressionManager) CompleteTask(ctx context.Context, taskCompleted *models.TaskCompleted) (error, int) {
	args := m.Called(ctx, taskCompleted)
	return args.Error(0), args.Int(1)
}
