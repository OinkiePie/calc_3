package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/OinkiePie/calc_3/pkg/models"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
)

// ExpressionsRepository предоставляет методы для работы с выражениями в базе данных.
type ExpressionsRepository struct {
	db       *sql.DB          // Подключение к базе данных
	taskRepo *TasksRepository // Репозиторий задач
}

// NewExpressionsRepository создает новый экземпляр репозитория выражений.
//
// Args:
//
//	db: *sql.DB - Подключение к базе данных
//	tr: *TasksRepository - Репозиторий задач
//
// Returns:
//
//	*ExpressionsRepository - Новый экземпляр репозитория
func NewExpressionsRepository(db *sql.DB, tr *TasksRepository) *ExpressionsRepository {
	return &ExpressionsRepository{db: db, taskRepo: tr}
}

// CreateExpression создает новое выражение и связанные с ним задачи.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса
//	tx: *sql.Tx - Транзакция базы данных
//	expr: *models.Expression - Выражение для создания
//
// Returns:
//
//	int64 - ID созданного выражения
//	error - Ошибка выполнения операции
//	int - HTTP статус код:
//	    - 200 OK при успешном создании
//	    - 500 Internal Server Error при ошибках
func (r *ExpressionsRepository) CreateExpression(ctx context.Context, tx *sql.Tx, expr *models.Expression) (int64, error, int) {

	query := `
	INSERT INTO expressions 
    	(user_id, expression_string) 
    VALUES
	       (?, ?)
    RETURNING
    	id`

	var expressionID int64
	err := tx.QueryRowContext(ctx,
		query,
		expr.UserID,
		expr.ExpressionString,
	).Scan(&expressionID)

	if err != nil {
		return 0, fmt.Errorf("не удалось вставить выражение: %w", err), http.StatusInternalServerError
	}

	for i := range expr.Tasks {
		task := expr.Tasks[i]
		task.Expression = expressionID
		taskID, err, code := r.taskRepo.CreateTask(ctx, tx, task)
		if err != nil {
			return 0, err, code
		}
		task.ID = taskID
	}

	for i := range expr.Tasks {
		task := expr.Tasks[i]
		task.Dependencies = make([]int64, len(task.DependencyIndexes))
		for j, depIndex := range task.DependencyIndexes {
			task.Dependencies[j] = 0
			if depIndex > 0 && depIndex < len(expr.Tasks) {
				task.Dependencies[j] = expr.Tasks[depIndex].ID - 1
			} else {
				task.Dependencies[j] = -1
			}
		}
		if err, code := r.taskRepo.UpdateTaskDependencies(ctx, tx, task); err != nil {
			return 0, err, code
		}
	}

	return expressionID, nil, http.StatusOK
}

// ReadExpressionByID получает выражение по его ID вместе с задачами.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса
//	tx: *sql.Tx - Транзакция базы данных
//	id: int64 - ID выражения
//
// Returns:
//
//	*models.Expression - Найденное выражение
//	error - Ошибка выполнения операции
//	int - HTTP статус код:
//	    - 200 OK при успешном получении
//	    - 404 Not Found если выражение не найдено
//	    - 500 Internal Server Error при ошибках
func (r *ExpressionsRepository) ReadExpressionByID(ctx context.Context, tx *sql.Tx, id int64) (*models.Expression, error, int) {
	var expr models.Expression
	query := `
		SELECT
		    id, status, result, expression_string,
		    error, user_id
		FROM
		    expressions
		WHERE
		    id = ?
	`
	err := tx.QueryRowContext(ctx, query, id).Scan(
		&expr.ID,
		&expr.Status,
		&expr.Result,
		&expr.ExpressionString,
		&expr.Error,
		&expr.UserID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("выражение не найдено"), http.StatusNotFound
		}
		return nil, fmt.Errorf("не удалось получать выражение: %w", err), http.StatusInternalServerError
	}

	tasks, err, code := r.taskRepo.ReadTasksByExpressionID(ctx, tx, expr.ID)
	if err != nil {
		return nil, err, code
	}
	expr.Tasks = tasks

	return &expr, nil, http.StatusOK
}

// ReadExpressionsByUserID получает все выражения пользователя.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса
//	tx: *sql.Tx - Транзакция базы данных
//	userID: int64 - ID пользователя
//
// Returns:
//
//	[]*models.Expression - Список выражений пользователя
//	error - Ошибка выполнения операции
//	int - HTTP статус код:
//	    - 200 OK при успешном получении
//	    - 404 Not Found если выражения не найдены
//	    - 500 Internal Server Error при ошибках
func (r *ExpressionsRepository) ReadExpressionsByUserID(ctx context.Context, tx *sql.Tx, userID int64) ([]*models.Expression, error, int) {
	var expressions []*models.Expression
	query := `
		SELECT
		    id, status, result, expression_string,
		    error, user_id
		FROM
		    expressions
		WHERE
		    user_id = ?
	`

	rows, err := tx.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить выражения: %w", err), http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		expr := &models.Expression{}
		err := rows.Scan(
			&expr.ID,
			&expr.Status,
			&expr.Result,
			&expr.ExpressionString,
			&expr.Error,
			&expr.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("не удалось прочитать выражение: %w", err), http.StatusInternalServerError
		}

		tasks, err, code := r.taskRepo.ReadTasksByExpressionID(ctx, tx, expr.ID)
		if err != nil {
			return nil, err, code
		}
		expr.Tasks = tasks
		expressions = append(expressions, expr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке строк: %w", err), http.StatusInternalServerError
	}

	if len(expressions) == 0 {
		return nil, fmt.Errorf("выражения пользователя №%d не найдены", userID), http.StatusNotFound
	}

	return expressions, nil, http.StatusOK
}

// ReadExpressionTasks получает все задачи, связанные с указанным выражением.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса
//	tx: *sql.Tx - Транзакция базы данных (не используется, создается новая)
//	id: int64 - ID выражения
//
// Returns:
//
//	[]*models.Task - Список задач выражения
//	error - Ошибка выполнения операции
//	int - HTTP статус код:
//	    - 200 OK при успешном получении
//	    - 500 Internal Server Error при ошибках
func (r *ExpressionsRepository) ReadExpressionTasks(ctx context.Context, tx *sql.Tx, id int64) ([]*models.Task, error, int) {
	expressions, err, code := r.taskRepo.ReadTasksByExpressionID(ctx, tx, id)
	if err != nil {
		return nil, err, code
	}
	return expressions, nil, http.StatusOK
}

// UpdateExpressionStatus обновляет статус выражения.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса
//	tx: *sql.Tx - Транзакция базы данных
//	id: int64 - ID выражения
//	status: string - Новый статус
//
// Returns:
//
//	error - Ошибка выполнения операции
//	int - HTTP статус код:
//	    - 200 OK при успешном обновлении
//	    - 500 Internal Server Error при ошибках
func (r *ExpressionsRepository) UpdateExpressionStatus(ctx context.Context, tx *sql.Tx, id int64, status string) (error, int) {
	query := `
		UPDATE
		    expressions
		SET
		    status = ?
		WHERE
		    id = ?`

	_, err := tx.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("не удалось обновить статус выражения: %w", err), http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

// UpdateExpressionError обновляет сообщение об ошибке выражения.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса
//	tx: *sql.Tx - Транзакция базы данных
//	id: int64 - ID выражения
//	errContent: string - Текст ошибки
//
// Returns:
//
//	error - Ошибка выполнения операции
//	int - HTTP статус код:
//	    - 200 OK при успешном обновлении
//	    - 500 Internal Server Error при ошибках
func (r *ExpressionsRepository) UpdateExpressionError(ctx context.Context, tx *sql.Tx, id int64, errContent string) (error, int) {
	query := `
		UPDATE
		    expressions
		SET
		    error = ?
		WHERE
		    id = ?`

	_, err := tx.ExecContext(ctx, query, errContent, id)
	if err != nil {
		return fmt.Errorf("не удалось обновить ошибку выражения: %w", err), http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

// UpdateExpressionResult обновляет результат вычисления выражения.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса
//	tx: *sql.Tx - Транзакция базы данных
//	id: int64 - ID выражения
//	result: float64 - Результат вычисления
//
// Returns:
//
//	error - Ошибка выполнения операции
//	int - HTTP статус код:
//	    - 200 OK при успешном обновлении
//	    - 500 Internal Server Error при ошибках
func (r *ExpressionsRepository) UpdateExpressionResult(ctx context.Context, tx *sql.Tx, id int64, result float64) (error, int) {
	query := `
		UPDATE
		    expressions
		SET
		    result = ?
		WHERE
		    id = ?`

	_, err := tx.ExecContext(ctx, query, result, id)
	if err != nil {
		return fmt.Errorf("не удалось обновить результат выражения: %w", err), http.StatusInternalServerError
	}
	return nil, http.StatusOK
}
