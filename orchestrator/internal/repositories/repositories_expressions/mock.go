package repositories_expressions

import (
	"context"
	"database/sql"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/stretchr/testify/mock"
)

func float64Ptr(i float64) *float64 {
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
