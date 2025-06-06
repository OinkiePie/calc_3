package managers

import (
	"context"
	"github.com/OinkiePie/calc_3/pkg/models"
)

type UserManagerInterface interface {
	// Register регистрирует нового пользователя в системе.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения.
	//	login: string - Логин пользователя.
	//	password: string - Пароль пользователя (в открытом виде).
	//
	// Returns:
	//
	//	int64 - ID зарегистрированного пользователя.
	//	error - Ошибка выполнения.
	//	int - HTTP статус код:
	//		- 201 Created при успешном создании
	//		- 409 Conflict при дубликате логина
	//		- 500 Internal Server Error при других ошибках
	Register(ctx context.Context, login, password string) (int64, error, int)

	// Login выполняет аутентификацию пользователя и генерирует JWT-токен.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения.
	//	login: string - Логин пользователя.
	//	password: string - Пароль пользователя (в открытом виде).
	//
	// Returns:
	//
	//	string - Сгенерированный JWT-токен.
	//	int64 - ID аутентифицированного пользователя.
	//	error - Ошибка выполнения.
	//	int - HTTP статус код:
	//		- 200 OK при успешной аутентификации
	//		- 401 Unauthorized если пользователь не найден
	//		- 500 Internal Server Error при ошибках
	Login(ctx context.Context, login, password string) (string, int64, error, int)

	// Logout удаляет сессию из базы данных.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст для контроля выполнения запроса.
	//	jti: string - Идентификатор сессии.
	//
	// Returns:
	//
	//	*models.Session - Указатель на сессию.
	//	error - Ошибка выполнения операции.
	//	int - HTTP-статус код результата операции:
	//		- 200 OK при успешном получении
	//		- 500 Internal Server Error при ошибках
	Logout(ctx context.Context, jti string) (error, int)

	// Delete удаляет учетную запись пользователя после проверки пароля.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения.
	//	login: string - Логин пользователя.
	//	password: string - Пароль пользователя (в открытом виде).
	//
	// Returns:
	//
	//	int64 - ID удаленного пользователя.
	//	error - Ошибка выполнения.
	//	int - HTTP статус код:
	//		- 200 OK при успешном удалении
	//		- 401 Unauthorized при несовпадении паролей
	//		- 500 Internal Server Error при ошибках
	Delete(ctx context.Context, login, password string) (int64, error, int)

	// SessionExists проверяет существование сессии.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения.
	//	jti: string - Идентификатор сессии.
	//
	// Returns:
	//
	//	error - Ошибка выполнения.
	//	bool - Индикатор существования.
	SessionExists(ctx context.Context, jti string) (error, bool)
}

type ExpressionManagerInterface interface {
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
	AddExpression(ctx context.Context, expressionString string, claims int64) (int64, error, int)

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
	ReadExpressions(ctx context.Context, id int64) ([]*models.Expression, error, int)

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
	ReadExpression(ctx context.Context, id int64) (*models.Expression, error, int)

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
	//		- 500 Internal Server Error при ошибках
	ReadTask(ctx context.Context) (*models.Task, error, int)

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
	//		- 500 Internal Server Error при ошибках
	CompleteTask(ctx context.Context, taskCompleted *models.TaskCompleted) (error, int)
}
