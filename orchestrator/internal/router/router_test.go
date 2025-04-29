package router_test

//
//import (
//	"io"
//	"log"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//
//	"github.com/OinkiePie/calc_3/config"
//	"github.com/OinkiePie/calc_3/orchestrator/internal/router"
//	"github.com/OinkiePie/calc_3/pkg/logger"
//	"github.com/stretchr/testify/assert"
//)
//
//func init() {
//	// Отключаем выводы и инициализируем конфиг
//	log.SetOutput(io.Discard)
//	config.InitConfig()
//	logger.InitLogger(logger.Options{Level: 6})
//}
//
//// TestNewOrchestratorRouter тестирует создание, настройку роутера и маршруты (без аутентификации).
//func TestNewOrchestratorRouterUnauthorized(t *testing.T) {
//	config.Cfg.Middleware.ApiKeyPrefix = "Bearer "
//	config.Cfg.Middleware.Authorization = "Skibidi"
//
//	router := router.NewOrchestratorRouter()
//
//	testsGet := []struct {
//		method       string
//		path         string
//		expectedCode int
//	}{
//		{"POST", "/api/v1/calculate", http.StatusBadRequest}, // Пустое тело запроса
//		{"GET", "/api/v1/expressions", http.StatusOK},
//		{"GET", "/api/v1/expressions/1", http.StatusNotFound}, // Нет выражения с таким ID
//		{"GET", "/internal/task", http.StatusUnauthorized},
//		{"GET", "/internal/task/1", http.StatusUnauthorized},
//		{"POST", "/internal/task", http.StatusUnauthorized},
//	}
//
//	for _, tt := range testsGet {
//		req, err := http.NewRequest(tt.method, tt.path, nil)
//		assert.NoError(t, err)
//
//		rr := httptest.NewRecorder()
//		router.ServeHTTP(rr, req)
//
//		assert.Equal(t, tt.expectedCode, rr.Code, "для %s %s ожидался статус %d, получен %d", tt.method, tt.path, tt.expectedCode, rr.Code)
//	}
//}
//
//// TestNewOrchestratorRouterWithAuth тестирует маршруты с аутентификацией.
//func TestNewOrchestratorRouterAuthorized(t *testing.T) {
//	config.Cfg.Middleware.ApiKeyPrefix = "Bearer "
//	config.Cfg.Middleware.Authorization = "2KluchaAvtorizaciiMneIli2Drugomu"
//
//	router := router.NewOrchestratorRouter()
//
//	authHeader := config.Cfg.Middleware.ApiKeyPrefix + config.Cfg.Middleware.Authorization
//	tests := []struct {
//		method       string
//		path         string
//		expectedCode int
//	}{
//		{"GET", "/internal/task", http.StatusNotFound},    // Авторизацию прошел, но задач нет
//		{"POST", "/internal/task", http.StatusBadRequest}, // Авторизацию прошел, но зпустое тело запроса
//		{"GET", "/internal/task/1", http.StatusNotFound},  // Авторизацию прошел, выражения не существует
//	}
//
//	for _, tt := range tests {
//		req, err := http.NewRequest(tt.method, tt.path, nil)
//		assert.NoError(t, err)
//
//		req.Header.Set("Authorization", authHeader)
//		rr := httptest.NewRecorder()
//		router.ServeHTTP(rr, req)
//
//		assert.Equal(t, tt.expectedCode, rr.Code, "для %s %s с авторизацией ожидался статус %d, получен %d", tt.method, tt.path, tt.expectedCode, rr.Code)
//	}
//}
