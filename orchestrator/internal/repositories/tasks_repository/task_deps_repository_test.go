package tasks_repository_test

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories/tasks_repository"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTaskDeps_CorrectDeps_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	repo := tasks_repository.NewTaskDepsRepository(db)

	mock.ExpectExec(`INSERT INTO task_deps`).
		WithArgs(int64(1), int64(2), int64(3)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateTaskDeps(context.Background(), tx, &models.Task{ID: 1, Dependencies: []int64{2, 3}})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTaskDeps_CorrectDeps_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	repo := tasks_repository.NewTaskDepsRepository(db)

	mock.ExpectExec(`INSERT INTO task_deps`).
		WithArgs(int64(1), int64(2), int64(3)).
		WillReturnError(errors.New("error"))

	err = repo.CreateTaskDeps(context.Background(), tx, &models.Task{ID: 1, Dependencies: []int64{2, 3}})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось установить зависимости задачи")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTaskDeps_CanceledContext_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	repo := tasks_repository.NewTaskDepsRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = repo.CreateTaskDeps(ctx, tx, &models.Task{ID: 1, Dependencies: []int64{2, 3}})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось установить зависимости задачи")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadTaskDeps_CorrectId_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	repo := tasks_repository.NewTaskDepsRepository(db)

	rows := sqlmock.NewRows([]string{"first", "second"}).AddRow(int64(1), int64(2))
	mock.ExpectQuery(`SELECT first, second FROM task_deps WHERE task_id = ?`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	deps, err := repo.ReadTaskDeps(context.Background(), tx, int64(1))

	assert.NoError(t, err)
	assert.Equal(t, []int64{1, 2}, deps)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadTaskDeps_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	repo := tasks_repository.NewTaskDepsRepository(db)

	mock.ExpectQuery(`SELECT first, second FROM task_deps WHERE task_id = ?`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("error"))

	_, err = repo.ReadTaskDeps(context.Background(), tx, int64(1))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось получить зависимости задачи")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadTaskDeps_CanceledContext_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	repo := tasks_repository.NewTaskDepsRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = repo.ReadTaskDeps(ctx, tx, int64(1))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось получить зависимости задачи")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTaskDeps_CorrectDeps_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	repo := tasks_repository.NewTaskDepsRepository(db)

	mock.ExpectExec(`UPDATE task_deps`).
		WithArgs(int64(2), int64(3), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateTaskDeps(context.Background(), tx, 1, []int64{2, 3})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTaskDeps_CorrectDeps_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	repo := tasks_repository.NewTaskDepsRepository(db)

	mock.ExpectExec(`UPDATE task_deps`).
		WithArgs(int64(2), int64(3), int64(1)).
		WillReturnError(errors.New("error"))

	err = repo.UpdateTaskDeps(context.Background(), tx, 1, []int64{2, 3})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось обновить зависимости задачи")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTaskDeps_CanceledContext_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	repo := tasks_repository.NewTaskDepsRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = repo.UpdateTaskDeps(ctx, tx, 1, []int64{2, 3})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось обновить зависимости задачи")
	assert.NoError(t, mock.ExpectationsWereMet())
}
