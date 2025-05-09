package session_repository_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories/session_repository"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestCreateSession_CorrectSession_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	repo := session_repository.NewSessionRepository(db)

	mock.ExpectExec(`INSERT INTO sessions`).
		WithArgs("test_id", 1, int64(1234567890)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err, code := repo.CreateSession(context.Background(), tx, "test_id", int64(1), int64(1234567890))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateSession_CorrectSession_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	repo := session_repository.NewSessionRepository(db)

	mock.ExpectExec(`INSERT INTO sessions`).
		WithArgs("test_id", 1, int64(1234567890)).
		WillReturnError(errors.New("error"))

	err, code := repo.CreateSession(context.Background(), tx, "test_id", int64(1), int64(1234567890))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось создать сессию")
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateSession_CanceledContext_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	repo := session_repository.NewSessionRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err, code := repo.CreateSession(ctx, tx, "test_id", int64(1), int64(1234567890))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось создать сессию")
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadSession_CorrectJti_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := session_repository.NewSessionRepository(db)

	rows := sqlmock.NewRows([]string{"id", "user_id", "expires"}).AddRow("test_jti", int64(1), int64(1234567890))
	mock.ExpectQuery(`SELECT id, user_id, expires FROM sessions WHERE id = ?`).
		WithArgs("test_jti").
		WillReturnRows(rows)

	session, err, code := repo.ReadSession(context.Background(), "test_jti")

	assert.NoError(t, err)
	assert.Equal(t, &models.Session{ID: "test_jti", UserID: int64(1), Expires: int64(1234567890)}, session)
	assert.Equal(t, http.StatusOK, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadSession_UndefinedJti_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := session_repository.NewSessionRepository(db)

	mock.ExpectQuery(`SELECT id, user_id, expires FROM sessions WHERE id = ?`).
		WithArgs("test_jti").
		WillReturnError(sql.ErrNoRows)

	_, err, code := repo.ReadSession(context.Background(), "test_jti")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadSession_CorrectJti_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := session_repository.NewSessionRepository(db)

	mock.ExpectQuery(`SELECT id, user_id, expires FROM sessions WHERE id = ?`).
		WithArgs("test_jti").
		WillReturnError(errors.New("error"))

	_, err, code := repo.ReadSession(context.Background(), "test_jti")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось получить сессию")
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadSession_CanceledContext_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := session_repository.NewSessionRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err, code := repo.ReadSession(ctx, "test_jti")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось получить сессию")
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteSession_CorrectId_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := session_repository.NewSessionRepository(db)

	mock.ExpectExec(`DELETE FROM sessions WHERE id = ?`).
		WithArgs("test_id").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err, code := repo.DeleteSession(context.Background(), "test_id")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteSession_IncorrectId_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := session_repository.NewSessionRepository(db)

	mock.ExpectExec(`DELETE FROM sessions WHERE id = ?`).
		WithArgs("test_id").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err, code := repo.DeleteSession(context.Background(), "test_id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "сессия для пользователя не найдена")
	assert.Equal(t, http.StatusNotFound, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteSession_CorrectId_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := session_repository.NewSessionRepository(db)

	mock.ExpectExec(`DELETE FROM sessions WHERE id = ?`).
		WithArgs("test_id").
		WillReturnError(errors.New("error"))

	err, code := repo.DeleteSession(context.Background(), "test_id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось удалить сессию")
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteSession_CanceledContext_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := session_repository.NewSessionRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err, code := repo.DeleteSession(ctx, "test_id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось удалить сессию")
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
