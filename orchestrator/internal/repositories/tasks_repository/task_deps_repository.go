package tasks_repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/OinkiePie/calc_3/pkg/models"
	_ "github.com/mattn/go-sqlite3"
)

// TaskDepsRepository предоставляет методы для работы с зависимостями задач в базе данных.
type TaskDepsRepository struct {
	db *sql.DB // Подключение к базе данных
}

// NewTaskDepsRepository создает новый экземпляр репозитория зависимостей задач.
//
// Args:
//
//	db: *sql.DB - Подключение к базе данных.
//
// Returns:
//
//	*TaskDepsRepository - Новый экземпляр репозитория.
func NewTaskDepsRepository(db *sql.DB) *TaskDepsRepository {
	return &TaskDepsRepository{db: db}
}

// CreateTaskDeps создает записи зависимостей задачи в базе данных.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных.
//	task: *models.Task - Задача, содержащая зависимости для сохранения.
//
// Returns:
//
//	error - Ошибка выполнения операции.
func (r *TaskDepsRepository) CreateTaskDeps(ctx context.Context, tx *sql.Tx, task *models.Task) error {
	query := `
	INSERT INTO task_deps
	    (task_id, first, second)
	VALUES
	    (?, ?, ?)`

	_, err := tx.ExecContext(ctx, query, task.ID, task.Dependencies[0], task.Dependencies[1])
	if err != nil {
		return fmt.Errorf("не удалось установить зависимости задачи: %w", err)
	}
	return nil
}

// ReadTaskDeps получает зависимости задачи из базы данных.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных.
//	id: int64 - Идентификатор задачи.
//
// Returns:
//
//	[]int64 - Список идентификаторов зависимых задач.
//	error - Ошибка выполнения операции.
func (r *TaskDepsRepository) ReadTaskDeps(ctx context.Context, tx *sql.Tx, id int64) ([]int64, error) {
	deps := make([]int64, 2)
	query := `
	SELECT
	    first, second
	FROM
	    task_deps
	WHERE
	    task_id = ?`

	if err := tx.QueryRowContext(ctx, query, id).Scan(&deps[0], &deps[1]); err != nil {
		return []int64{}, fmt.Errorf("не удалось получить зависимости задачи: %w", err)
	}
	return deps, nil
}

// UpdateTaskDeps обновляет зависимости задачи в базе данных.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных.
//	id: int64 - Идентификатор задачи.
//	deps: []int64 - Новые зависимости задачи.
//
// Returns:
//
//	error - Ошибка выполнения операции.
func (r *TaskDepsRepository) UpdateTaskDeps(ctx context.Context, tx *sql.Tx, id int64, deps []int64) error {
	query := `
    UPDATE
    	task_deps
    SET
        first = ?,
        second = ?
	WHERE
	    task_id = ?`

	_, err := tx.ExecContext(ctx, query, deps[0], deps[1], id)
	if err != nil {
		return fmt.Errorf("не удалось обновить зависимости задачи: %w", err)
	}
	return nil
}
