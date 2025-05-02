package repositories_users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/OinkiePie/calc_3/pkg/models"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
)

// SessionRepository предоставляет методы для работы с сессиями в базе данных.
type SessionRepository struct {
	db *sql.DB // Подключение к базе данных
}

// NewSessionRepository создает новый экземпляр SessionRepository.
//
// Args:
//
//	db: *sql.DB - Подключение к базе данных.
//
// Returns:
//
//	*SessionRepository - Новый экземпляр репозитория пользователей.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession создает новую сессию в базе данных.
//
// Args:
//
//	ctx: context.Context - Контекст для контроля выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных
//	jti: string - Идентификатор сессии.
//	sub: int64 - Идентификатор пользователя.
//	exp: int64 - Время истечения сессии.
//
// Returns:
//
//	error - Ошибка выполнения операции.
//	int - HTTP-статус код результата операции:
//		- 201 Created при успешном создании
//		- 500 Internal Server Error при других ошибках
func (r *SessionRepository) CreateSession(ctx context.Context, tx *sql.Tx, jti string, sub, exp int64) (error, int) {
	query := `
	INSERT INTO sessions
		(id, user_id, expires)
	VALUES
		(?, ?, ?)`

	if _, err := tx.ExecContext(ctx, query, jti, sub, exp); err != nil {
		return fmt.Errorf("не удалось создать сессию: %w", err), http.StatusInternalServerError
	}

	return nil, http.StatusCreated
}

// ReadSession получает сессию из базы данных.
//
// Args:
//
//	ctx: context.Context - Контекст для контроля выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных.
//	jti: string - Идентификатор сессии.
//
// Returns:
//
//	*models.Session - Указатель на сессию.
//	error - Ошибка выполнения операции.
//	int - HTTP-статус код результата операции:
//		- 200 OK при успешном получении
//		- 404 Not Found если сессия не найдена
//		- 500 Internal Server Error при других ошибках
func (r *SessionRepository) ReadSession(ctx context.Context, jti string) (*models.Session, error, int) {
	var session models.Session
	query := `
	SELECT
		id, user_id, expires
	FROM
		sessions
	WHERE
		id = ?`

	err := r.db.QueryRowContext(ctx, query, jti).Scan(&session.ID, &session.UserID, &session.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, http.StatusNotFound
		}
		return nil, fmt.Errorf("не удалось получить сессию: %w", err), http.StatusInternalServerError
	}
	return &session, nil, http.StatusOK
}

// DeleteSession удаляет сессию из базы данных.
//
// Args:
//
//	ctx: context.Context - Контекст для контроля выполнения запроса.
//	jti: string - Идентификатор сессии.
//
// Returns:
//
//	*models.Session - Указатель на сессию.
//	error - Ошибка выполнения операции.
//	int - HTTP-статус код результата операции:
//		- 200 OK при успешном получении
//		- 500 Internal Server Error при ошибках
func (r *SessionRepository) DeleteSession(ctx context.Context, jti string) (error, int) {
	query := `DELETE FROM sessions WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, jti)
	if err != nil {
		return fmt.Errorf("не удалось удалить сессию: %w", err), http.StatusInternalServerError
	}
	return nil, http.StatusOK
}
