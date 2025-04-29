package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/OinkiePie/calc_3/pkg/models"
	_ "github.com/mattn/go-sqlite3"
)

// TaskArgsRepository предоставляет методы для работы с аргументами задач в базе данных.
type TaskArgsRepository struct {
	db *sql.DB // Подключение к базе данных
}

// NewTaskArgsRepository создает новый экземпляр репозитория аргументов задач.
//
// Args:
//
//	db: *sql.DB - Подключение к базе данных
//
// Returns:
//
//	*TaskArgsRepository - Новый экземпляр репозитория
func NewTaskArgsRepository(db *sql.DB) *TaskArgsRepository {
	return &TaskArgsRepository{db: db}
}

// CreateTaskArgs создает запись аргументов задачи в базе данных.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса
//	tx: *sql.Tx - Транзакция базы данных
//	task: *models.Task - Задача, содержащая аргументы для сохранения
//
// Returns:
//
//	error - Ошибка выполнения операции:
//		- nil при успешном создании
//		- "не удалось установить аргументы задачи" при ошибках
func (r *TaskArgsRepository) CreateTaskArgs(ctx context.Context, tx *sql.Tx, task *models.Task) error {
	query := `
	INSERT INTO task_args
	    (task_id, first, second)
	VALUES
	    (?, ?, ?)`

	_, err := tx.ExecContext(ctx, query, task.ID, task.Args[0], task.Args[1])
	if err != nil {
		return fmt.Errorf("не удалось установить аргументы задачи: %w", err)
	}
	return nil
}

// ReadTaskArgs получает аргументы задачи из базы данных.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса
//	tx: *sql.Tx - Транзакция базы данных
//	id: int64 - Идентификатор задачи
//
// Returns:
//
//	[]*float64 - Срез из двух элементов: [первый аргумент, второй аргумент]
//	error - Ошибка выполнения операции
func (r *TaskArgsRepository) ReadTaskArgs(ctx context.Context, tx *sql.Tx, id int64) ([]*float64, error) {
	args := make([]*float64, 2)
	query := `
	SELECT
	    first, second
	FROM
	    task_args
	WHERE
	    task_id = ?`

	if err := tx.QueryRowContext(ctx, query, id).Scan(&args[0], &args[1]); err != nil {
		return []*float64{}, fmt.Errorf("не удалось получить аргументы задачи: %w", err)
	}
	return args, nil
}

// UpdateTaskArgs обновляет один из аргументов задачи в базе данных.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса
//	tx: *sql.Tx - Транзакция базы данных
//	id: int64 - Идентификатор задачи
//	index: int - Индекс аргумента (0 - первый, 1 - второй)
//	value: *float64 - Новое значение аргумента
//
// Returns:
//
//	error - Ошибка выполнения операции
func (r *TaskArgsRepository) UpdateTaskArgs(ctx context.Context, tx *sql.Tx, id int64, index int, value *float64) error {
	position := map[int]string{
		0: "first",
		1: "second",
	}
	query := fmt.Sprintf(`
		UPDATE
		    task_args
		SET
		    %s = ?
		WHERE
		    task_id = ?`, position[index])

	_, err := tx.ExecContext(ctx, query, value, id)
	if err != nil {
		return fmt.Errorf("не удалось обновить аргументы задачи: %w", err)
	}
	return nil
}
