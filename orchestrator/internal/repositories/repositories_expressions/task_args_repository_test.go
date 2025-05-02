package repositories_expressions

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTaskArgs_CorrectArgs_Success(t *testing.T) {
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

	repo := NewTaskArgsRepository(db)

	mock.ExpectExec(`INSERT INTO task_args`).
		WithArgs(int64(1), float64Ptr(2), float64Ptr(3)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateTaskArgs(context.Background(), tx, &models.Task{ID: int64(1), Args: []*float64{float64Ptr(2), float64Ptr(3)}})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTaskArgs_CorrectArgs_InternalError(t *testing.T) {
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

	repo := NewTaskArgsRepository(db)

	mock.ExpectExec(`INSERT INTO task_args`).
		WithArgs(int64(1), float64Ptr(2), float64Ptr(3)).
		WillReturnError(errors.New("error"))

	err = repo.CreateTaskArgs(context.Background(), tx, &models.Task{ID: int64(1), Args: []*float64{float64Ptr(2), float64Ptr(3)}})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось установить аргументы задачи")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTaskArgs_CanceledContext_InternalError(t *testing.T) {
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

	repo := NewTaskArgsRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = repo.CreateTaskArgs(ctx, tx, &models.Task{ID: int64(1), Args: []*float64{float64Ptr(2), float64Ptr(3)}})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось установить аргументы задачи")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadTaskArgs_CorrectId_Success(t *testing.T) {
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

	repo := NewTaskArgsRepository(db)

	rows := sqlmock.NewRows([]string{"first", "second"}).AddRow(float64Ptr(1), float64Ptr(2))
	mock.ExpectQuery(`SELECT first, second FROM task_args WHERE task_id = ?`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	deps, err := repo.ReadTaskArgs(context.Background(), tx, int64(1))

	assert.NoError(t, err)
	assert.Equal(t, []*float64{float64Ptr(1), float64Ptr(2)}, deps)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadTaskArgs_InternalError(t *testing.T) {
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

	repo := NewTaskArgsRepository(db)

	mock.ExpectQuery(`SELECT first, second FROM task_args WHERE task_id = ?`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("error"))

	_, err = repo.ReadTaskArgs(context.Background(), tx, int64(1))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось получить аргументы задачи")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadTaskArgs_CanceledContext_InternalError(t *testing.T) {
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

	repo := NewTaskArgsRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = repo.ReadTaskArgs(ctx, tx, int64(1))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось получить аргументы задачи")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTaskArgs_CorrectArgs_Success(t *testing.T) {
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

	repo := NewTaskArgsRepository(db)

	mock.ExpectExec(`UPDATE task_args`).
		WithArgs(float64Ptr(3), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateTaskArgs(context.Background(), tx, int64(1), 2, float64Ptr(3))

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTaskArgs_CorrectArgs_InternalError(t *testing.T) {
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

	repo := NewTaskArgsRepository(db)

	mock.ExpectExec(`UPDATE task_args`).
		WithArgs(float64Ptr(3), int64(1)).
		WillReturnError(errors.New("error"))

	err = repo.UpdateTaskArgs(context.Background(), tx, int64(1), 2, float64Ptr(3))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось обновить аргументы задачи")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTaskArgs_CanceledContext_InternalError(t *testing.T) {
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

	repo := NewTaskArgsRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = repo.UpdateTaskArgs(ctx, tx, int64(1), 2, float64Ptr(3))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось обновить аргументы задачи")
	assert.NoError(t, mock.ExpectationsWereMet())
}
