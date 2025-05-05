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
