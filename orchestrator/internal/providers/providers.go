package providers

import (
	"context"
	"github.com/OinkiePie/calc_3/orchestrator/internal/managers"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories/expressions_repository"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories/session_repository"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories/tasks_repository"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories/user_repository"
	"github.com/OinkiePie/calc_3/pkg/database"
	"github.com/OinkiePie/calc_3/pkg/jwt_manager"
)

// Providers содержит все зависимости (репозитории и менеджеры) приложения.
// Используется для централизованного управления зависимостями и внедрения их в обработчики.
type Providers struct {
	SessionRepo repositories.SessionRepositoryInterface
	UserRepo    repositories.UserRepositoryInterface        // Репозиторий пользователей
	UserManager managers.UserManagerInterface               // Менеджер пользователей
	ArgsRepo    repositories.TasksArgsRepositoryInterface   // Репозиторий аргументов задач
	DepsRepo    repositories.TasksDepsRepositoryInterface   // Репозиторий зависимостей задач
	TaskRepo    repositories.TasksRepositoryInterface       // Репозиторий задач
	ExprRepo    repositories.ExpressionsRepositoryInterface // Репозиторий выражений
	ExprManager managers.ExpressionManagerInterface         // Менеджер выражений
	JWTManager  jwt_manager.JWTManagerInterface             // Менеджер JWT-токенов
	DB          *database.DataBase                          // Подключение к базе данных
}

// NewProviders создает и инициализирует все зависимости приложения.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения
//	dbPath: string - Путь к файлу базы данных
//	jwtKey: string - Ключ для подписи JWT-токенов
//
// Returns:
//
//	*Providers - Инициализированный контейнер зависимостей
//	error - Ошибка инициализации (например, проблемы с подключением к БД)
func NewProviders(ctx context.Context, dbPath string, jwtKey string) (*Providers, error) {

	db, err := database.NewDB(ctx, dbPath)
	if err != nil {
		return nil, err
	}

	jwtManager := jwt_manager.NewJWTManager(jwtKey)

	sessionRepo := session_repository.NewSessionRepository(db.DB)
	userRepo := user_repository.NewUserRepository(db.DB)
	userManager := managers.NewUserManager(db.DB, sessionRepo, userRepo, jwtManager)

	argsRepo := tasks_repository.NewTaskArgsRepository(db.DB)
	depsRepo := tasks_repository.NewTaskDepsRepository(db.DB)
	taskRepo := tasks_repository.NewTasksRepository(db.DB, depsRepo, argsRepo)
	exprRepo := expressions_repository.NewExpressionsRepository(db.DB, taskRepo)
	taskManager := managers.NewExpressionManager(db.DB, exprRepo, taskRepo)

	return &Providers{
		SessionRepo: sessionRepo,
		UserRepo:    userRepo,
		UserManager: userManager,
		ArgsRepo:    argsRepo,
		DepsRepo:    depsRepo,
		TaskRepo:    taskRepo,
		ExprRepo:    exprRepo,
		ExprManager: taskManager,
		JWTManager:  jwtManager,
		DB:          db,
	}, nil
}
