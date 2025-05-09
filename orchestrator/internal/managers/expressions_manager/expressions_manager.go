package expressions_manager

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories"
	"github.com/OinkiePie/calc_3/orchestrator/internal/task_splitter"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/OinkiePie/calc_3/pkg/operators"
	"net/http"
)

// ExpressionManager предоставляет методы для управления математическими выражениями.
type ExpressionManager struct {
	db       *sql.DB                                     // Подключение к базе данных
	exprRepo repositories.ExpressionsRepositoryInterface // Репозиторий выражений
	taskRepo repositories.TasksRepositoryInterface       // Репозиторий задач
}

// NewExpressionManager создает новый экземпляр менеджера выражений.
//
// Args:
//
//	db: *sql.DB - Подключение к базе данных
//	exprRepo: *repositories2.ExpressionsRepository - Репозиторий выражений
//	taskRepo: *repositories2.TasksRepository - Репозиторий задач
//
// Returns:
//
//	*ExpressionManager - Новый экземпляр менеджера
func NewExpressionManager(
	db *sql.DB,
	exprRepo repositories.ExpressionsRepositoryInterface,
	taskRepo repositories.TasksRepositoryInterface,
) *ExpressionManager {
	return &ExpressionManager{
		db:       db,
		exprRepo: exprRepo,
		taskRepo: taskRepo,
	}
}

// AddExpression добавляет новое выражение в систему и создает связанные задачи.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения.
//	expressionString: string - Строка с математическим выражением.
//	claims: int64 - ID пользователя-владельца.
//
// Returns:
//
//	int64 - ID созданного выражения.
//	error - Ошибка выполнения.
//	int - HTTP статус код:
//		- 201 Created при успешном выполнении
//		- 400 Bad Request при невозможность преобразовать выражение
//		- 500 Internal Server Error при ошибках
func (m *ExpressionManager) AddExpression(ctx context.Context, expressionString string, claims int64) (int64, error, int) {
	tasks, err := task_splitter.ParseExpression(expressionString)
	if err != nil {
		return 0, err, http.StatusBadRequest
	}

	expression := models.Expression{
		Tasks:            tasks,
		ExpressionString: expressionString,
		UserID:           claims,
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("не удалось начать добавление выражения: %w", err), http.StatusInternalServerError
	}
	defer tx.Rollback()

	id, err, code := m.exprRepo.CreateExpression(ctx, tx, &expression)
	if err != nil {
		return 0, err, code
	}
	for _, task := range expression.Tasks {
		task.Expression = id
		if err, code = m.taskRepo.UpdateTaskExpressionID(ctx, tx, task.ID, id); err != nil {
			return 0, err, code
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("не удалось создать выражение: %w", err), http.StatusInternalServerError
	}

	return id, nil, http.StatusCreated
}

// ReadExpressions получает все выражения пользователя.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения.
//	id: int64 - ID пользователя.
//
// Returns:
//
//	[]*models.Expression - Список выражений пользователя.
//	error - Ошибка выполнения.
//	int - HTTP статус код:
//		- 200 OK при успешном выполнении
//		- 404 Not Found если выражения не найдены
//		- 500 Internal Server Error при ошибках
func (m *ExpressionManager) ReadExpressions(ctx context.Context, id int64) ([]*models.Expression, error, int) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось начать получение выражения: %w", err), http.StatusInternalServerError
	}
	defer tx.Rollback()

	expressions, err, code := m.exprRepo.ReadExpressionsByUserID(ctx, tx, id)
	if err != nil {
		return nil, err, code
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("не удалось получить выражение: %w", err), http.StatusInternalServerError
	}

	return expressions, nil, http.StatusOK
}

// ReadExpression получает выражение по его ID.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения.
//	id: int64 - ID выражения.
//
// Returns:
//
//	*models.Expression - Найденное выражение.
//	error - Ошибка выполнения.
//	int - HTTP статус код:
//		- 200 OK при успешном выполнении
//		- 404 Not Found если выражение не найдено
//		- 500 Internal Server Error при ошибках
func (m *ExpressionManager) ReadExpression(ctx context.Context, id int64) (*models.Expression, error, int) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось начать отправку выражение: %w", err), http.StatusInternalServerError
	}
	defer tx.Rollback()

	expression, err, code := m.exprRepo.ReadExpressionByID(ctx, tx, id)
	if err != nil {
		return nil, err, code
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("не удалось отправить выражение: %w", err), http.StatusInternalServerError
	}

	return expression, nil, http.StatusOK
}

// ReadTask находит и возвращает следующую задачу для выполнения.
// Проверяет готовность зависимостей и обновляет статусы.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения
//
// Returns:
//
//	*models.Task - Готовая к выполнению задача
//	error - Ошибка выполнения
//	int - HTTP статус код:
//		- 200 OK при успешном получении
//		- 404 Not Found если задач нет
//	    - 500 Internal Server Error при ошибках
func (m *ExpressionManager) ReadTask(ctx context.Context) (*models.Task, error, int) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось начать отправку задачи: %w", err), http.StatusInternalServerError
	}
	defer tx.Rollback()

	tasks, err, code := m.taskRepo.ReadUncompletedTasks(ctx, tx)
	if err != nil {
		return nil, err, code
	}

outerLoop:
	for _, task := range tasks {
		for i := range task.Args {
			if task.Args[i] == nil && (i == 0 || task.Operation != operators.OpUnaryMinus) {
				dep, err, code := m.taskRepo.ReadTaskByID(ctx, tx, task.Dependencies[i])
				if dep == nil {
					return nil, err, code
				}
				if dep.Status != "completed" {
					continue outerLoop
				}
				err, code = m.taskRepo.UpdateTaskArguments(ctx, tx, task.ID, i, dep.Result)
				if err != nil {
					return nil, err, code
				}
				task.Args[i] = dep.Result
			}
		}

		if err, code = m.taskRepo.UpdateTaskStatus(ctx, tx, task.ID, "processing"); err != nil {
			return nil, err, code
		}
		task.Status = "processing"
		if err, code = m.exprRepo.UpdateExpressionStatus(ctx, tx, task.Expression, "processing"); err != nil {
			return nil, err, code
		}
		if err = tx.Commit(); err != nil {
			return nil, fmt.Errorf("не удалось отправить задачу: %w", err), http.StatusInternalServerError
		}
		return task, nil, http.StatusOK
	}

	return nil, nil, http.StatusNotFound
}

// CompleteTask завершает выполнение задачи и обновляет связанные данные.
// При ошибке в задаче помечает всё выражение как ошибочное.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения
//	taskCompleted: *models.TaskCompleted - Данные выполненной задачи
//
// Returns:
//
//	error - Ошибка выполнения
//	int - HTTP статус код:
//		- 200 OK при успешном выполнении
//	    - 500 Internal Server Error при ошибках
func (m *ExpressionManager) CompleteTask(ctx context.Context, taskCompleted *models.TaskCompleted) (error, int) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("не удалось начать завершение задачи: %w", err), http.StatusInternalServerError
	}
	defer tx.Rollback()

	if taskCompleted.Error != "" {
		if err, code := m.exprRepo.UpdateExpressionError(ctx, tx, taskCompleted.Expression, taskCompleted.Error); err != nil {
			return err, code
		}
		if err, code := m.exprRepo.UpdateExpressionStatus(ctx, tx, taskCompleted.Expression, "error"); err != nil {
			return err, code
		}
		if err, code := m.taskRepo.DeleteTasks(ctx, tx, taskCompleted.Expression); err != nil {
			return err, code
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("не удалось завершить задачу: %w", err), http.StatusInternalServerError
		}
		return nil, http.StatusOK
	}

	if err, code := m.taskRepo.UpdateTaskResult(ctx, tx, taskCompleted.Result, taskCompleted.ID); err != nil {
		return err, code
	}
	if err, code := m.taskRepo.UpdateTaskStatus(ctx, tx, taskCompleted.ID, "completed"); err != nil {
		return err, code
	}

	tasks, err, code := m.taskRepo.ReadTasksByExpressionID(ctx, tx, taskCompleted.Expression)
	if err != nil {
		return err, code
	}
	allCompleted := true
	for _, task := range tasks {
		if task.Status != "completed" {
			allCompleted = false
			break
		}
	}

	if allCompleted {
		if err, code = m.exprRepo.UpdateExpressionStatus(ctx, tx, taskCompleted.Expression, "completed"); err != nil {
			return err, code
		}
		if err, code = m.exprRepo.UpdateExpressionResult(ctx, tx, taskCompleted.Expression, taskCompleted.Result); err != nil {
			return err, code
		}
		if err, code = m.taskRepo.DeleteTasks(ctx, tx, taskCompleted.Expression); err != nil {
			return err, code
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("не удалось завершить задачу: %w", err), http.StatusInternalServerError
	}

	return nil, http.StatusOK
}
