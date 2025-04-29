package client_test

//
//import (
//	"encoding/json"
//	"fmt"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//
//	"github.com/OinkiePie/calc_3/agent/internal/client"
//	"github.com/OinkiePie/calc_3/pkg/models"
//	"github.com/stretchr/testify/assert"
//)
//
//func TestGetTask(t *testing.T) {
//	expectedNil := &models.TaskResponse{}
//	expectedNil = nil
//
//	// 1. Успешный запрос (200 OK)
//	t.Run("GET Success", func(t *testing.T) {
//		expectedTask := &models.TaskResponse{
//			ID:             "task-id-777",
//			Operation:      "+",
//			Args:           []*float64{nil, nil},
//			Operation_time: 100,
//			Expression:     "expr-id-2228",
//		}
//
//		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			if r.Method != http.MethodGet {
//				t.Errorf("Ожидался метод GET, получено: %s", r.Method)
//			}
//			if r.Header.Get("Authorization") != "test_token" {
//				t.Errorf("Неверный заголовок Authorization: %s", r.Header.Get("Authorization"))
//			}
//
//			w.Header().Set("Content-Type", "application/json")
//			w.WriteHeader(http.StatusOK)
//			json.NewEncoder(w).Encode(expectedTask)
//		}))
//
//		apiClient := client.NewAPIClient(server.URL, "test_token", &http.Client{})
//		task, err := apiClient.GetTask()
//
//		assert.NoError(t, err)
//		assert.Equal(t, expectedTask, task, fmt.Sprintf("Неверные данные задачи. Ожидалось: %+v, получено: %+v", expectedTask, task))
//	})
//
//	// 2. Задача не найдена (404 Not Found)
//	t.Run("GET NotFound", func(t *testing.T) {
//
//		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusNotFound)
//		}))
//
//		apiClient := client.NewAPIClient(server.URL, "test_token", &http.Client{})
//		task, err := apiClient.GetTask()
//
//		assert.NoError(t, err)
//		assert.Equal(t, expectedNil, task, fmt.Sprintf("Ожидалось nil, получено: %+v", task))
//	})
//
//	// 3. Ошибка сервера (500 Internal Server Error)
//	t.Run("GET ServerError", func(t *testing.T) {
//		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusInternalServerError)
//		}))
//
//		apiClient := client.NewAPIClient(server.URL, "test_token", &http.Client{})
//		task, err := apiClient.GetTask()
//
//		assert.Error(t, err)
//		assert.Equal(t, expectedNil, task, fmt.Sprintf("Ожидалось nil, получено: %+v", task))
//	})
//
//}
//func TestCompleteTask(t *testing.T) {
//	completedTask := models.TaskCompleted{
//		ID:         "task-id-666",
//		Expression: "expr-id-666",
//		Result:     4,
//		Error:      "",
//	}
//	// 1.  Успешный запрос (200 OK)
//	t.Run("POST Success", func(t *testing.T) {
//		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			if r.Method != http.MethodPost {
//				t.Errorf("Ожидался метод POST, получено: %s", r.Method)
//			}
//
//			if r.Header.Get("Authorization") != "test_token" {
//				t.Errorf("Неверный заголовок Authorization: %s", r.Header.Get("Authorization"))
//			}
//
//			w.WriteHeader(http.StatusOK)
//		}))
//
//		apiClient := client.NewAPIClient(server.URL, "test_token", &http.Client{})
//		err := apiClient.CompleteTask(completedTask)
//
//		assert.NoError(t, err)
//	})
//
//	// 2. Плохой JSON (422 Unprocessable Entity)
//	t.Run("POST Unprocessable", func(t *testing.T) {
//		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusUnprocessableEntity)
//			json.NewEncoder(w).Encode(map[string]string{"error": "не удалось декодировать JSON"})
//		}))
//
//		apiClient := client.NewAPIClient(server.URL, "test_token", &http.Client{})
//		err := apiClient.CompleteTask(completedTask)
//
//		assert.Error(t, err)
//	})
//
//	// 3. Плохой запрос (422 Unprocessable Entity)
//	t.Run("POST Unprocessable", func(t *testing.T) {
//		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusUnprocessableEntity)
//			json.NewEncoder(w).Encode(map[string]string{"error": "не удалось прочитать тело запроса"})
//		}))
//
//		apiClient := client.NewAPIClient(server.URL, "test_token", &http.Client{})
//		err := apiClient.CompleteTask(completedTask)
//
//		assert.Error(t, err)
//	})
//
//	// 4. Плохой JSON (500 Internal Server Error)
//	t.Run("GET Internal Error", func(t *testing.T) {
//		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusInternalServerError)
//			json.NewEncoder(w).Encode(map[string]string{"error": "не удалось декодировать JSON"})
//		}))
//
//		apiClient := client.NewAPIClient(server.URL, "test_token", &http.Client{})
//		err := apiClient.CompleteTask(completedTask)
//
//		assert.Error(t, err)
//	})
//
//	// 5. Задача не найдена (404 Not Found)
//	t.Run("GET Internal Error", func(t *testing.T) {
//		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusNotFound)
//			json.NewEncoder(w).Encode(map[string]string{"error": "задача не найдена"})
//		}))
//
//		apiClient := client.NewAPIClient(server.URL, "test_token", &http.Client{})
//		err := apiClient.CompleteTask(completedTask)
//
//		assert.Error(t, err)
//	})
//}
