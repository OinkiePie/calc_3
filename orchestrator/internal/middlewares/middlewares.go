package middlewares

import (
	"github.com/OinkiePie/calc_3/pkg/jwt"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"net/http"
	"strings"
)

// Middleware содержит middleware-функции для обработки HTTP-запросов.
type Middleware struct {
	allowOrigin []string        // Список разрешенных источников для CORS
	jwtManager  *jwt.JWTManager // Менеджер для работы с JWT-токенами
}

// NewOrchestratorMiddlewares создает новый экземпляр Middleware.
//
// Args:
//
//	allowOrigin: []string - Список разрешенных доменов для CORS
//	jwtm: *jwt.JWTManager - Менеджер JWT-токенов
//
// Returns:
//
//	*Middleware - Новый экземпляр Middleware
func NewOrchestratorMiddlewares(allowOrigin []string, jwtm *jwt.JWTManager) *Middleware {
	return &Middleware{
		allowOrigin: allowOrigin,
		jwtManager:  jwtm,
	}
}

// EnableAuth проверяет JWT-токен в заголовке Authorization.
// Возвращает 401 если токен отсутствует или невалиден.
//
// Args:
//
//	next: http.Handler - Следующий обработчик в цепочке
//
// Returns:
//
//	http.Handler - Обработчик с проверкой авторизации
func (m *Middleware) EnableAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Получаем API-ключ из заголовка Authorization.
		authHeader := r.Header.Get("Authorization")

		// Проверяем, что заголовок Authorization присутствует.
		if authHeader == "" {
			logger.Log.Debugf("Отсутствует заголовок Authorization")
			http.Error(w, "Unauthorized: Missing authorization header", http.StatusUnauthorized) // 401
			return
		}

		// Проверяем, что заголовок начинается с префикса (например, "Bearer ").
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Log.Debugf("Неверный формат заголовка Authorization")
			http.Error(w, "Unauthorized: Invalid authorization header format", http.StatusUnauthorized) // 401
			return
		}

		// Извлекаем API-ключ из заголовка.
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Проверяем, что API-ключ не пустой.
		if token == "" {
			logger.Log.Debugf("Пустой API-ключ")
			http.Error(w, "Unauthorized: Empty API key", http.StatusUnauthorized) // 401
			return
		}

		if _, err := m.jwtManager.Validate(token); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized) // 401
			return
		}

		next.ServeHTTP(w, r)
	})
}

// EnableCORS добавляет CORS-заголовки для разрешенных источников.
//
// Args:
//
//	next: http.Handler - Следующий обработчик в цепочке
//
// Returns:
//
//	http.Handler - Обработчик с CORS-политиками
func (m *Middleware) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			// Проверяем, есть ли origin в списке разрешенных
			allowed := false
			for _, allowedOrigin := range m.allowOrigin {
				if strings.EqualFold(origin, allowedOrigin) { //Сравнение без учета регистра
					allowed = true
					break
				}
			}
			if allowed {
				// Если origin разрешен, устанавливаем заголовок Access-Control-Allow-Origin
				w.Header().Set("Access-Control-Allow-Origin", origin)
				// Дополнительные заголовки CORS
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
			}
		}

		next.ServeHTTP(w, r)
	})
}
