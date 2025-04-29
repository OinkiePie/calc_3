package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
)

// UserRepository предоставляет методы для работы с пользователями в базе данных.
type UserRepository struct {
	db *sql.DB // Подключение к базе данных
}

// NewUserRepository создает новый экземпляр UserRepository.
//
// Args:
//
//	db: *sql.DB - Подключение к базе данных
//
// Returns:
//
//	*UserRepository - Новый экземпляр репозитория пользователей
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser создает нового пользователя в базе данных с указанными учетными данными.
//
// Args:
//
//	ctx: context.Context - Контекст для контроля выполнения запроса.
//	login: string - Логин пользователя (должен быть уникальным).
//	pas: string - Пароль пользователя в незашифрованном виде.
//
// Returns:
//
//	*models.User - Созданный объект пользователя с заполненным ID.
//	error - Ошибка при нарушении уникальности логина или проблемах с БД.
//	int - HTTP-статус код результата операции:
//		- 201 Created при успешном создании
//		- 409 Conflict при дубликате логина
//		- 500 Internal Server Error при других ошибках
func (r *UserRepository) CreateUser(ctx context.Context, login, pas string) (*models.User, error, int) {
	var id int64
	query := `
	INSERT INTO users
	    (login, pas)
	VALUES
	    (?, ?)
	RETURNING
		id`

	if err := r.db.QueryRowContext(ctx, query, login, pas).Scan(&id); err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return nil, fmt.Errorf("логин %s уже существует", login), http.StatusConflict
			}
		}
		return nil, fmt.Errorf("не удалось создать пользователя: %w", err), http.StatusInternalServerError
	}

	user := models.User{
		ID:       id,
		Login:    login,
		Password: pas,
	}
	return &user, nil, http.StatusCreated
}

// ReadUserByLogin получает пользователя из базы данных по его логину.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса
//	login: string - Логин пользователя для поиска
//
// Returns:
//
//	*models.User - Найденный пользователь
//	error - Ошибка выполнения запроса
//	int - HTTP-статус код:
//		- 200 OK при успешном поиске
//		- 401 Unauthorized если пользователь не найден
//		- 500 Internal Server Error при ошибках
func (r *UserRepository) ReadUserByLogin(ctx context.Context, login string) (*models.User, error, int) {
	var user models.User
	query := `
	SELECT
	    id, login, pas
	FROM
	    users
	WHERE
	    login = ?`

	err := r.db.QueryRowContext(ctx, query, login).Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("пользователь не найден"), http.StatusUnauthorized
		}
		return nil, fmt.Errorf("не удалось поулчить пользователя: %w", err), http.StatusInternalServerError
	}
	return &user, nil, http.StatusOK
}

// DeleteUser удаляет пользователя из базы данных по указанному идентификатору.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса
//	id: int64 - Идентификатор пользователя для удаления
//
// Returns:
//
//	error - Ошибка выполнения операции
//	int - HTTP-статус код:
//		- 200 OK при успешном удалении
//		- 500 Internal Server Error при ошибках
func (r *UserRepository) DeleteUser(ctx context.Context, id int64) (error, int) {
	query := `DELETE FROM users WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("не удалось удалить пользователя: %w", err), http.StatusInternalServerError
	}
	return nil, http.StatusOK
}
