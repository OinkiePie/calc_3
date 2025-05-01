package router

import (
	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/orchestrator/internal/handlers"
	"github.com/OinkiePie/calc_3/orchestrator/internal/middlewares"
	"github.com/OinkiePie/calc_3/orchestrator/internal/providers"
	"github.com/gorilla/mux"
)

// NewOrchestratorRouter создает и настраивает маршрутизатор для Orchestrator API.
// Разделяет конечные точки на публичные, защищенные и внутренние.
//
// Args:
//
//	provider: *providers.Providers - Контейнер зависимостей приложения
//
// Returns:
//
//	*mux.Router - Настроенный маршрутизатор с зарегистрированными обработчиками
//
// Маршруты:
//
//	Публичные:
//	    POST /api/register - Регистрация пользователя
//	    POST /api/login - Аутентификация пользователя
//
//	Защищенные (требуют JWT):
//	    POST /api/p/delete - Удаление пользователя
//	    POST /api/p/calculate - Добавление выражения
//	    GET /api/p/expressions - Получение списка выражений
//	    GET /api/p/expressions/{id} - Получение выражения по ID
//
// Middleware:
//
//	EnableCORS - Для всех запросов
//	EnableAuth - Только для защищенных маршрутов
func NewOrchestratorRouter(provider *providers.Providers) *mux.Router {
	handler := handlers.NewOrchestratorHandlers(provider.UserManager, provider.ExprManager, provider.JWTManager)
	middleware := middlewares.NewOrchestratorMiddlewares(config.Cfg.Middleware.AllowOrigin, provider.UserManager, provider.JWTManager)
	router := mux.NewRouter()

	router.Use(middleware.EnableCORS)
	// API endpoints (внешние конечные точки, доступные клиентам)
	router.HandleFunc("/api/register", handler.RegisterUserHandler).Methods("POST")
	router.HandleFunc("/api/login", handler.LoginUserHandler).Methods("POST")

	// Protected API endpoints (внешние конечные точки, доступные только клиентам с авторизацией)
	authRouter := router.PathPrefix("/api/p/").Subrouter()
	authRouter.Use(middleware.EnableAuth)

	authRouter.HandleFunc("/logout", handler.LogoutUserHandler).Methods("GET")
	authRouter.HandleFunc("/delete", handler.DeleteUserHandler).Methods("POST")
	authRouter.HandleFunc("/calculate", handler.AddExpressionHandler).Methods("POST")
	authRouter.HandleFunc("/expressions", handler.GetExpressionsHandler).Methods("GET")
	authRouter.HandleFunc("/expressions/{id}", handler.GetExpressionHandler).Methods("GET")

	return router
}
