package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/OinkiePie/calc_3/pkg/models"
	"net/http"
)

type SessionRepository struct {
	db *sql.DB // Подключение к базе данных
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

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
		return nil, fmt.Errorf("не удалось сессию: %w", err), http.StatusInternalServerError
	}
	return &session, nil, http.StatusOK
}

func (r *SessionRepository) DeleteSession(ctx context.Context, jti string) (error, int) {
	query := `DELETE FROM sessions WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, jti)
	if err != nil {
		return fmt.Errorf("не удалось удалить сессию: %w", err), http.StatusInternalServerError
	}
	return nil, http.StatusOK
}
