package repositories_expressions

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"regexp"
	"testing"
)

func TestCreateTask_CorrectTask_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	task := &models.Task{
		Expression:   1,
		Operation:    "+",
		Args:         []*float64{float64Ptr(2), float64Ptr(3)},
		Dependencies: []int64{},
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	sqlMock.ExpectQuery(`INSERT INTO tasks`).
		WithArgs(task.Expression, task.Operation).
		WillReturnRows(rows)

	argsRepoMock.On("CreateTaskArgs", mock.Anything, tx, task).Return(nil)
	depsRepoMock.On("CreateTaskDeps", mock.Anything, tx, task).Return(nil)

	id, err, status := repo.CreateTask(context.Background(), tx, task)

	assert.Equal(t, int64(1), id)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	argsRepoMock.AssertExpectations(t)
	depsRepoMock.AssertExpectations(t)
}

func TestCreateTask_CorrectTask_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	task := &models.Task{
		Expression:   1,
		Operation:    "+",
		Args:         []*float64{float64Ptr(2), float64Ptr(3)},
		Dependencies: []int64{},
	}

	sqlMock.ExpectQuery(`INSERT INTO tasks`).
		WithArgs(task.Expression, task.Operation).
		WillReturnError(errors.New("error"))

	argsRepoMock.On("CreateTaskArgs", mock.Anything, tx, task).Return(nil)
	depsRepoMock.On("CreateTaskDeps", mock.Anything, tx, task).Return(nil)

	id, err, status := repo.CreateTask(context.Background(), tx, task)

	assert.Equal(t, int64(0), id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось создать задачу")
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCreateTask_IncorrectArgs_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	task := &models.Task{
		Expression:   1,
		Operation:    "+",
		Args:         []*float64{float64Ptr(2), float64Ptr(3)},
		Dependencies: []int64{},
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	sqlMock.ExpectQuery(`INSERT INTO tasks`).
		WithArgs(task.Expression, task.Operation).
		WillReturnRows(rows)

	argsRepoMock.On("CreateTaskArgs", mock.Anything, tx, task).Return(errors.New("error"))
	depsRepoMock.On("CreateTaskDeps", mock.Anything, tx, task).Return(nil)

	id, err, status := repo.CreateTask(context.Background(), tx, task)

	assert.Equal(t, int64(0), id)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	argsRepoMock.AssertExpectations(t)
}

func TestCreateTask_IncorrectDeps_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	task := &models.Task{
		Expression:   1,
		Operation:    "+",
		Args:         []*float64{float64Ptr(2), float64Ptr(3)},
		Dependencies: []int64{},
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	sqlMock.ExpectQuery(`INSERT INTO tasks`).
		WithArgs(task.Expression, task.Operation).
		WillReturnRows(rows)

	argsRepoMock.On("CreateTaskArgs", mock.Anything, tx, task).Return(nil)
	depsRepoMock.On("CreateTaskDeps", mock.Anything, tx, task).Return(errors.New("error"))

	id, err, status := repo.CreateTask(context.Background(), tx, task)

	assert.Equal(t, int64(0), id)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	depsRepoMock.AssertExpectations(t)
}

func TestCreateTask_CanceledContext_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	id, err, status := repo.CreateTask(ctx, tx, &models.Task{})

	assert.Equal(t, int64(0), id)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	depsRepoMock.AssertExpectations(t)
}

func TestReadTaskByID_CorrectId_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"}).
		AddRow(int64(1), int64(2), "+", float64(3), "completed")
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE id = ?`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(1)).Return([]*float64{float64Ptr(3), float64Ptr(0)}, nil)
	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(1)).Return([]int64{1, 2}, nil)

	task, err, status := repo.ReadTaskByID(context.Background(), tx, int64(1))

	assert.Equal(t, &models.Task{
		ID:           int64(1),
		Expression:   int64(2),
		Operation:    "+",
		Result:       float64Ptr(3),
		Status:       "completed",
		Args:         []*float64{float64Ptr(3), float64Ptr(0)},
		Dependencies: []int64{1, 2},
	}, task)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	argsRepoMock.AssertExpectations(t)
	depsRepoMock.AssertExpectations(t)
}

func TestReadTaskByID_UndefinedId_Error(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE id = ?`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(1)).Return([]*float64{}, nil)
	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(1)).Return([]int64{}, nil)

	task, err, status := repo.ReadTaskByID(context.Background(), tx, int64(1))

	assert.Nil(t, task)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadTaskByID_CorrectId_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE id = ?`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("error"))

	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(1)).Return([]*float64{}, nil)
	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(1)).Return([]int64{}, nil)

	task, err, status := repo.ReadTaskByID(context.Background(), tx, int64(1))

	assert.Nil(t, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось получить задачу")
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadTaskByID_IncorrectArgs_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"}).
		AddRow(int64(1), int64(2), "+", float64(3), "completed")
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE id = ?`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(1)).Return([]*float64{}, errors.New("error"))
	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(1)).Return([]int64{}, nil)

	task, err, status := repo.ReadTaskByID(context.Background(), tx, int64(1))

	assert.Nil(t, task)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	argsRepoMock.AssertExpectations(t)
}

func TestReadTaskByID_IncorrectDeps_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"}).
		AddRow(int64(1), int64(2), "+", float64(3), "completed")
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE id = ?`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(1)).Return([]*float64{}, nil)
	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(1)).Return([]int64{}, errors.New("error"))

	task, err, status := repo.ReadTaskByID(context.Background(), tx, int64(1))

	assert.Nil(t, task)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	depsRepoMock.AssertExpectations(t)
}

func TestReadTaskByID_CanceledContext_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	task, err, status := repo.ReadTaskByID(ctx, tx, int64(1))

	assert.Nil(t, task)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	depsRepoMock.AssertExpectations(t)
}

func TestReadTasksByExpressionID_CorrectId_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)
	expectedTasks := []*models.Task{
		{
			ID:           1,
			Expression:   expressionID,
			Operation:    "+",
			Result:       float64Ptr(3),
			Status:       "completed",
			Args:         []*float64{float64Ptr(1), float64Ptr(2)},
			Dependencies: []int64{2, 3},
		},
		{
			ID:           2,
			Expression:   expressionID,
			Operation:    "*",
			Result:       float64Ptr(6),
			Status:       "completed",
			Args:         []*float64{float64Ptr(3), float64Ptr(2)},
			Dependencies: []int64{},
		},
	}
	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"}).
		AddRow(1, expressionID, "+", expectedTasks[0].Result, "completed").
		AddRow(2, expressionID, "*", expectedTasks[1].Result, "completed")
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE expression_id = ?`).
		WithArgs(int64(2)).
		WillReturnRows(rows)

	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(1)).Return([]int64{2, 3}, nil)
	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(1)).Return(expectedTasks[0].Args, nil)

	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(2)).Return([]int64{}, nil)
	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(2)).Return(expectedTasks[1].Args, nil)

	tasks, err, status := repo.ReadTasksByExpressionID(context.Background(), tx, int64(2))

	if assert.Len(t, tasks, 2) {
		assert.Equal(t, expectedTasks[0], tasks[0])
		assert.Equal(t, expectedTasks[1], tasks[1])
	}
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	argsRepoMock.AssertExpectations(t)
	depsRepoMock.AssertExpectations(t)
}

func TestReadTasksByExpressionID_UndefinedId_Error(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"})
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE expression_id = ?`).
		WithArgs(int64(0)).
		WillReturnRows(rows)

	tasks, err, status := repo.ReadTasksByExpressionID(context.Background(), tx, int64(0))

	assert.Nil(t, tasks)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadTasksByExpressionID_CorrectId_Error(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE expression_id = ?`).
		WithArgs(int64(0)).
		WillReturnError(errors.New("error"))

	tasks, err, status := repo.ReadTasksByExpressionID(context.Background(), tx, int64(0))

	assert.Nil(t, tasks)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadTasksByExpressionID_IncorrectRows_ReadError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)

	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"}).
		AddRow("string", expressionID, "+", float64Ptr(3), "completed").
		AddRow(2, expressionID, "*", float64Ptr(6), "completed")
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE expression_id = ?`).
		WithArgs(expressionID).
		WillReturnRows(rows)

	tasks, err, status := repo.ReadTasksByExpressionID(context.Background(), tx, expressionID)

	assert.Nil(t, tasks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось прочитать задачи")
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadTasksByExpressionID_IncorrectRows_ProcessingError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)

	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"}).
		AddRow(1, expressionID, "+", float64Ptr(3), "completed")
	rows.RowError(0, errors.New("error"))
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE expression_id = ?`).
		WithArgs(expressionID).
		WillReturnRows(rows)

	tasks, err, status := repo.ReadTasksByExpressionID(context.Background(), tx, expressionID)

	assert.Nil(t, tasks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка при обработке строк")
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadTasksByExpressionID_IncorrectArgs_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)

	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"}).
		AddRow(1, expressionID, "+", float64Ptr(3), "completed").
		AddRow(2, expressionID, "*", float64Ptr(6), "completed")
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE expression_id = ?`).
		WithArgs(int64(2)).
		WillReturnRows(rows)

	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(1)).Return([]int64{}, nil)
	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(1)).Return([]*float64{}, nil)

	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(2)).Return([]int64{}, nil)
	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(2)).Return([]*float64{}, errors.New("error"))

	tasks, err, status := repo.ReadTasksByExpressionID(context.Background(), tx, int64(2))

	assert.Nil(t, tasks)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	argsRepoMock.AssertExpectations(t)
	depsRepoMock.AssertExpectations(t)
}

func TestReadTasksByExpressionID_IncorrectDeps_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)

	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"}).
		AddRow(1, expressionID, "+", float64Ptr(3), "completed").
		AddRow(2, expressionID, "*", float64Ptr(6), "completed")
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE expression_id = ?`).
		WithArgs(int64(2)).
		WillReturnRows(rows)

	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(1)).Return([]int64{2, 3}, nil)
	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(1)).Return([]*float64{}, nil)

	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(2)).Return([]int64{}, errors.New("error"))

	tasks, err, status := repo.ReadTasksByExpressionID(context.Background(), tx, int64(2))

	assert.Nil(t, tasks)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	argsRepoMock.AssertExpectations(t)
	depsRepoMock.AssertExpectations(t)
}

func TestReadTasksByExpressionID_CanceledContext_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tasks, err, status := repo.ReadTasksByExpressionID(ctx, tx, int64(2))

	assert.Nil(t, tasks)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadUncompletedTasks_CorrectId_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)
	expectedTasks := []*models.Task{
		{
			ID:           1,
			Expression:   expressionID,
			Operation:    "+",
			Result:       float64Ptr(3),
			Status:       "pending",
			Args:         []*float64{float64Ptr(1), float64Ptr(2)},
			Dependencies: []int64{2, 3},
		},
		{
			ID:           2,
			Expression:   expressionID,
			Operation:    "*",
			Result:       float64Ptr(6),
			Status:       "pending",
			Args:         []*float64{float64Ptr(3), float64Ptr(2)},
			Dependencies: []int64{},
		},
	}
	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"}).
		AddRow(1, expressionID, "+", expectedTasks[0].Result, "pending").
		AddRow(2, expressionID, "*", expectedTasks[1].Result, "pending")
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE status = 'pending'`).
		WillReturnRows(rows)

	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(1)).Return([]int64{2, 3}, nil)
	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(1)).Return(expectedTasks[0].Args, nil)

	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(2)).Return([]int64{}, nil)
	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(2)).Return(expectedTasks[1].Args, nil)

	tasks, err, status := repo.ReadUncompletedTasks(context.Background(), tx)

	if assert.Len(t, tasks, 2) {
		assert.Equal(t, expectedTasks[0], tasks[0])
		assert.Equal(t, expectedTasks[1], tasks[1])
	}
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	argsRepoMock.AssertExpectations(t)
	depsRepoMock.AssertExpectations(t)
}

func TestReadUncompletedTasks_UndefinedId_Error(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"})
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE status = 'pending'`).
		WillReturnRows(rows)

	tasks, err, status := repo.ReadUncompletedTasks(context.Background(), tx)

	assert.Nil(t, tasks)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadUncompletedTasks_CorrectId_Error(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE status = 'pending'`).
		WillReturnError(errors.New("error"))

	tasks, err, status := repo.ReadUncompletedTasks(context.Background(), tx)

	assert.Nil(t, tasks)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadUncompletedTasks_IncorrectRows_ReadError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"}).
		AddRow("string", int64(1), "+", float64Ptr(3), "pending").
		AddRow(2, int64(1), "*", float64Ptr(6), "pending")
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE status = 'pending'`).
		WillReturnRows(rows)

	tasks, err, status := repo.ReadUncompletedTasks(context.Background(), tx)

	assert.Nil(t, tasks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось прочитать задачи")
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadUncompletedTasks_IncorrectRows_ProcessingError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)

	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"}).
		AddRow(1, expressionID, "+", float64Ptr(3), "pending")
	rows.RowError(0, errors.New("error"))
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE status = 'pending'`).
		WillReturnRows(rows)

	tasks, err, status := repo.ReadUncompletedTasks(context.Background(), tx)

	assert.Nil(t, tasks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка при обработке строк")
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadUncompletedTasks_IncorrectArgs_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)

	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"}).
		AddRow(1, expressionID, "+", float64Ptr(3), "pending").
		AddRow(2, expressionID, "*", float64Ptr(6), "pending")
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE status = 'pending'`).
		WillReturnRows(rows)

	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(1)).Return([]int64{}, nil)
	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(1)).Return([]*float64{}, nil)

	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(2)).Return([]int64{}, nil)
	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(2)).Return([]*float64{}, errors.New("error"))

	tasks, err, status := repo.ReadUncompletedTasks(context.Background(), tx)

	assert.Nil(t, tasks)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	argsRepoMock.AssertExpectations(t)
	depsRepoMock.AssertExpectations(t)
}

func TestReadUncompletedTasks_IncorrectDeps_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)

	rows := sqlmock.NewRows([]string{"id", "expression_id", "operation", "result", "status"}).
		AddRow(1, expressionID, "+", float64Ptr(3), "pending").
		AddRow(2, expressionID, "*", float64Ptr(6), "pending")
	sqlMock.ExpectQuery(`SELECT id, expression_id, operation, result, status FROM tasks WHERE status = 'pending'`).
		WillReturnRows(rows)

	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(1)).Return([]int64{2, 3}, nil)
	argsRepoMock.On("ReadTaskArgs", mock.Anything, tx, int64(1)).Return([]*float64{}, nil)

	depsRepoMock.On("ReadTaskDeps", mock.Anything, tx, int64(2)).Return([]int64{}, errors.New("error"))

	tasks, err, status := repo.ReadUncompletedTasks(context.Background(), tx)

	assert.Nil(t, tasks)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	argsRepoMock.AssertExpectations(t)
	depsRepoMock.AssertExpectations(t)
}

func TestReadUncompletedTasks_CanceledContext_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)
	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tasks, err, status := repo.ReadUncompletedTasks(ctx, tx)

	assert.Nil(t, tasks)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateTaskDependencies_CorrectTask_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	task := &models.Task{
		ID:           1,
		Dependencies: []int64{2, 3},
	}

	depsRepoMock.On("UpdateTaskDeps", mock.Anything, tx, task.ID, task.Dependencies).
		Return(nil)

	err, status := repo.UpdateTaskDependencies(context.Background(), tx, task)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	depsRepoMock.AssertExpectations(t)
}

func TestUpdateTaskDependencies_CorrectTask_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	task := &models.Task{
		ID:           1,
		Dependencies: []int64{2, 3},
	}

	depsRepoMock.On("UpdateTaskDeps", mock.Anything, tx, task.ID, task.Dependencies).
		Return(errors.New("error"))

	err, status := repo.UpdateTaskDependencies(context.Background(), tx, task)

	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	depsRepoMock.AssertExpectations(t)
}

func TestUpdateTaskDependencies_CanceledContext_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	depsRepoMock := new(MockDepsRepository)

	repo := NewTasksRepository(db, depsRepoMock, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	task := &models.Task{
		ID:           1,
		Dependencies: []int64{2, 3},
	}

	depsRepoMock.On("UpdateTaskDeps", mock.Anything, tx, task.ID, task.Dependencies).
		Return(context.Canceled)

	err, status := repo.UpdateTaskDependencies(ctx, tx, task)

	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateTaskArguments_CorrectTask_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)

	repo := NewTasksRepository(db, nil, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	value := float64Ptr(1)
	task := &models.Task{
		ID:   1,
		Args: []*float64{float64Ptr(2), float64Ptr(3)},
	}

	argsRepoMock.On("UpdateTaskArgs", mock.Anything, tx, task.ID, 1, value).
		Return(nil)

	err, status := repo.UpdateTaskArguments(context.Background(), tx, task.ID, 1, value)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	argsRepoMock.AssertExpectations(t)
}

func TestUpdateTaskArguments_CorrectTask_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)

	repo := NewTasksRepository(db, nil, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	value := float64Ptr(1)
	task := &models.Task{
		ID:   1,
		Args: []*float64{float64Ptr(2), float64Ptr(3)},
	}

	argsRepoMock.On("UpdateTaskArgs", mock.Anything, tx, task.ID, 1, value).
		Return(errors.New("error"))

	err, status := repo.UpdateTaskArguments(context.Background(), tx, task.ID, 1, value)

	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	argsRepoMock.AssertExpectations(t)
}

func TestUpdateTaskArguments_CanceledContext_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	argsRepoMock := new(MockArgsRepository)

	repo := NewTasksRepository(db, nil, argsRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	value := float64Ptr(1)
	task := &models.Task{
		ID:   1,
		Args: []*float64{float64Ptr(2), float64Ptr(3)},
	}

	argsRepoMock.On("UpdateTaskArgs", mock.Anything, tx, task.ID, 1, value).
		Return(context.Canceled)

	err, status := repo.UpdateTaskArguments(ctx, tx, task.ID, 1, value)

	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateTaskStatus_CorrectStatus_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewTasksRepository(db, nil, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	taskID := int64(1)
	newStatus := "completed"

	sqlMock.ExpectExec(`UPDATE tasks`).
		WithArgs(newStatus, taskID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err, status := repo.UpdateTaskStatus(context.Background(), tx, taskID, newStatus)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateTaskStatus_CorrectStatus_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &TasksRepository{db: db}

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	taskID := int64(1)
	newStatus := "completed"

	sqlMock.ExpectExec(`UPDATE tasks`).
		WithArgs(newStatus, taskID).
		WillReturnError(errors.New("error"))

	err, status := repo.UpdateTaskStatus(context.Background(), tx, taskID, newStatus)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось обновить статус задачи")
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateTaskStatus_CanceledContext_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &TasksRepository{db: db}

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err, status := repo.UpdateTaskStatus(ctx, tx, int64(1), "completed")

	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateTaskExpressionID_CorrectStatus_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewTasksRepository(db, nil, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	taskID := int64(1)
	exprId := int64(2)

	sqlMock.ExpectExec(`UPDATE tasks`).
		WithArgs(exprId, taskID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err, status := repo.UpdateTaskExpressionID(context.Background(), tx, taskID, exprId)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateTaskExpressionID_CorrectStatus_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &TasksRepository{db: db}

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	taskID := int64(1)
	exprId := int64(2)

	sqlMock.ExpectExec(`UPDATE tasks`).
		WithArgs(exprId, taskID).
		WillReturnError(errors.New("error"))

	err, status := repo.UpdateTaskExpressionID(context.Background(), tx, taskID, exprId)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось обновить id задачи выражения")
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateTaskExpressionID_CanceledContext_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &TasksRepository{db: db}

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err, status := repo.UpdateTaskExpressionID(ctx, tx, int64(1), int64(2))

	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateTaskResult_CorrectStatus_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewTasksRepository(db, nil, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	taskID := int64(1)
	result := 2.0

	sqlMock.ExpectExec(`UPDATE tasks`).
		WithArgs(result, taskID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err, status := repo.UpdateTaskResult(context.Background(), tx, result, taskID)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateTaskResult_CorrectStatus_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewTasksRepository(db, nil, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	taskID := int64(1)
	result := 2.0

	sqlMock.ExpectExec(`UPDATE tasks`).
		WithArgs(result, taskID).
		WillReturnError(errors.New("error"))

	err, status := repo.UpdateTaskResult(context.Background(), tx, result, taskID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось обновить результат задачи")
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateTaskResult_CanceledContext_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewTasksRepository(db, nil, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err, status := repo.UpdateTaskResult(ctx, tx, 2.0, int64(1))

	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestDeleteTasks_CorrectId_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewTasksRepository(db, nil, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	taskID := int64(1)

	sqlMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tasks WHERE expression_id = ?`)).
		WithArgs(taskID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err, status := repo.DeleteTasks(context.Background(), tx, taskID)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestDeleteTasks_CorrectId_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewTasksRepository(db, nil, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	taskID := int64(1)
	expectedError := errors.New("database error")

	sqlMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tasks WHERE expression_id = ?`)).
		WithArgs(taskID).
		WillReturnError(expectedError)

	err, status := repo.DeleteTasks(context.Background(), tx, taskID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось удалить задачи")
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestDeleteTasks_CanceledContext_InternalError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewTasksRepository(db, nil, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err, status := repo.DeleteTasks(ctx, tx, int64(1))

	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
