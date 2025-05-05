package tasks_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories"
	"github.com/OinkiePie/calc_3/pkg/models"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
)

// TasksRepository предоставляет методы для работы с задачами в базе данных.
type TasksRepository struct {
	db       *sql.DB                                   // Подключение к базе данных
	depsRepo repositories.TasksDepsRepositoryInterface // Репозиторий зависимостей задач
	argsRepo repositories.TasksArgsRepositoryInterface // Репозиторий аргументов задач
}

// NewTasksRepository создает новый экземпляр репозитория задач.
//
// Args:
//
//	db: *sql.DB - Подключение к базе данных.
//	dr: *TaskDepsRepository - Репозиторий зависимостей задач.
//	ar: *TaskArgsRepository - Репозиторий аргументов задач.
//
// Returns:
//
//	*TasksRepository - Новый экземпляр репозитория задач
func NewTasksRepository(db *sql.DB, dr repositories.TasksDepsRepositoryInterface, ar repositories.TasksArgsRepositoryInterface) *TasksRepository {
	return &TasksRepository{db: db, depsRepo: dr, argsRepo: ar}
}

// CreateTask создает новую задачу в базе данных вместе с ее аргументами и зависимостями.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных.
//	task: *models.Task - Задача для создания.
//
// Returns:
//
//	int64 - ID созданной задачи.
//	error - Ошибка выполнения операции.
//	int - HTTP статус код:
//	    - 201 StatusCreated при успешном создании
//	    - 500 Internal Server Error при ошибках
func (r *TasksRepository) CreateTask(ctx context.Context, tx *sql.Tx, task *models.Task) (int64, error, int) {
	query := `
	INSERT INTO tasks 
    	(expression_id, operation) 
    VALUES
        (?, ?)
    RETURNING
    	id`

	if err := tx.QueryRowContext(ctx, query, task.Expression, task.Operation).Scan(&task.ID); err != nil {
		return 0, fmt.Errorf("не удалось создать задачу: %w", err), http.StatusInternalServerError
	}

	if err := r.argsRepo.CreateTaskArgs(ctx, tx, task); err != nil {
		return 0, err, http.StatusInternalServerError
	}

	if err := r.depsRepo.CreateTaskDeps(ctx, tx, task); err != nil {
		return 0, err, http.StatusInternalServerError
	}

	return task.ID, nil, http.StatusCreated
}

// ReadTaskByID получает задачу из базы данных по ее ID вместе с аргументами и зависимостями.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных.
//	id: int64 - ID задачи.
//
// Returns:
//
//	*models.Task - Найденная задача.
//	error - Ошибка выполнения операции.
//	int - HTTP статус код:
//	    - 200 OK при успешном получении
//	    - 404 Not Found если задача не найдена
//	    - 500 Internal Server Error при ошибках ДБ
func (r *TasksRepository) ReadTaskByID(ctx context.Context, tx *sql.Tx, id int64) (*models.Task, error, int) {
	var task models.Task
	query := `
	SELECT
	    id, expression_id, operation,
	    result, status
	FROM
	    tasks
	WHERE
	    id = ?`
	if err := tx.QueryRowContext(ctx, query, id).Scan(&task.ID, &task.Expression, &task.Operation, &task.Result, &task.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, http.StatusNotFound
		}
		return nil, fmt.Errorf("не удалось получить задачу: %w", err), http.StatusInternalServerError
	}

	deps, err := r.depsRepo.ReadTaskDeps(ctx, tx, id)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	task.Dependencies = deps

	args, err := r.argsRepo.ReadTaskArgs(ctx, tx, id)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	task.Args = args

	return &task, nil, http.StatusOK
}

// ReadTasksByExpressionID получает все задачи, связанные с указанным выражением.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных.
//	expressionID: int64 - ID выражения.
//
// Returns:
//
//	[]*models.Task - Список найденных задач.
//	error - Ошибка выполнения операции.
//	int - HTTP статус код:
//	    - 200 OK при успешном получении
//	    - 404 Not Found если задачи не найдены
//	    - 500 Internal Server Error при ошибках
func (r *TasksRepository) ReadTasksByExpressionID(ctx context.Context, tx *sql.Tx, expressionID int64) ([]*models.Task, error, int) {
	var tasks []*models.Task

	query := `
	SELECT
	    id, expression_id, operation,
	    result, status
	FROM
	    tasks
	WHERE
	    expression_id = ?`

	rows, err := tx.QueryContext(ctx, query, expressionID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить задачи: %w", err), http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.Expression, &task.Operation, &task.Result, &task.Status); err != nil {
			return nil, fmt.Errorf("не удалось прочитать задачи: %w", err), http.StatusInternalServerError
		}

		deps, err := r.depsRepo.ReadTaskDeps(ctx, tx, task.ID)
		if err != nil {
			return nil, err, http.StatusInternalServerError
		}
		task.Dependencies = deps

		args, err := r.argsRepo.ReadTaskArgs(ctx, tx, task.ID)
		if err != nil {
			return nil, err, http.StatusInternalServerError
		}
		task.Args = args

		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке строк: %w", err), http.StatusInternalServerError
	}

	if len(tasks) == 0 {
		return nil, nil, http.StatusNotFound
	}

	return tasks, nil, http.StatusOK
}

// ReadUncompletedTasks получает все невыполненные задачи (со статусом 'pending').
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных.
//
// Returns:
//
//	[]*models.Task - Список невыполненных задач.
//	error - Ошибка выполнения операции.
//	int - HTTP статус код:
//	    - 200 OK при успешном получении
//	    - 404 Not Found если задачи не найдены
//	    - 500 Internal Server Error при ошибках
func (r *TasksRepository) ReadUncompletedTasks(ctx context.Context, tx *sql.Tx) ([]*models.Task, error, int) {
	var tasks []*models.Task

	query := `
	SELECT
	    id, expression_id, operation,
	    result, status
	FROM
	    tasks
	WHERE
	    status = 'pending'`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить задачи: %w", err), http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.Expression, &task.Operation, &task.Result, &task.Status); err != nil {
			return nil, fmt.Errorf("не удалось прочитать задачи: %w", err), http.StatusInternalServerError
		}

		deps, err := r.depsRepo.ReadTaskDeps(ctx, tx, task.ID)
		if err != nil {
			return nil, err, http.StatusInternalServerError
		}
		task.Dependencies = deps

		args, err := r.argsRepo.ReadTaskArgs(ctx, tx, task.ID)
		if err != nil {
			return nil, err, http.StatusInternalServerError
		}
		task.Args = args

		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке строк: %w", err), http.StatusInternalServerError
	}

	if len(tasks) == 0 {
		return nil, nil, http.StatusNotFound
	}

	return tasks, nil, http.StatusOK
}

// UpdateTaskDependencies обновляет зависимости задачи.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных.
//	task: *models.Task - Задача с обновленными зависимостями.
//
// Returns:
//
//	error - Ошибка выполнения операции.
//	int - HTTP статус код:
//	    - 200 OK при успешном обновлении
//	    - 500 Internal Server Error при ошибках
func (r *TasksRepository) UpdateTaskDependencies(ctx context.Context, tx *sql.Tx, task *models.Task) (error, int) {
	err := r.depsRepo.UpdateTaskDeps(ctx, tx, task.ID, task.Dependencies)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

// UpdateTaskArguments обновляет один из аргументов задачи.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных.
//	id: int64 - ID задачи
//	index: int - Индекс аргумента (0 или 1).
//	value: *float64 - Новое значение аргумента.
//
// Returns:
//
//	error - Ошибка выполнения операции.
//	int - HTTP статус код:
//	    - 200 OK при успешном обновлении
//	    - 500 Internal Server Error при ошибках
func (r *TasksRepository) UpdateTaskArguments(ctx context.Context, tx *sql.Tx, id int64, index int, value *float64) (error, int) {
	err := r.argsRepo.UpdateTaskArgs(ctx, tx, id, index, value)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

// UpdateTaskStatus обновляет статус задачи.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных.
//	id: int64 - ID задачи.
//	status: string - Новый статус задачи.
//
// Returns:
//
//	error - Ошибка выполнения операции.
//	int - HTTP статус код:
//	    - 200 OK при успешном обновлении
//	    - 500 Internal Server Error при ошибках
func (r *TasksRepository) UpdateTaskStatus(ctx context.Context, tx *sql.Tx, id int64, status string) (error, int) {
	query := `
		UPDATE
		    tasks
		SET
		    status = ?
		WHERE
		    id = ?`

	_, err := tx.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("не удалось обновить статус задачи: %w", err), http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

// UpdateTaskExpressionID обновляет ID выражения для задачи.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных.
//	id: int64 - ID задачи.
//	exprId: int64 - Новый ID выражения.
//
// Returns:
//
//	error - Ошибка выполнения операции
//	int - HTTP статус код:
//	    - 200 OK при успешном обновлении
//	    - 500 Internal Server Error при ошибках
func (r *TasksRepository) UpdateTaskExpressionID(ctx context.Context, tx *sql.Tx, id, exprId int64) (error, int) {
	query := `
	UPDATE
	    tasks
	SET
	    expression_id = ?
	WHERE
	    id = ?`

	_, err := tx.ExecContext(ctx, query, exprId, id)
	if err != nil {
		return fmt.Errorf("не удалось обновить id задачи выражения: %w", err), http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

// UpdateTaskResult обновляет результат выполнения задачи.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных.
//	result: float64 - Новый результат.
//	id: int64 - ID задачи.
//
// Returns:
//
//	error - Ошибка выполнения операции
//	int - HTTP статус код:
//	    - 200 OK при успешном обновлении
//	    - 500 Internal Server Error при ошибках
func (r *TasksRepository) UpdateTaskResult(ctx context.Context, tx *sql.Tx, result float64, id int64) (error, int) {
	query := `
	UPDATE
	    tasks
	SET
	    result = ?
	WHERE
	    id = ?`

	_, err := tx.ExecContext(ctx, query, result, id)
	if err != nil {
		return fmt.Errorf("не удалось обновить результат задачи: %w", err), http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

// DeleteTasks удаляет все задачи, связанные с указанным выражением.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения запроса.
//	tx: *sql.Tx - Транзакция базы данных.
//	id: int64 - ID выражения.
//
// Returns:
//
//	error - Ошибка выполнения операции
//	int - HTTP статус код:
//	    - 200 OK при успешном удалении
//	    - 500 Internal Server Error при ошибках
func (r *TasksRepository) DeleteTasks(ctx context.Context, tx *sql.Tx, id int64) (error, int) {
	query := `DELETE FROM tasks WHERE expression_id = ?`

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("не удалось удалить задачи %w", err), http.StatusInternalServerError
	}
	return nil, http.StatusOK
}
