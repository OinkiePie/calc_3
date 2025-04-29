package middlewares_test

//
//import (
//	"io"
//	"log"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//
//	"github.com/OinkiePie/calc_3/config"
//	"github.com/OinkiePie/calc_3/orchestrator/internal/middlewares"
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
//func TestDisbledAuthorization(t *testing.T) {
//	// Создаем middleware с тестовыми параметрами
//	middleware := middlewares.NewOrchestratorMiddlewares("", "", []string{})
//
//	// Создаем фиктивный обработчик, который возвращает 200 OK
//	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK)
//	})
//
//	// Обертываем фиктивный обработчик в middleware
//	handler := middleware.EnableAuthorization(nextHandler)
//
//	t.Run("Successful disabled auth", func(t *testing.T) {
//		req := httptest.NewRequest("GET", "/", nil)
//
//		rr := httptest.NewRecorder()
//		handler.ServeHTTP(rr, req)
//
//		assert.Equal(t, http.StatusOK, rr.Code)
//	})
//
//	t.Run("Successful enabled auth with header", func(t *testing.T) {
//		req := httptest.NewRequest("GET", "/", nil)
//		req.Header.Set("Authorization", "Bearer mySecretApiKey")
//
//		rr := httptest.NewRecorder()
//		handler.ServeHTTP(rr, req)
//
//		assert.Equal(t, http.StatusOK, rr.Code)
//	})
//}
//
//func TestEnabledAuthorization(t *testing.T) {
//	// Создаем middleware с тестовыми параметрами
//	middleware := middlewares.NewOrchestratorMiddlewares("Bearer ", "mySecretApiKey", []string{})
//
//	// Создаем фиктивный обработчик, который возвращает 200 OK
//	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK)
//	})
//
//	// Обертываем фиктивный обработчик в middleware
//	handler := middleware.EnableAuthorization(nextHandler)
//
//	t.Run("Successful enabled auth", func(t *testing.T) {
//		req := httptest.NewRequest("GET", "/", nil)
//		req.Header.Set("Authorization", "Bearer mySecretApiKey")
//
//		rr := httptest.NewRecorder()
//		handler.ServeHTTP(rr, req)
//
//		assert.Equal(t, http.StatusOK, rr.Code)
//	})
//
//	t.Run("Отсутствует заголовок Authorization", func(t *testing.T) {
//		req := httptest.NewRequest("GET", "/", nil)
//
//		rr := httptest.NewRecorder()
//		handler.ServeHTTP(rr, req)
//
//		assert.Equal(t, http.StatusUnauthorized, rr.Code)
//		assert.Contains(t, rr.Body.String(), "Unauthorized: Missing authorization header")
//	})
//
//	t.Run("Некорректный формат заголовка Authorization", func(t *testing.T) {
//		req := httptest.NewRequest("GET", "/", nil)
//		req.Header.Set("Authorization", "плахойфарматбебебе")
//
//		rr := httptest.NewRecorder()
//		handler.ServeHTTP(rr, req)
//
//		assert.Equal(t, http.StatusUnauthorized, rr.Code)
//		assert.Contains(t, rr.Body.String(), "Unauthorized: Invalid authorization header format")
//	})
//
//	t.Run("Пустой API-ключ", func(t *testing.T) {
//		req := httptest.NewRequest("GET", "/", nil)
//		req.Header.Set("Authorization", "Bearer ")
//
//		rr := httptest.NewRecorder()
//		handler.ServeHTTP(rr, req)
//
//		assert.Equal(t, http.StatusUnauthorized, rr.Code)
//		assert.Contains(t, rr.Body.String(), "Unauthorized: Empty API key")
//	})
//
//	t.Run("Неверный API-ключ", func(t *testing.T) {
//		req := httptest.NewRequest("GET", "/", nil)
//		req.Header.Set("Authorization", "Bearer ключверныйинфасоткапатамушагладиолус")
//
//		rr := httptest.NewRecorder()
//		handler.ServeHTTP(rr, req)
//
//		assert.Equal(t, http.StatusUnauthorized, rr.Code)
//		assert.Contains(t, rr.Body.String(), "Unauthorized: Invalid API key")
//	})
//}
//
//func TestEnableCORS(t *testing.T) {
//	// Создаем middleware с тестовыми параметрами
//	middleware := middlewares.NewOrchestratorMiddlewares("", "", []string{"https://example.com"})
//
//	// Создаем фиктивный обработчик, который возвращает 200 OK
//	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK)
//	})
//
//	// Обертываем фиктивный обработчик в middleware
//	handler := middleware.EnableCORS(nextHandler)
//
//	t.Run("Запрос с разрешённым Origin", func(t *testing.T) {
//		req := httptest.NewRequest("GET", "/", nil)
//		req.Header.Set("Origin", "https://example.com")
//
//		rr := httptest.NewRecorder()
//		handler.ServeHTTP(rr, req)
//
//		assert.Equal(t, http.StatusOK, rr.Code)
//		assert.Equal(t, "https://example.com", rr.Header().Get("Access-Control-Allow-Origin"))
//		assert.Equal(t, "GET, POST, OPTIONS, PUT, DELETE", rr.Header().Get("Access-Control-Allow-Methods"))
//		assert.Equal(t, "Accept, Content-Type, Content-Length, Authorization", rr.Header().Get("Access-Control-Allow-Headers"))
//	})
//
//	t.Run("Запрос с запрещённым Origin", func(t *testing.T) {
//		req := httptest.NewRequest("GET", "/", nil)
//		req.Header.Set("Origin", "https://t.me/artrubadur")
//
//		rr := httptest.NewRecorder()
//		handler.ServeHTTP(rr, req)
//
//		assert.Equal(t, http.StatusOK, rr.Code)
//		assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
//	})
//
//	t.Run("Запрос без заголовка Origin", func(t *testing.T) {
//		req := httptest.NewRequest("GET", "/", nil)
//
//		rr := httptest.NewRecorder()
//		handler.ServeHTTP(rr, req)
//
//		assert.Equal(t, http.StatusOK, rr.Code)
//		assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
//	})
//}
