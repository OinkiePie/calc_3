package expressions_manager_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/orchestrator/internal/managers/expressions_manager"
	mr "github.com/OinkiePie/calc_3/orchestrator/internal/repositories"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories/expressions_repository"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories/tasks_repository"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/OinkiePie/calc_3/pkg/operators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"log"
	"net/http"
	"testing"
)

func init() {
	// Отключаем выводы и инициализируем конфиг
	log.SetOutput(io.Discard)
	_ = config.InitConfig()
	logger.InitLogger(logger.Options{Level: 6})
}

func TestNew(t *testing.T) {
	mockDB := &sql.DB{}
	mockExpressionsRepo := new(mr.MockExpressionsRepository)
	mockTasksRepo := new(mr.MockTasksRepository)

	manager := expressions_manager.NewExpressionManager(
		mockDB,
		mockExpressionsRepo,
		mockTasksRepo,
	)

	assert.NotNil(t, manager)
}

// МОДУЛЬНЫЕ ТЕСТЫ

func TestExpressionManager_AddExpression(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mockExprRepo := new(mr.MockExpressionsRepository)
	mockTaskRepo := new(mr.MockTasksRepository)

	manager := expressions_manager.NewExpressionManager(db, mockExprRepo, mockTaskRepo)

	ctx := context.Background()
	validExpression := "2 + 2"
	invalidExpression := "2 + "
	userID := int64(1)

	t.Run("successful expression addition", func(t *testing.T) {
		mockExprRepo.On("CreateExpression", ctx, mock.AnythingOfType("*sql.Tx"), mock.AnythingOfType("*models.Expression")).
			Return(int64(1), nil, http.StatusCreated).Once()

		mockTaskRepo.On("UpdateTaskExpressionID", ctx, mock.AnythingOfType("*sql.Tx"), int64(1), int64(1)).
			Return(nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		id, err, code := manager.AddExpression(ctx, validExpression, userID)

		assert.Equal(t, int64(1), id)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, code)

		mockExprRepo.AssertExpectations(t)
		mockTaskRepo.AssertExpectations(t)

	})

	t.Run("invalid expression", func(t *testing.T) {
		_, err, code := manager.AddExpression(ctx, invalidExpression, userID)

		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, code)
	})

	t.Run("failed to create expression", func(t *testing.T) {
		mockExprRepo.On("CreateExpression", ctx, mock.AnythingOfType("*sql.Tx"), mock.Anything).
			Return(int64(0), errors.New("database error"), http.StatusInternalServerError).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectRollback()

		_, err, code := manager.AddExpression(ctx, validExpression, userID)

		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, code)

	})

	t.Run("failed to update task expression ID", func(t *testing.T) {
		mockExprRepo.On("CreateExpression", ctx, mock.AnythingOfType("*sql.Tx"), mock.Anything).
			Return(int64(1), nil, http.StatusCreated).Once()

		mockTaskRepo.On("UpdateTaskExpressionID", ctx, mock.AnythingOfType("*sql.Tx"), int64(1), int64(1)).
			Return(errors.New("update error"), http.StatusInternalServerError).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectRollback()

		_, err, code := manager.AddExpression(ctx, validExpression, userID)

		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, code)

	})

	t.Run("transaction begin error", func(t *testing.T) {
		mockDB.ExpectBegin().WillReturnError(errors.New("begin error"))

		_, err, code := manager.AddExpression(ctx, validExpression, userID)

		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, code)

	})

	t.Run("transaction commit error", func(t *testing.T) {
		mockExprRepo.On("CreateExpression", ctx, mock.AnythingOfType("*sql.Tx"), mock.AnythingOfType("*models.Expression")).
			Return(int64(1), nil, http.StatusCreated).Once()

		mockTaskRepo.On("UpdateTaskExpressionID", ctx, mock.AnythingOfType("*sql.Tx"), int64(1), int64(1)).
			Return(nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit().WillReturnError(errors.New("commit error"))

		_, err, code := manager.AddExpression(ctx, validExpression, userID)

		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, code)

	})
}

func TestExpressionManager_ReadExpressions(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mockExprRepo := new(mr.MockExpressionsRepository)
	mockTaskRepo := new(mr.MockTasksRepository)

	manager := expressions_manager.NewExpressionManager(db, mockExprRepo, mockTaskRepo)

	ctx := context.Background()
	userID := int64(1)
	expressions := []*models.Expression{
		{ID: 1, ExpressionString: "2 + 2", UserID: userID, Status: "completed"},
		{ID: 2, ExpressionString: "3 * 3", UserID: userID, Status: "pending"},
	}

	t.Run("successful read expressions", func(t *testing.T) {
		mockExprRepo.On("ReadExpressionsByUserID", ctx, mock.AnythingOfType("*sql.Tx"), userID).
			Return(expressions, nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		result, err, code := manager.ReadExpressions(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, expressions, result)
		mockExprRepo.AssertExpectations(t)

	})

	t.Run("expressions not found", func(t *testing.T) {
		mockExprRepo.On("ReadExpressionsByUserID", ctx, mock.AnythingOfType("*sql.Tx"), userID).
			Return(([]*models.Expression)(nil), errors.New("not found error"), http.StatusNotFound).Once()

		mockDB.ExpectBegin()

		result, err, code := manager.ReadExpressions(ctx, userID)

		assert.Error(t, err)
		assert.Equal(t, http.StatusNotFound, code)
		assert.Empty(t, result)
		mockExprRepo.AssertExpectations(t)

	})

	t.Run("repository error", func(t *testing.T) {
		mockExprRepo.On("ReadExpressionsByUserID", ctx, mock.AnythingOfType("*sql.Tx"), userID).
			Return(([]*models.Expression)(nil), errors.New("database error"), http.StatusInternalServerError).Once()

		mockDB.ExpectBegin()

		result, err, code := manager.ReadExpressions(ctx, userID)

		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
		assert.Equal(t, http.StatusInternalServerError, code)
		assert.Nil(t, result)
		mockExprRepo.AssertExpectations(t)

	})

	t.Run("transaction begin error", func(t *testing.T) {
		mockDB.ExpectBegin().WillReturnError(errors.New("begin error"))

		result, err, code := manager.ReadExpressions(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось начать получение выражения")
		assert.Equal(t, http.StatusInternalServerError, code)
		assert.Nil(t, result)

	})

	t.Run("transaction commit error", func(t *testing.T) {
		mockExprRepo.On("ReadExpressionsByUserID", ctx, mock.AnythingOfType("*sql.Tx"), userID).
			Return(expressions, nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit().WillReturnError(errors.New("commit error"))

		result, err, code := manager.ReadExpressions(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось получить выражение")
		assert.Equal(t, http.StatusInternalServerError, code)
		assert.Nil(t, result)

	})
}

func TestExpressionManager_ReadExpression(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mockExprRepo := new(mr.MockExpressionsRepository)
	mockTaskRepo := new(mr.MockTasksRepository)

	manager := expressions_manager.NewExpressionManager(db, mockExprRepo, mockTaskRepo)

	ctx := context.Background()
	exprID := int64(1)
	testExpression := &models.Expression{
		ID:               exprID,
		ExpressionString: "2 + 2",
		UserID:           1,
		Status:           "completed",
	}

	t.Run("successful read expression", func(t *testing.T) {
		mockExprRepo.On("ReadExpressionByID", ctx, mock.AnythingOfType("*sql.Tx"), exprID).
			Return(testExpression, nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		result, err, code := manager.ReadExpression(ctx, exprID)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, testExpression, result)
		mockExprRepo.AssertExpectations(t)

	})

	t.Run("expression not found", func(t *testing.T) {
		mockExprRepo.On("ReadExpressionByID", ctx, mock.AnythingOfType("*sql.Tx"), exprID).
			Return((*models.Expression)(nil), errors.New("not found error"), http.StatusNotFound).Once()

		mockDB.ExpectBegin()

		result, err, code := manager.ReadExpression(ctx, exprID)

		assert.Error(t, err)
		assert.Equal(t, http.StatusNotFound, code)
		assert.Nil(t, result)
		mockExprRepo.AssertExpectations(t)

	})

	t.Run("repository error", func(t *testing.T) {
		mockExprRepo.On("ReadExpressionByID", ctx, mock.AnythingOfType("*sql.Tx"), exprID).
			Return((*models.Expression)(nil), errors.New("database error"), http.StatusInternalServerError).Once()

		mockDB.ExpectBegin()

		result, err, code := manager.ReadExpression(ctx, exprID)

		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
		assert.Equal(t, http.StatusInternalServerError, code)
		assert.Nil(t, result)
		mockExprRepo.AssertExpectations(t)

	})

	t.Run("transaction begin error", func(t *testing.T) {
		mockDB.ExpectBegin().WillReturnError(errors.New("begin error"))

		result, err, code := manager.ReadExpression(ctx, exprID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось начать отправку выражение")
		assert.Equal(t, http.StatusInternalServerError, code)
		assert.Nil(t, result)

	})

	t.Run("transaction commit error", func(t *testing.T) {
		mockDB.ExpectBegin()
		mockDB.ExpectCommit().WillReturnError(errors.New("commit error"))

		mockExprRepo.On("ReadExpressionByID", ctx, mock.AnythingOfType("*sql.Tx"), exprID).
			Return(testExpression, nil, http.StatusOK).Once()

		result, err, code := manager.ReadExpression(ctx, exprID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось отправить выражение")
		assert.Equal(t, http.StatusInternalServerError, code)
		assert.Nil(t, result)

	})
}

func TestExpressionManager_ReadTask(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mockExprRepo := new(mr.MockExpressionsRepository)
	mockTaskRepo := new(mr.MockTasksRepository)

	manager := expressions_manager.NewExpressionManager(db, mockExprRepo, mockTaskRepo)

	ctx := context.Background()

	t.Run("successful read ready task", func(t *testing.T) {
		readyTask := &models.Task{
			ID:         1,
			Operation:  "+",
			Args:       []*float64{mr.Float64Ptr(2), mr.Float64Ptr(3)},
			Status:     "pending",
			Expression: 1,
		}

		mockTaskRepo.On("ReadUncompletedTasks", ctx, mock.AnythingOfType("*sql.Tx")).
			Return([]*models.Task{readyTask}, nil, http.StatusOK).Once()

		mockTaskRepo.On("UpdateTaskStatus", ctx, mock.AnythingOfType("*sql.Tx"), readyTask.ID, "processing").
			Return(nil, http.StatusOK).Once()

		mockExprRepo.On("UpdateExpressionStatus", ctx, mock.AnythingOfType("*sql.Tx"), readyTask.Expression, "processing").
			Return(nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		result, err, code := manager.ReadTask(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, readyTask.ID, result.ID)
		assert.Equal(t, "processing", result.Status)
		mockTaskRepo.AssertExpectations(t)
		mockExprRepo.AssertExpectations(t)

	})

	t.Run("task with unresolved dependencies", func(t *testing.T) {
		unresolvedDepID := int64(2)
		taskWithDeps := &models.Task{
			ID:           1,
			Operation:    "+",
			Args:         []*float64{nil, mr.Float64Ptr(3)},
			Dependencies: []int64{unresolvedDepID, 0},
			Status:       "pending",
		}

		depTask := &models.Task{
			ID:     unresolvedDepID,
			Status: "pending",
		}

		mockTaskRepo.On("ReadUncompletedTasks", ctx, mock.AnythingOfType("*sql.Tx")).
			Return([]*models.Task{taskWithDeps}, nil, http.StatusOK).Once()

		mockTaskRepo.On("ReadTaskByID", ctx, mock.AnythingOfType("*sql.Tx"), unresolvedDepID).
			Return(depTask, nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectRollback()

		result, err, code := manager.ReadTask(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, code)
		assert.Nil(t, result)
		mockTaskRepo.AssertExpectations(t)

	})

	t.Run("task with resolved dependencies", func(t *testing.T) {
		depID := int64(2)
		resolvedDepValue := float64(5)
		taskWithDeps := &models.Task{
			ID:           1,
			Operation:    "+",
			Args:         []*float64{nil, mr.Float64Ptr(3)},
			Dependencies: []int64{depID, -1},
			Status:       "pending",
			Expression:   1,
		}

		completedDepTask := &models.Task{
			ID:     depID,
			Status: "completed",
			Result: &resolvedDepValue,
		}

		mockTaskRepo.On("ReadUncompletedTasks", ctx, mock.AnythingOfType("*sql.Tx")).
			Return([]*models.Task{taskWithDeps}, nil, http.StatusOK).Once()

		mockTaskRepo.On("ReadTaskByID", ctx, mock.AnythingOfType("*sql.Tx"), depID).
			Return(completedDepTask, nil, http.StatusOK).Once()

		mockTaskRepo.On("UpdateTaskArguments", ctx, mock.AnythingOfType("*sql.Tx"), taskWithDeps.ID, 0, &resolvedDepValue).
			Return(nil, http.StatusOK).Once()

		mockTaskRepo.On("UpdateTaskStatus", ctx, mock.AnythingOfType("*sql.Tx"), taskWithDeps.ID, "processing").
			Return(nil, http.StatusOK).Once()

		mockExprRepo.On("UpdateExpressionStatus", ctx, mock.AnythingOfType("*sql.Tx"), taskWithDeps.Expression, "processing").
			Return(nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		result, err, code := manager.ReadTask(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, taskWithDeps.ID, result.ID)
		assert.Equal(t, "processing", result.Status)
		assert.Equal(t, &resolvedDepValue, result.Args[0])
		mockTaskRepo.AssertExpectations(t)
		mockExprRepo.AssertExpectations(t)

	})

	t.Run("unary minus task", func(t *testing.T) {
		unaryTask := &models.Task{
			ID:         1,
			Operation:  operators.OpUnaryMinus,
			Args:       []*float64{mr.Float64Ptr(1), nil},
			Status:     "pending",
			Expression: 1,
		}

		mockTaskRepo.On("ReadUncompletedTasks", ctx, mock.AnythingOfType("*sql.Tx")).
			Return([]*models.Task{unaryTask}, nil, http.StatusOK).Once()

		mockTaskRepo.On("UpdateTaskStatus", ctx, mock.AnythingOfType("*sql.Tx"), unaryTask.ID, "processing").
			Return(nil, http.StatusOK).Once()

		mockExprRepo.On("UpdateExpressionStatus", ctx, mock.AnythingOfType("*sql.Tx"), unaryTask.Expression, "processing").
			Return(nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		result, err, code := manager.ReadTask(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, unaryTask.ID, result.ID)
		mockTaskRepo.AssertExpectations(t)
		mockExprRepo.AssertExpectations(t)

	})

	t.Run("no tasks available", func(t *testing.T) {
		mockTaskRepo.On("ReadUncompletedTasks", ctx, mock.AnythingOfType("*sql.Tx")).
			Return([]*models.Task{}, nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectRollback()

		result, err, code := manager.ReadTask(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, code)
		assert.Nil(t, result)
		mockTaskRepo.AssertExpectations(t)

	})

	t.Run("error reading uncompleted tasks", func(t *testing.T) {
		mockTaskRepo.On("ReadUncompletedTasks", ctx, mock.AnythingOfType("*sql.Tx")).
			Return(([]*models.Task)(nil), errors.New("database error"), http.StatusInternalServerError).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectRollback()

		result, err, code := manager.ReadTask(ctx)

		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, code)
		assert.Nil(t, result)
		mockTaskRepo.AssertExpectations(t)

	})

	t.Run("transaction begin error", func(t *testing.T) {
		mockDB.ExpectBegin().WillReturnError(errors.New("begin error"))

		result, err, code := manager.ReadTask(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось начать отправку задачи")
		assert.Equal(t, http.StatusInternalServerError, code)
		assert.Nil(t, result)

	})

	t.Run("transaction commit error", func(t *testing.T) {
		readyTask := &models.Task{
			ID:         1,
			Operation:  "+",
			Args:       []*float64{mr.Float64Ptr(2), mr.Float64Ptr(3)},
			Status:     "pending",
			Expression: 1,
		}

		mockTaskRepo.On("ReadUncompletedTasks", ctx, mock.AnythingOfType("*sql.Tx")).
			Return([]*models.Task{readyTask}, nil, http.StatusOK).Once()

		mockTaskRepo.On("UpdateTaskStatus", ctx, mock.AnythingOfType("*sql.Tx"), readyTask.ID, "processing").
			Return(nil, http.StatusOK).Once()

		mockExprRepo.On("UpdateExpressionStatus", ctx, mock.AnythingOfType("*sql.Tx"), readyTask.Expression, "processing").
			Return(nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit().WillReturnError(errors.New("commit error"))

		result, err, code := manager.ReadTask(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось отправить задачу")
		assert.Equal(t, http.StatusInternalServerError, code)
		assert.Nil(t, result)

	})
}

func TestExpressionManager_CompleteTask(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mockExprRepo := new(mr.MockExpressionsRepository)
	mockTaskRepo := new(mr.MockTasksRepository)

	manager := expressions_manager.NewExpressionManager(db, mockExprRepo, mockTaskRepo)

	ctx := context.Background()
	successResult := float64(5)
	exprID := int64(1)
	taskID := int64(10)

	t.Run("successful task completion", func(t *testing.T) {
		taskCompleted := &models.TaskCompleted{
			ID:         taskID,
			Expression: exprID,
			Result:     successResult,
			Error:      "",
		}

		mockTaskRepo.On("UpdateTaskResult", ctx, mock.AnythingOfType("*sql.Tx"), successResult, taskID).
			Return(nil, http.StatusOK).Once()

		mockTaskRepo.On("UpdateTaskStatus", ctx, mock.AnythingOfType("*sql.Tx"), taskID, "completed").
			Return(nil, http.StatusOK).Once()

		mockTaskRepo.On("ReadTasksByExpressionID", ctx, mock.AnythingOfType("*sql.Tx"), exprID).
			Return([]*models.Task{
				{ID: taskID, Status: "completed"},
				{ID: 2, Status: "completed"},
			}, nil, http.StatusOK).Once()

		mockExprRepo.On("UpdateExpressionStatus", ctx, mock.AnythingOfType("*sql.Tx"), exprID, "completed").
			Return(nil, http.StatusOK).Once()

		mockExprRepo.On("UpdateExpressionResult", ctx, mock.AnythingOfType("*sql.Tx"), exprID, successResult).
			Return(nil, http.StatusOK).Once()

		mockTaskRepo.On("DeleteTasks", ctx, mock.AnythingOfType("*sql.Tx"), exprID).
			Return(nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		err, code := manager.CompleteTask(ctx, taskCompleted)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		mockTaskRepo.AssertExpectations(t)
		mockExprRepo.AssertExpectations(t)

	})

	t.Run("task with error marks expression as error", func(t *testing.T) {
		taskError := "division by zero"
		taskCompleted := &models.TaskCompleted{
			ID:         taskID,
			Expression: exprID,
			Error:      taskError,
		}

		mockExprRepo.On("UpdateExpressionError", ctx, mock.AnythingOfType("*sql.Tx"), exprID, taskError).
			Return(nil, http.StatusOK).Once()

		mockExprRepo.On("UpdateExpressionStatus", ctx, mock.AnythingOfType("*sql.Tx"), exprID, "error").
			Return(nil, http.StatusOK).Once()

		mockTaskRepo.On("DeleteTasks", ctx, mock.AnythingOfType("*sql.Tx"), exprID).
			Return(nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		err, code := manager.CompleteTask(ctx, taskCompleted)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		mockExprRepo.AssertExpectations(t)
		mockTaskRepo.AssertExpectations(t)

	})

	t.Run("not all tasks completed", func(t *testing.T) {
		taskCompleted := &models.TaskCompleted{
			ID:         taskID,
			Expression: exprID,
			Result:     successResult,
			Error:      "",
		}

		mockTaskRepo.On("UpdateTaskResult", ctx, mock.AnythingOfType("*sql.Tx"), successResult, taskID).
			Return(nil, http.StatusOK).Once()

		mockTaskRepo.On("UpdateTaskStatus", ctx, mock.AnythingOfType("*sql.Tx"), taskID, "completed").
			Return(nil, http.StatusOK).Once()

		mockTaskRepo.On("ReadTasksByExpressionID", ctx, mock.AnythingOfType("*sql.Tx"), exprID).
			Return([]*models.Task{
				{ID: taskID, Status: "completed"},
				{ID: 2, Status: "pending"},
			}, nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		err, code := manager.CompleteTask(ctx, taskCompleted)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		mockTaskRepo.AssertExpectations(t)
		mockExprRepo.AssertNotCalled(t, "UpdateExpressionStatus")

	})

	t.Run("error updating task result", func(t *testing.T) {
		taskCompleted := &models.TaskCompleted{
			ID:         taskID,
			Expression: exprID,
			Result:     successResult,
			Error:      "",
		}

		mockTaskRepo.On("UpdateTaskResult", ctx, mock.AnythingOfType("*sql.Tx"), successResult, taskID).
			Return(errors.New("db error"), http.StatusInternalServerError).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectRollback()

		err, code := manager.CompleteTask(ctx, taskCompleted)

		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, code)
		mockTaskRepo.AssertExpectations(t)

	})

	t.Run("transaction begin error", func(t *testing.T) {
		taskCompleted := &models.TaskCompleted{
			ID:         taskID,
			Expression: exprID,
		}

		mockDB.ExpectBegin().WillReturnError(errors.New("begin error"))

		err, code := manager.CompleteTask(ctx, taskCompleted)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось начать завершение задачи")
		assert.Equal(t, http.StatusInternalServerError, code)

	})

	t.Run("transaction commit error", func(t *testing.T) {
		taskCompleted := &models.TaskCompleted{
			ID:         taskID,
			Expression: exprID,
			Result:     successResult,
			Error:      "",
		}

		mockTaskRepo.On("UpdateTaskResult", ctx, mock.AnythingOfType("*sql.Tx"), successResult, taskID).
			Return(nil, http.StatusOK).Once()
		mockTaskRepo.On("UpdateTaskStatus", ctx, mock.AnythingOfType("*sql.Tx"), taskID, "completed").
			Return(nil, http.StatusOK).Once()
		mockTaskRepo.On("ReadTasksByExpressionID", ctx, mock.AnythingOfType("*sql.Tx"), exprID).
			Return([]*models.Task{
				{ID: taskID, Status: "completed"},
				{ID: 2, Status: "completed"},
			}, nil, http.StatusOK).Once()
		mockExprRepo.On("UpdateExpressionStatus", ctx, mock.AnythingOfType("*sql.Tx"), exprID, "completed").
			Return(nil, http.StatusOK).Once()
		mockExprRepo.On("UpdateExpressionResult", ctx, mock.AnythingOfType("*sql.Tx"), exprID, successResult).
			Return(nil, http.StatusOK).Once()
		mockTaskRepo.On("DeleteTasks", ctx, mock.AnythingOfType("*sql.Tx"), exprID).
			Return(nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit().WillReturnError(errors.New("commit error"))

		err, code := manager.CompleteTask(ctx, taskCompleted)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось завершить задачу")
		assert.Equal(t, http.StatusInternalServerError, code)

	})
}

// ИНТЕГРАЦИОННЫЕ ТЕСТЫ

func TestExpressionManager_AddExpression_Integration(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:testdb?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if setupTestDatabase(db) != nil {
		t.Fatal(err)
	}

	depsRepo := tasks_repository.NewTaskDepsRepository(db)
	argsRepo := tasks_repository.NewTaskArgsRepository(db)
	taskRepo := tasks_repository.NewTasksRepository(db, depsRepo, argsRepo)
	exprRepo := expressions_repository.NewExpressionsRepository(db, taskRepo)

	manager := expressions_manager.NewExpressionManager(db, exprRepo, taskRepo)

	ctx := context.Background()
	validExpression := "2 + 2"
	userID := int64(1)

	t.Run("successful integration", func(t *testing.T) {
		id, err, code := manager.AddExpression(ctx, validExpression, userID)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, code)
		assert.True(t, id > 0)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()

		expr, err, _ := exprRepo.ReadExpressionByID(ctx, tx, id)
		assert.NoError(t, err)
		assert.Equal(t, validExpression, expr.ExpressionString)
		assert.Equal(t, userID, expr.UserID)

		tasks, err, _ := taskRepo.ReadTasksByExpressionID(ctx, tx, id)
		assert.NoError(t, err)
		assert.True(t, len(tasks) > 0)

		if err := tx.Commit(); err != nil {
			t.Fatalf("не удалось закоммитить транзакцию: %v", err)
		}
	})
}

func TestExpressionManager_ReadExpressions_Integration(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:testdb?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if setupTestDatabase(db) != nil {
		t.Fatal(err)
	}

	depsRepo := tasks_repository.NewTaskDepsRepository(db)
	argsRepo := tasks_repository.NewTaskArgsRepository(db)
	taskRepo := tasks_repository.NewTasksRepository(db, depsRepo, argsRepo)
	exprRepo := expressions_repository.NewExpressionsRepository(db, taskRepo)

	manager := expressions_manager.NewExpressionManager(db, exprRepo, taskRepo)

	ctx := context.Background()
	userID := int64(1)

	testExpressions := []*models.Expression{
		{ExpressionString: "2 + 2", UserID: userID, Status: "completed"},
		{ExpressionString: "3 * 3", UserID: userID, Status: "pending"},
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	var createdIDs []int64
	for _, expr := range testExpressions {
		id, err, code := exprRepo.CreateExpression(ctx, tx, expr)
		if err != nil || code != http.StatusOK {
			t.Fatal(err)
		}
		createdIDs = append(createdIDs, id)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("не удалось закоммитить транзакцию: %v", err)
	}

	t.Run("successful integration", func(t *testing.T) {
		result, err, code := manager.ReadExpressions(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		assert.Len(t, result, len(testExpressions))

		for _, expr := range result {
			assert.Contains(t, createdIDs, expr.ID)
			assert.Equal(t, userID, expr.UserID)
		}
	})

	t.Run("no expressions for user", func(t *testing.T) {
		nonExistentUserID := int64(999)
		result, err, code := manager.ReadExpressions(ctx, nonExistentUserID)

		assert.Error(t, err)
		assert.Equal(t, http.StatusNotFound, code)
		assert.Empty(t, result)
	})
}

func TestExpressionManager_ReadExpression_Integration(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:testdb?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if setupTestDatabase(db) != nil {
		t.Fatal(err)
	}

	depsRepo := tasks_repository.NewTaskDepsRepository(db)
	argsRepo := tasks_repository.NewTaskArgsRepository(db)
	taskRepo := tasks_repository.NewTasksRepository(db, depsRepo, argsRepo)
	exprRepo := expressions_repository.NewExpressionsRepository(db, taskRepo)

	manager := expressions_manager.NewExpressionManager(db, exprRepo, taskRepo)

	ctx := context.Background()

	testExpr := &models.Expression{
		ExpressionString: "2 + 2",
		UserID:           1,
		Status:           "completed",
		Tasks:            []*models.Task{},
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	id, err, code := exprRepo.CreateExpression(ctx, tx, testExpr)
	if err != nil || code != http.StatusOK {
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("не удалось закоммитить транзакцию: %v", err)
	}

	t.Run("successful integration", func(t *testing.T) {
		result, err, code := manager.ReadExpression(ctx, id)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		assert.NotNil(t, result)
		assert.Equal(t, id, result.ID)
		assert.Equal(t, testExpr.ExpressionString, result.ExpressionString)
	})

	t.Run("expression not found", func(t *testing.T) {
		nonExistentID := int64(999)
		result, err, code := manager.ReadExpression(ctx, nonExistentID)

		assert.Error(t, err)
		assert.Equal(t, http.StatusNotFound, code)
		assert.Nil(t, result)
	})
}

func TestExpressionManager_ReadTask_Integration(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:testdb?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if setupTestDatabase(db) != nil {
		t.Fatal(err)
	}

	depsRepo := tasks_repository.NewTaskDepsRepository(db)
	argsRepo := tasks_repository.NewTaskArgsRepository(db)
	taskRepo := tasks_repository.NewTasksRepository(db, depsRepo, argsRepo)
	exprRepo := expressions_repository.NewExpressionsRepository(db, taskRepo)

	manager := expressions_manager.NewExpressionManager(db, exprRepo, taskRepo)

	ctx := context.Background()

	t.Run("successful integration with ready task", func(t *testing.T) {
		setupTx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}

		expr := &models.Expression{
			ExpressionString: "2 + 3",
			UserID:           1,
			Status:           "pending",
			Tasks: []*models.Task{
				{
					ID:                1,
					Operation:         "+",
					Args:              []*float64{mr.Float64Ptr(2), mr.Float64Ptr(3)},
					Status:            "pending",
					Dependencies:      []int64{0, 0},
					DependencyIndexes: []int{0, 0},
				},
			},
		}

		exprID, err, code := exprRepo.CreateExpression(ctx, setupTx, expr)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)

		if err := setupTx.Commit(); err != nil {
			t.Fatal(err)
		}

		testTx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer testTx.Rollback()

		result, err, code := manager.ReadTask(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		assert.NotNil(t, result)
		assert.Equal(t, "processing", result.Status)

		updatedTask, _, _ := taskRepo.ReadTaskByID(ctx, testTx, result.ID)
		assert.Equal(t, "processing", updatedTask.Status)

		updatedExpr, _, _ := exprRepo.ReadExpressionByID(ctx, testTx, exprID)
		assert.Equal(t, "processing", updatedExpr.Status)

		if err := testTx.Commit(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("successful integration with resolved dependencies", func(t *testing.T) {
		setupTx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}

		expr := &models.Expression{
			ExpressionString: "2 + (3 * 4)",
			UserID:           1,
			Status:           "pending",
			Tasks: []*models.Task{
				{
					ID:                1,
					Operation:         "*",
					Args:              []*float64{mr.Float64Ptr(3), mr.Float64Ptr(4)},
					Status:            "pending",
					Dependencies:      []int64{0, 0},
					DependencyIndexes: []int{0, 0},
					Expression:        1,
				},
				{
					ID:                2,
					Operation:         "+",
					Args:              []*float64{mr.Float64Ptr(2), nil},
					Dependencies:      []int64{0, 1},
					DependencyIndexes: []int{0, 1},
					Status:            "pending",
					Expression:        1,
				},
			},
		}

		_, err, code := exprRepo.CreateExpression(ctx, setupTx, expr)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)

		if err := setupTx.Commit(); err != nil {
			t.Fatal(err)
		}

		testTx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer testTx.Rollback()

		firstTask, err, code := manager.ReadTask(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, expr.Tasks[0].ID, firstTask.ID)

		_, _ = taskRepo.UpdateTaskResult(ctx, testTx, 12, expr.Tasks[0].ID)
		_, _ = taskRepo.UpdateTaskStatus(ctx, testTx, expr.Tasks[0].ID, "completed")

		if err := testTx.Commit(); err != nil {
			t.Fatal(err)
		}
		testTx, err = db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer testTx.Rollback()

		secondTask, err, code := manager.ReadTask(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, expr.Tasks[1].ID, secondTask.ID)
		assert.Equal(t, mr.Float64Ptr(2), secondTask.Args[0])
		assert.Equal(t, mr.Float64Ptr(12), secondTask.Args[1])

		if err := testTx.Commit(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestExpressionManager_CompleteTask_Integration(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:testdb?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if setupTestDatabase(db) != nil {
		t.Fatal(err)
	}

	depsRepo := tasks_repository.NewTaskDepsRepository(db)
	argsRepo := tasks_repository.NewTaskArgsRepository(db)
	taskRepo := tasks_repository.NewTasksRepository(db, depsRepo, argsRepo)
	exprRepo := expressions_repository.NewExpressionsRepository(db, taskRepo)

	manager := expressions_manager.NewExpressionManager(db, exprRepo, taskRepo)

	ctx := context.Background()

	t.Run("successful completion with all tasks done", func(t *testing.T) {
		if clearTestDatabase(db) != nil {
			t.Fatal(err)
		}

		setupTx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}

		expr := &models.Expression{
			ID:               1,
			ExpressionString: "2 + 3",
			UserID:           1,
			Status:           "processing",
			Tasks: []*models.Task{
				{
					Operation:         "+",
					Args:              []*float64{mr.Float64Ptr(2), mr.Float64Ptr(3)},
					Dependencies:      []int64{0, 0},
					DependencyIndexes: []int{0, 0},
					Status:            "processing",
				},
			},
		}

		exprID, _, _ := exprRepo.CreateExpression(ctx, setupTx, expr)

		if err := setupTx.Commit(); err != nil {
			t.Fatal(err)
		}

		testTx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer testTx.Rollback()

		taskCompleted := &models.TaskCompleted{
			ID:         1,
			Expression: exprID,
			Result:     5,
		}

		err, code := manager.CompleteTask(ctx, taskCompleted)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)

		updatedExpr, _, _ := exprRepo.ReadExpressionByID(ctx, testTx, exprID)
		assert.Equal(t, "completed", updatedExpr.Status)
		assert.Equal(t, float64(5), *updatedExpr.Result)

		tasks, _, _ := taskRepo.ReadTasksByExpressionID(ctx, testTx, exprID)
		assert.Empty(t, tasks)

		if err = testTx.Commit(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("task with error marks expression as error", func(t *testing.T) {
		if clearTestDatabase(db) != nil {
			t.Fatal(err)
		}

		setupTx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}

		expr := &models.Expression{
			ID:               1,
			ExpressionString: "2 / 0",
			UserID:           1,
			Status:           "processing",
			Tasks: []*models.Task{
				{
					Operation:         "/",
					Args:              []*float64{mr.Float64Ptr(2), mr.Float64Ptr(0)},
					Dependencies:      []int64{0, 0},
					DependencyIndexes: []int{0, 0},
					Status:            "processing",
				},
			},
		}
		exprID, _, _ := exprRepo.CreateExpression(ctx, setupTx, expr)

		if err := setupTx.Commit(); err != nil {
			t.Fatal(err)
		}

		testTx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer testTx.Rollback()

		errorMsg := "division by zero"
		taskCompleted := &models.TaskCompleted{
			ID:         1,
			Expression: exprID,
			Error:      errorMsg,
		}

		err, code := manager.CompleteTask(ctx, taskCompleted)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)

		updatedExpr, _, _ := exprRepo.ReadExpressionByID(ctx, testTx, exprID)
		assert.Equal(t, "error", updatedExpr.Status)
		assert.Equal(t, errorMsg, updatedExpr.Error)

		tasks, _, _ := taskRepo.ReadTasksByExpressionID(ctx, testTx, exprID)
		assert.Empty(t, tasks)

		if err = testTx.Commit(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("not all tasks completed", func(t *testing.T) {
		if clearTestDatabase(db) != nil {
			t.Fatal(err)
		}

		setupTx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}

		expr := &models.Expression{
			ID:               1,
			ExpressionString: "2 + 3 * 4",
			UserID:           1,
			Status:           "processing",
			Tasks: []*models.Task{
				{
					Operation:         "*",
					Args:              []*float64{mr.Float64Ptr(3), mr.Float64Ptr(4)},
					Dependencies:      []int64{0, 0},
					DependencyIndexes: []int{0, 0},
					Status:            "processing",
				},
				{
					Operation:         "+",
					Args:              []*float64{mr.Float64Ptr(2), nil},
					Dependencies:      []int64{0, 0},
					DependencyIndexes: []int{0, 1},
					Status:            "processing",
				},
			},
		}

		exprID, _, _ := exprRepo.CreateExpression(ctx, setupTx, expr)

		if err := setupTx.Commit(); err != nil {
			t.Fatal(err)
		}

		testTx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer testTx.Rollback()

		_, _, _ = manager.ReadTask(ctx)

		if err = testTx.Commit(); err != nil {
			t.Fatal(err)
		}

		testTx, err = db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer testTx.Rollback()

		taskCompleted := &models.TaskCompleted{
			ID:         1,
			Expression: exprID,
			Result:     12,
		}

		err, code := manager.CompleteTask(ctx, taskCompleted)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)

		updatedExpr, _, _ := exprRepo.ReadExpressionByID(ctx, testTx, exprID)
		assert.Equal(t, "processing", updatedExpr.Status)
		assert.Nil(t, updatedExpr.Result)

		tasks, _, _ := taskRepo.ReadTasksByExpressionID(ctx, testTx, exprID)
		assert.Len(t, tasks, 2)

		if err = testTx.Commit(); err != nil {
			t.Fatalf("не удалось закоммитить транзакцию: %v", err)
		}
	})
}

func setupTestDatabase(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE expressions(
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			user_id INTEGER NOT NULL,
			expression_string TEXT NOT NULL,
			status TEXT CHECK(status IN ('pending', 'processing', 'completed', 'error')) DEFAULT 'pending',
			result REAL,
			error TEXT DEFAULT ''
		);`); err != nil {
		return err
	}
	if _, err := db.Exec(`
		CREATE TABLE tasks(
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			expression_id INTEGER NOT NULL,
			operation TEXT NOT NULL CHECK(operation IN ('+', '-', '*', '/', '^', 'u-')),
		    result REAL,
			status TEXT CHECK(status IN ('pending', 'processing', 'completed', 'error')) DEFAULT 'pending',
		    
			FOREIGN KEY (expression_id) REFERENCES expressions(id) ON DELETE CASCADE
		);`); err != nil {
		return err
	}
	if _, err := db.Exec(`
		CREATE TABLE task_args (
			task_id INTEGER PRIMARY KEY NOT NULL,
			first REAL,
			second REAL,
			
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
		);`); err != nil {
		return err
	}
	if _, err := db.Exec(`
		CREATE TABLE task_deps (
			task_id INTEGER PRIMARY KEY NOT NULL,
			first INTEGER,
			second INTEGER,
			
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
		);`); err != nil {
		return err
	}
	return nil
}

func clearTestDatabase(db *sql.DB) error {
	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return err
	}

	tables := []string{
		"task_deps",
		"task_args",
		"tasks",
		"expressions",
	}

	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			return err
		}
	}

	if _, err := db.Exec("DELETE FROM sqlite_sequence"); err != nil {
		return err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return err
	}

	return nil
}
