package middlewares

import (
	"github.com/OinkiePie/calc_3/orchestrator/internal/managers"
	"github.com/OinkiePie/calc_3/pkg/jwt_manager"
	"net/http"
	"strings"
)

// Middleware содержит middleware-функции для обработки HTTP-запросов.
type Middleware struct {
	allowOrigin []string                        // Список разрешенных источников для CORS
	allAllowed  bool                            // Флаг указывающий на разрешение всех источников
	userManager managers.UserManagerInterface   // Менеджер пользователей
	jwtManager  jwt_manager.JWTManagerInterface // Менеджер для работы с JWT-токенами
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
func NewOrchestratorMiddlewares(allowOrigin []string, userManager managers.UserManagerInterface, jwtManager jwt_manager.JWTManagerInterface) *Middleware {
	allAllowed := false
	for _, allowedOrigin := range allowOrigin {
		if allowedOrigin == "*" {
			allAllowed = true
			break
		}
	}
	return &Middleware{
		allowOrigin: allowOrigin,
		allAllowed:  allAllowed,
		userManager: userManager,
		jwtManager:  jwtManager,
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
			http.Error(w, "Отсутствует заголовок Authorization", http.StatusUnauthorized)
			return
		}

		// Проверяем, что заголовок начинается с префикса (например, "Bearer ").
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Неверный формат заголовка Authorization", http.StatusUnauthorized)
			return
		}

		// Извлекаем API-ключ из заголовка.
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Проверяем, что API-ключ не пустой.
		if token == "" {
			http.Error(w, "Пустой ключ авторизации", http.StatusUnauthorized)
			return
		}

		claims, err := m.jwtManager.Validate(token)

		if err != nil {
			if claims.JWTID != "" {
				err, _ := m.userManager.Logout(r.Context(), claims.JWTID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
			}
			http.Error(w, "Сессия была завершена или истекла", http.StatusUnauthorized)
			return
		}

		if err, exists := m.userManager.SessionExists(r.Context(), claims.JWTID); !exists {
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			}
			http.Error(w, "Сессия была завершена или истекла", http.StatusUnauthorized)
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

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}

		if origin != "" {
			allowed := m.allAllowed
			if !allowed {
				for _, allowedOrigin := range m.allowOrigin {
					if strings.EqualFold(origin, allowedOrigin) {
						allowed = true
						break
					}
				}
			}

			if allowed {
				if m.allAllowed {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
			}
		}

		next.ServeHTTP(w, r)
	})
}
