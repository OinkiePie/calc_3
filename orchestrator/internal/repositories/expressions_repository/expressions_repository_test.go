package expressions_repository_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	m "github.com/OinkiePie/calc_3/orchestrator/internal/repositories"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories/expressions_repository"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
)

func TestCreateExpression_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	taskRepoMock := new(m.MockTasksRepository)

	repo := expressions_repository.NewExpressionsRepository(db, taskRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expr := &models.Expression{
		UserID:           1,
		ExpressionString: "2+2",
		Tasks: []*models.Task{
			{
				Operation:         "+",
				Args:              []*float64{m.Float64Ptr(2), m.Float64Ptr(2)},
				DependencyIndexes: []int{},
			},
		},
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	sqlMock.ExpectQuery(`INSERT INTO expressions`).
		WithArgs(expr.UserID, expr.ExpressionString).
		WillReturnRows(rows)

	taskRepoMock.On("CreateTask", mock.Anything, tx, expr.Tasks[0]).
		Return(int64(1), nil, http.StatusCreated)

	taskRepoMock.On("UpdateTaskDependencies", mock.Anything, tx, expr.Tasks[0]).
		Return(nil, http.StatusOK)

	id, err, status := repo.CreateExpression(context.Background(), tx, expr)

	assert.Equal(t, int64(1), id)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	taskRepoMock.AssertExpectations(t)
}

func TestCreateExpression_InsertError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(db, taskRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expr := &models.Expression{
		UserID:           1,
		ExpressionString: "2+2",
		Tasks:            []*models.Task{},
	}

	sqlMock.ExpectQuery(`INSERT INTO expressions`).
		WithArgs(expr.UserID, expr.ExpressionString).
		WillReturnError(fmt.Errorf("database error"))

	id, err, status := repo.CreateExpression(context.Background(), tx, expr)

	assert.Equal(t, int64(0), id)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCreateExpression_TaskCreationError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(db, taskRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expr := &models.Expression{
		UserID:           1,
		ExpressionString: "2+2",
		Tasks: []*models.Task{
			{
				Operation:         "+",
				Args:              []*float64{m.Float64Ptr(2), m.Float64Ptr(2)},
				DependencyIndexes: []int{},
			},
		},
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	sqlMock.ExpectQuery(`INSERT INTO expressions`).
		WithArgs(expr.UserID, expr.ExpressionString).
		WillReturnRows(rows)

	taskRepoMock.On("CreateTask", mock.Anything, tx, expr.Tasks[0]).
		Return(int64(0), fmt.Errorf("task creation error"), http.StatusInternalServerError)

	id, err, status := repo.CreateExpression(context.Background(), tx, expr)

	assert.Equal(t, int64(0), id)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	taskRepoMock.AssertExpectations(t)
}

func TestCreateExpression_DependenciesUpdateError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(db, taskRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expr := &models.Expression{
		UserID:           1,
		ExpressionString: "2+2",
		Tasks: []*models.Task{
			{
				Operation:         "+",
				Args:              []*float64{m.Float64Ptr(2), m.Float64Ptr(2)},
				DependencyIndexes: []int{},
				ID:                1,
			},
		},
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	sqlMock.ExpectQuery(`INSERT INTO expressions`).
		WithArgs(expr.UserID, expr.ExpressionString).
		WillReturnRows(rows)

	taskRepoMock.On("CreateTask", mock.Anything, tx, expr.Tasks[0]).
		Return(int64(1), nil, http.StatusCreated)

	taskRepoMock.On("UpdateTaskDependencies", mock.Anything, tx, expr.Tasks[0]).
		Return(fmt.Errorf("deps update error"), http.StatusInternalServerError)

	id, err, status := repo.CreateExpression(context.Background(), tx, expr)

	assert.Equal(t, int64(0), id)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	taskRepoMock.AssertExpectations(t)
}

func TestReadExpressionByID_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(db, taskRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expectedExpr := models.Expression{
		ID:               1,
		Status:           "completed",
		Result:           m.Float64Ptr(4),
		ExpressionString: "2+2",
		UserID:           1,
	}

	rows := sqlmock.NewRows([]string{"id", "status", "result", "expression_string", "error", "user_id"}).
		AddRow(expectedExpr.ID, expectedExpr.Status, expectedExpr.Result,
			expectedExpr.ExpressionString, "", expectedExpr.UserID)

	sqlMock.ExpectQuery(`SELECT.*FROM expressions WHERE id = \?`).
		WithArgs(expectedExpr.ID).
		WillReturnRows(rows)

	expectedTasks := []*models.Task{
		{ID: 1, Expression: expectedExpr.ID, Operation: "+"},
	}
	taskRepoMock.On("ReadTasksByExpressionID", mock.Anything, tx, expectedExpr.ID).
		Return(expectedTasks, nil, http.StatusOK)

	expr, err, status := repo.ReadExpressionByID(context.Background(), tx, expectedExpr.ID)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, expectedExpr.ID, expr.ID)
	assert.Equal(t, expectedExpr.ExpressionString, expr.ExpressionString)
	assert.Len(t, expr.Tasks, 1)
	assert.Equal(t, expectedTasks[0].ID, expr.Tasks[0].ID)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	taskRepoMock.AssertExpectations(t)
}

func TestReadExpressionByID_NotFound(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(db, taskRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	sqlMock.ExpectQuery(`SELECT.* FROM expressions WHERE id = \?`).
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	expr, err, status := repo.ReadExpressionByID(context.Background(), tx, 1)

	assert.Nil(t, expr)
	assert.Error(t, err)
	assert.Equal(t, "выражение не найдено", err.Error())
	assert.Equal(t, http.StatusNotFound, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadExpressionByID_DBError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(db, taskRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	sqlMock.ExpectQuery(`SELECT.*FROM expressions WHERE id = \?`).
		WithArgs(1).
		WillReturnError(fmt.Errorf("database error"))

	expr, err, status := repo.ReadExpressionByID(context.Background(), tx, 1)

	assert.Nil(t, expr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось получить выражение")
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadExpressionByID_TasksError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(db, taskRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	rows := sqlmock.NewRows([]string{"id", "status", "result", "expression_string", "error", "user_id"}).
		AddRow(int64(1), "completed", 4, "2+2", "", int64(1))

	sqlMock.ExpectQuery(`SELECT.*FROM expressions WHERE id = \?`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	taskRepoMock.On("ReadTasksByExpressionID", mock.Anything, tx, int64(1)).
		Return(([]*models.Task)(nil), fmt.Errorf("tasks error"), http.StatusInternalServerError)

	expr, err, status := repo.ReadExpressionByID(context.Background(), tx, 1)

	assert.Nil(t, expr)
	assert.Error(t, err)
	assert.Equal(t, "tasks error", err.Error())
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	taskRepoMock.AssertExpectations(t)
}

func TestReadExpressionsByUserID_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(db, taskRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	userID := int64(1)
	expectedExpressions := []*models.Expression{
		{
			ID:               1,
			Status:           "completed",
			Result:           m.Float64Ptr(4),
			ExpressionString: "2+2",
			UserID:           userID,
		},
		{
			ID:               2,
			Status:           "processing",
			ExpressionString: "3*3",
			UserID:           userID,
		},
	}

	rows := sqlmock.NewRows([]string{"id", "status", "result", "expression_string", "error", "user_id"}).
		AddRow(expectedExpressions[0].ID, expectedExpressions[0].Status, expectedExpressions[0].Result,
			expectedExpressions[0].ExpressionString, "", expectedExpressions[0].UserID).
		AddRow(expectedExpressions[1].ID, expectedExpressions[1].Status, nil,
			expectedExpressions[1].ExpressionString, "", expectedExpressions[1].UserID)

	sqlMock.ExpectQuery(`SELECT.*FROM expressions WHERE user_id = \?`).
		WithArgs(userID).
		WillReturnRows(rows)

	taskRepoMock.On("ReadTasksByExpressionID", mock.Anything, tx, expectedExpressions[0].ID).
		Return([]*models.Task{{ID: 1, Operation: "+"}}, nil, http.StatusOK)
	taskRepoMock.On("ReadTasksByExpressionID", mock.Anything, tx, expectedExpressions[1].ID).
		Return([]*models.Task{{ID: 2, Operation: "*"}}, nil, http.StatusOK)

	expressions, err, status := repo.ReadExpressionsByUserID(context.Background(), tx, userID)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, expressions, 2)
	assert.Equal(t, expectedExpressions[0].ID, expressions[0].ID)
	assert.Equal(t, expectedExpressions[1].ID, expressions[1].ID)
	assert.Len(t, expressions[0].Tasks, 1)
	assert.Len(t, expressions[1].Tasks, 1)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
	taskRepoMock.AssertExpectations(t)
}

func TestReadExpressionsByUserID_NotFound(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(db, taskRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	userID := int64(1)

	rows := sqlmock.NewRows([]string{"id", "status", "result", "expression_string", "error", "user_id"})
	sqlMock.ExpectQuery(`SELECT.*FROM expressions WHERE user_id = \?`).
		WithArgs(userID).
		WillReturnRows(rows)

	expressions, err, status := repo.ReadExpressionsByUserID(context.Background(), tx, userID)

	assert.Nil(t, expressions)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("выражения пользователя №%d не найдены", userID), err.Error())
	assert.Equal(t, http.StatusNotFound, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadExpressionsByUserID_DBError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(db, taskRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	userID := int64(1)

	sqlMock.ExpectQuery(`SELECT.*FROM expressions WHERE user_id = \?`).
		WithArgs(userID).
		WillReturnError(fmt.Errorf("database error"))

	expressions, err, status := repo.ReadExpressionsByUserID(context.Background(), tx, userID)

	assert.Nil(t, expressions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось получить выражения")
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadExpressionsByUserID_RowsScanError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(db, taskRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	userID := int64(1)

	rows := sqlmock.NewRows([]string{"id", "status"}).AddRow(1, "completed")
	sqlMock.ExpectQuery(`SELECT.*FROM expressions WHERE user_id = \?`).
		WithArgs(userID).
		WillReturnRows(rows)

	expressions, err, status := repo.ReadExpressionsByUserID(context.Background(), tx, userID)

	assert.Nil(t, expressions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось прочитать выражение")
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestReadExpressionsByUserID_TasksError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(db, taskRepoMock)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	userID := int64(1)
	exprID := int64(1)

	rows := sqlmock.NewRows([]string{"id", "status", "result", "expression_string", "error", "user_id"}).
		AddRow(exprID, "completed", 4, "2+2", "", userID)

	sqlMock.ExpectQuery(`SELECT.*FROM expressions WHERE user_id = \?`).
		WithArgs(userID).
		WillReturnRows(rows)

	taskRepoMock.On("ReadTasksByExpressionID", mock.Anything, tx, exprID).
		Return(([]*models.Task)(nil), fmt.Errorf("tasks error"), http.StatusInternalServerError)

	expressions, err, status := repo.ReadExpressionsByUserID(context.Background(), tx, userID)

	assert.Nil(t, expressions)
	assert.Error(t, err)
	assert.Equal(t, "tasks error", err.Error())
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	taskRepoMock.AssertExpectations(t)
}

func TestReadExpressionTasks_Success(t *testing.T) {
	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(nil, taskRepoMock)

	expressionID := int64(1)
	expectedTasks := []*models.Task{
		{ID: 1, Expression: expressionID, Operation: "+"},
		{ID: 2, Expression: expressionID, Operation: "*"},
	}

	taskRepoMock.On("ReadTasksByExpressionID", mock.Anything, (*sql.Tx)(nil), expressionID).
		Return(expectedTasks, nil, http.StatusOK)

	tasks, err, status := repo.ReadExpressionTasks(context.Background(), nil, expressionID)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, tasks, 2)
	assert.Equal(t, expectedTasks[0].ID, tasks[0].ID)
	assert.Equal(t, expectedTasks[1].ID, tasks[1].ID)

	taskRepoMock.AssertExpectations(t)
}

func TestReadExpressionTasks_NotFound(t *testing.T) {
	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(nil, taskRepoMock)

	expressionID := int64(1)

	taskRepoMock.On("ReadTasksByExpressionID", mock.Anything, (*sql.Tx)(nil), expressionID).
		Return(([]*models.Task)(nil), errors.New("задачи не найдены"), http.StatusNotFound)

	tasks, err, status := repo.ReadExpressionTasks(context.Background(), nil, expressionID)

	assert.Nil(t, tasks)
	assert.Error(t, err)
	assert.Equal(t, "задачи не найдены", err.Error())
	assert.Equal(t, http.StatusNotFound, status)
	taskRepoMock.AssertExpectations(t)
}

func TestReadExpressionTasks_DBError(t *testing.T) {
	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(nil, taskRepoMock)

	expressionID := int64(1)

	taskRepoMock.On("ReadTasksByExpressionID", mock.Anything, (*sql.Tx)(nil), expressionID).
		Return(([]*models.Task)(nil), fmt.Errorf("ошибка базы данных"), http.StatusInternalServerError)

	tasks, err, status := repo.ReadExpressionTasks(context.Background(), nil, expressionID)

	assert.Nil(t, tasks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка базы данных")
	assert.Equal(t, http.StatusInternalServerError, status)
	taskRepoMock.AssertExpectations(t)
}

func TestReadExpressionTasks_EmptyResult(t *testing.T) {
	taskRepoMock := new(m.MockTasksRepository)
	repo := expressions_repository.NewExpressionsRepository(nil, taskRepoMock)

	expressionID := int64(1)

	taskRepoMock.On("ReadTasksByExpressionID", mock.Anything, (*sql.Tx)(nil), expressionID).
		Return([]*models.Task{}, nil, http.StatusOK)

	tasks, err, status := repo.ReadExpressionTasks(context.Background(), nil, expressionID)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Empty(t, tasks)
	taskRepoMock.AssertExpectations(t)
}

func TestUpdateExpressionStatus_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := expressions_repository.NewExpressionsRepository(db, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)
	newStatus := "completed"

	sqlMock.ExpectExec(`UPDATE expressions SET status = \? WHERE id = \?`).
		WithArgs(newStatus, expressionID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err, status := repo.UpdateExpressionStatus(context.Background(), tx, expressionID, newStatus)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateExpressionStatus_DBError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := expressions_repository.NewExpressionsRepository(db, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)
	newStatus := "completed"

	sqlMock.ExpectExec(`UPDATE expressions SET status = \? WHERE id = \?`).
		WithArgs(newStatus, expressionID).
		WillReturnError(fmt.Errorf("database error"))

	err, status := repo.UpdateExpressionStatus(context.Background(), tx, expressionID, newStatus)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось обновить статус выражения")
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateExpressionStatus_NoRowsAffected(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := expressions_repository.NewExpressionsRepository(db, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)
	newStatus := "completed"

	sqlMock.ExpectExec(`UPDATE expressions SET status = \? WHERE id = \?`).
		WithArgs(newStatus, expressionID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err, status := repo.UpdateExpressionStatus(context.Background(), tx, expressionID, newStatus)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateExpressionError_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := expressions_repository.NewExpressionsRepository(db, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)
	errorContent := "division by zero"

	sqlMock.ExpectExec(`UPDATE expressions SET error = \? WHERE id = \?`).
		WithArgs(errorContent, expressionID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err, status := repo.UpdateExpressionError(context.Background(), tx, expressionID, errorContent)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)

	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateExpressionError_DBError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := expressions_repository.NewExpressionsRepository(db, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)
	errorContent := "division by zero"

	sqlMock.ExpectExec(`UPDATE expressions SET error = \? WHERE id = \?`).
		WithArgs(errorContent, expressionID).
		WillReturnError(fmt.Errorf("database error"))

	err, status := repo.UpdateExpressionError(context.Background(), tx, expressionID, errorContent)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось обновить ошибку выражения")
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateExpressionError_EmptyError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := expressions_repository.NewExpressionsRepository(db, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)
	emptyError := ""

	sqlMock.ExpectExec(`UPDATE expressions SET error = \? WHERE id = \?`).
		WithArgs(emptyError, expressionID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err, status := repo.UpdateExpressionError(context.Background(), tx, expressionID, emptyError)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateExpressionError_NoRowsAffected(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := expressions_repository.NewExpressionsRepository(db, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)
	errorContent := "invalid syntax"

	sqlMock.ExpectExec(`UPDATE expressions SET error = \? WHERE id = \?`).
		WithArgs(errorContent, expressionID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err, status := repo.UpdateExpressionError(context.Background(), tx, expressionID, errorContent)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateExpressionResult_Success(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := expressions_repository.NewExpressionsRepository(db, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)
	result := 42.0

	sqlMock.ExpectExec(`UPDATE expressions SET result = \? WHERE id = \?`).
		WithArgs(result, expressionID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err, status := repo.UpdateExpressionResult(context.Background(), tx, expressionID, result)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateExpressionResult_DBError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := expressions_repository.NewExpressionsRepository(db, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)
	result := 42.0

	sqlMock.ExpectExec(`UPDATE expressions SET result = \? WHERE id = \?`).
		WithArgs(result, expressionID).
		WillReturnError(fmt.Errorf("database error"))

	err, status := repo.UpdateExpressionResult(context.Background(), tx, expressionID, result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось обновить результат выражения")
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateExpressionResult_NoRowsAffected(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := expressions_repository.NewExpressionsRepository(db, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)
	result := 42.0

	sqlMock.ExpectExec(`UPDATE expressions SET result = \? WHERE id = \?`).
		WithArgs(result, expressionID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err, status := repo.UpdateExpressionResult(context.Background(), tx, expressionID, result)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestUpdateExpressionResult_WithZeroValue(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := expressions_repository.NewExpressionsRepository(db, nil)

	sqlMock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Ошибка начала транзакции: %v", err)
	}

	expressionID := int64(1)
	zeroResult := 0.0

	sqlMock.ExpectExec(`UPDATE expressions SET result = \? WHERE id = \?`).
		WithArgs(zeroResult, expressionID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err, status := repo.UpdateExpressionResult(context.Background(), tx, expressionID, zeroResult)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
