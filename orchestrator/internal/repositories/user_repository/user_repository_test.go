package user_repository_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories/user_repository"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestCreateUser_CorrectUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := user_repository.NewUserRepository(db)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("test_login", "test_pass").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	user, err, code := repo.CreateUser(context.Background(), "test_login", "test_pass")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, int64(1), user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_DuplicateLogin_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := user_repository.NewUserRepository(db)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("existing_login", "test_pass").
		WillReturnError(sqlite3.Error{
			ExtendedCode: sqlite3.ErrConstraintUnique,
		})

	_, err, code := repo.CreateUser(context.Background(), "existing_login", "test_pass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "уже существует")
	assert.Equal(t, http.StatusConflict, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_CorrectUser_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := user_repository.NewUserRepository(db)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("test_login", "test_pass").
		WillReturnError(errors.New("error"))

	_, err, code := repo.CreateUser(context.Background(), "test_login", "test_pass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось создать пользователя")
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_CanceledContext_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := user_repository.NewUserRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err, code := repo.CreateUser(ctx, "test_login", "test_pass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось создать пользователя")
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadUserByLogin_CorrectLogin_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := user_repository.NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"test_id", "test_login", "test_pass"}).AddRow(1, "test_login", "test_pass")
	mock.ExpectQuery(`SELECT id, login, pas FROM users WHERE login = ?`).
		WithArgs("test_login").
		WillReturnRows(rows)

	user, err, code := repo.ReadUserByLogin(context.Background(), "test_login")

	assert.NoError(t, err)
	assert.Equal(t, &models.User{ID: 1, Login: "test_login", Password: "test_pass"}, user)
	assert.Equal(t, http.StatusOK, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadUserByLogin_CorrectLogin_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := user_repository.NewUserRepository(db)

	mock.ExpectQuery(`SELECT id, login, pas FROM users WHERE login = ?`).
		WithArgs("test_login").
		WillReturnError(errors.New("error"))

	_, err, code := repo.ReadUserByLogin(context.Background(), "test_login")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось получить пользователя")
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadUserByLogin_UndefinedLogin_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := user_repository.NewUserRepository(db)

	mock.ExpectQuery(`SELECT id, login, pas FROM users WHERE login = ?`).
		WithArgs("test_login").
		WillReturnError(sql.ErrNoRows)

	_, err, code := repo.ReadUserByLogin(context.Background(), "test_login")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "пользователь не найден")
	assert.Equal(t, http.StatusUnauthorized, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReadUserByLogin_CanceledContext_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := user_repository.NewUserRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err, code := repo.ReadUserByLogin(ctx, "test_login")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось получить пользователя")
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUser_CorrectId_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := user_repository.NewUserRepository(db)

	mock.ExpectExec(`DELETE FROM users WHERE id = ?`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err, code := repo.DeleteUser(context.Background(), int64(1))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUser_IncorrectId_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := user_repository.NewUserRepository(db)

	mock.ExpectExec(`DELETE FROM users WHERE id = ?`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err, code := repo.DeleteUser(context.Background(), int64(1))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUser_CorrectId_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := user_repository.NewUserRepository(db)

	mock.ExpectExec(`DELETE FROM users WHERE id = ?`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("error"))

	err, code := repo.DeleteUser(context.Background(), int64(1))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось удалить пользователя")
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUser_CanceledContext_InternalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := user_repository.NewUserRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err, code := repo.DeleteUser(ctx, int64(1))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось удалить пользователя")
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
