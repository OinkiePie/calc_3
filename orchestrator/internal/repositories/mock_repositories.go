package repositories

import (
	"context"
	"database/sql"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/stretchr/testify/mock"
)

func Float64Ptr(i float64) *float64 {
	return &i
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, login, pas string) (*models.User, error, int) {
	args := m.Called(ctx, login, pas)
	return args.Get(0).(*models.User), args.Error(1), args.Int(2)
}

func (m *MockUserRepository) ReadUserByLogin(ctx context.Context, login string) (*models.User, error, int) {
	args := m.Called(ctx, login)
	return args.Get(0).(*models.User), args.Error(1), args.Int(2)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id int64) (error, int) {
	args := m.Called(ctx, id)
	return args.Error(0), args.Int(1)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) CreateSession(ctx context.Context, tx *sql.Tx, jti string, sub, exp int64) (error, int) {
	args := m.Called(ctx, tx, jti, sub, exp)
	return args.Error(0), args.Int(1)
}

func (m *MockSessionRepository) ReadSession(ctx context.Context, jti string) (*models.Session, error, int) {
	args := m.Called(ctx, jti)
	return args.Get(0).(*models.Session), args.Error(1), args.Int(2)
}

func (m *MockSessionRepository) DeleteSession(ctx context.Context, jti string) (error, int) {
	args := m.Called(ctx, jti)
	return args.Error(0), args.Int(1)
}

type MockExpressionsRepository struct {
	mock.Mock
}

func (m *MockExpressionsRepository) CreateExpression(ctx context.Context, tx *sql.Tx, expr *models.Expression) (int64, error, int) {
	args := m.Called(ctx, tx, expr)
	return args.Get(0).(int64), args.Error(1), args.Int(2)
}

func (m *MockExpressionsRepository) ReadExpressionByID(ctx context.Context, tx *sql.Tx, id int64) (*models.Expression, error, int) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(*models.Expression), args.Error(1), args.Int(2)
}

func (m *MockExpressionsRepository) ReadExpressionsByUserID(ctx context.Context, tx *sql.Tx, userID int64) ([]*models.Expression, error, int) {
	args := m.Called(ctx, tx, userID)
	return args.Get(0).([]*models.Expression), args.Error(1), args.Int(2)
}

func (m *MockExpressionsRepository) ReadExpressionTasks(ctx context.Context, tx *sql.Tx, id int64) ([]*models.Task, error, int) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).([]*models.Task), args.Error(1), args.Int(2)
}

func (m *MockExpressionsRepository) UpdateExpressionStatus(ctx context.Context, tx *sql.Tx, id int64, status string) (error, int) {
	args := m.Called(ctx, tx, id, status)
	return args.Error(0), args.Int(1)
}

func (m *MockExpressionsRepository) UpdateExpressionError(ctx context.Context, tx *sql.Tx, id int64, errContent string) (error, int) {
	args := m.Called(ctx, tx, id, errContent)
	return args.Error(0), args.Int(1)
}

func (m *MockExpressionsRepository) UpdateExpressionResult(ctx context.Context, tx *sql.Tx, id int64, result float64) (error, int) {
	args := m.Called(ctx, tx, id, result)
	return args.Error(0), args.Int(1)
}

type MockTasksRepository struct {
	mock.Mock
}

func (m *MockTasksRepository) CreateTask(ctx context.Context, tx *sql.Tx, task *models.Task) (int64, error, int) {
	args := m.Called(ctx, tx, task)
	return args.Get(0).(int64), args.Error(1), args.Int(2)
}

func (m *MockTasksRepository) ReadTaskByID(ctx context.Context, tx *sql.Tx, id int64) (*models.Task, error, int) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(*models.Task), args.Error(1), args.Int(2)
}

func (m *MockTasksRepository) ReadTasksByExpressionID(ctx context.Context, tx *sql.Tx, expressionID int64) ([]*models.Task, error, int) {
	args := m.Called(ctx, tx, expressionID)
	return args.Get(0).([]*models.Task), args.Error(1), args.Int(2)
}

func (m *MockTasksRepository) ReadUncompletedTasks(ctx context.Context, tx *sql.Tx) ([]*models.Task, error, int) {
	args := m.Called(ctx, tx)
	return args.Get(0).([]*models.Task), args.Error(1), args.Int(2)
}

func (m *MockTasksRepository) UpdateTaskDependencies(ctx context.Context, tx *sql.Tx, task *models.Task) (error, int) {
	args := m.Called(ctx, tx, task)
	return args.Error(0), args.Int(1)
}

func (m *MockTasksRepository) UpdateTaskArguments(ctx context.Context, tx *sql.Tx, id int64, index int, value *float64) (error, int) {
	args := m.Called(ctx, tx, id, index, value)
	return args.Error(0), args.Int(1)
}

func (m *MockTasksRepository) UpdateTaskStatus(ctx context.Context, tx *sql.Tx, id int64, status string) (error, int) {
	args := m.Called(ctx, tx, id, status)
	return args.Error(0), args.Int(1)
}

func (m *MockTasksRepository) UpdateTaskExpressionID(ctx context.Context, tx *sql.Tx, id, exprId int64) (error, int) {
	args := m.Called(ctx, tx, id, exprId)
	return args.Error(0), args.Int(1)
}

func (m *MockTasksRepository) UpdateTaskResult(ctx context.Context, tx *sql.Tx, result float64, id int64) (error, int) {
	args := m.Called(ctx, tx, result, id)
	return args.Error(0), args.Int(1)
}

func (m *MockTasksRepository) DeleteTasks(ctx context.Context, tx *sql.Tx, id int64) (error, int) {
	args := m.Called(ctx, tx, id)
	return args.Error(0), args.Int(1)
}

type MockArgsRepository struct {
	mock.Mock
}

func (m *MockArgsRepository) CreateTaskArgs(ctx context.Context, tx *sql.Tx, task *models.Task) error {
	args := m.Called(ctx, tx, task)
	return args.Error(0)
}

func (m *MockArgsRepository) ReadTaskArgs(ctx context.Context, tx *sql.Tx, id int64) ([]*float64, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).([]*float64), args.Error(1)
}

func (m *MockArgsRepository) UpdateTaskArgs(ctx context.Context, tx *sql.Tx, id int64, index int, value *float64) error {
	args := m.Called(ctx, tx, id, index, value)
	return args.Error(0)
}

type MockDepsRepository struct {
	mock.Mock
}

func (m *MockDepsRepository) CreateTaskDeps(ctx context.Context, tx *sql.Tx, task *models.Task) error {
	args := m.Called(ctx, tx, task)
	return args.Error(0)
}

func (m *MockDepsRepository) ReadTaskDeps(ctx context.Context, tx *sql.Tx, id int64) ([]int64, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockDepsRepository) UpdateTaskDeps(ctx context.Context, tx *sql.Tx, id int64, deps []int64) error {
	args := m.Called(ctx, tx, id, deps)
	return args.Error(0)
}
