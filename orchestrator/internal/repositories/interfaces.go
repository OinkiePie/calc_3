package repositories

import (
	"context"
	"database/sql"
	"github.com/OinkiePie/calc_3/pkg/models"
)

type UserRepositoryInterface interface {
	// CreateUser создает нового пользователя в базе данных с указанными учетными данными.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст для контроля выполнения запроса.
	//	login: string - Логин пользователя (должен быть уникальным).
	//	pas: string - Пароль пользователя в незашифрованном виде.
	//
	// Returns:
	//
	//	*models.User - Созданный объект пользователя с заполненным ID.
	//	error - Ошибка при нарушении уникальности логина или проблемах с БД.
	//	int - HTTP-статус код результата операции:
	//		- 201 Created при успешном создании
	//		- 409 Conflict при дубликате логина
	//		- 500 Internal Server Error при других ошибках
	CreateUser(ctx context.Context, login, pas string) (*models.User, error, int)

	// ReadUserByLogin получает пользователя из базы данных по его логину.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	login: string - Логин пользователя для поиска.
	//
	// Returns:
	//
	//	*models.User - Найденный пользователь.
	//	error - Ошибка выполнения запроса.
	//	int - HTTP-статус код:
	//		- 200 OK при успешном поиске
	//		- 401 Unauthorized если пользователь не найден
	//		- 500 Internal Server Error при ошибках
	ReadUserByLogin(ctx context.Context, login string) (*models.User, error, int)

	// DeleteUser удаляет пользователя из базы данных по указанному идентификатору.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	id: int64 - Идентификатор пользователя для удаления.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции
	//	int - HTTP-статус код:
	//		- 200 OK при успешном удалении
	//		- 500 Internal Server Error при ошибках
	DeleteUser(ctx context.Context, id int64) (error, int)
}

type SessionRepositoryInterface interface {
	// CreateSession создает новую сессию в базе данных.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст для контроля выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных
	//	jti: string - Идентификатор сессии.
	//	sub: int64 - Идентификатор пользователя.
	//	exp: int64 - Время истечения сессии.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции.
	//	int - HTTP-статус код результата операции:
	//		- 201 Created при успешном создании
	//		- 500 Internal Server Error при других ошибках
	CreateSession(ctx context.Context, tx *sql.Tx, jti string, sub, exp int64) (error, int)

	// ReadSession получает сессию из базы данных.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст для контроля выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	jti: string - Идентификатор сессии.
	//
	// Returns:
	//
	//	*models.Session - Указатель на сессию.
	//	error - Ошибка выполнения операции.
	//	int - HTTP-статус код результата операции:
	//		- 200 OK при успешном получении
	//		- 404 Not Found если сессия не найдена
	//		- 500 Internal Server Error при других ошибках
	ReadSession(ctx context.Context, jti string) (*models.Session, error, int)

	// DeleteSession удаляет сессию из базы данных.
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
	DeleteSession(ctx context.Context, jti string) (error, int)
}

type ExpressionsRepositoryInterface interface {
	// CreateExpression создает новое выражение и связанные с ним задачи.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	expr: *models.Expression - Выражение для создания.
	//
	// Returns:
	//
	//	int64 - ID созданного выражения.
	//	error - Ошибка выполнения операции.
	//	int - HTTP-статус код:
	//	    - 200 OK при успешном создании
	//	    - 500 Internal Server Error при ошибках
	CreateExpression(ctx context.Context, tx *sql.Tx, expr *models.Expression) (int64, error, int)

	// ReadExpressionByID получает выражение по его ID вместе с задачами.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	id: int64 - ID выражения.
	//
	// Returns:
	//
	//	*models.Expression - Найденное выражение.
	//	error - Ошибка выполнения операции.
	//	int - HTTP-статус код:
	//	    - 200 OK при успешном получении
	//	    - 404 Not Found если выражение не найдено
	//	    - 500 Internal Server Error при ошибках
	ReadExpressionByID(ctx context.Context, tx *sql.Tx, id int64) (*models.Expression, error, int)

	// ReadExpressionsByUserID получает все выражения пользователя.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	userID: int64 - ID пользователя.
	//
	// Returns:
	//
	//	[]*models.Expression - Список выражений пользователя.
	//	error - Ошибка выполнения операции.
	//	int - HTTP статус код:
	//	    - 200 OK при успешном получении
	//	    - 404 Not Found если выражения не найдены
	//	    - 500 Internal Server Error при ошибках
	ReadExpressionsByUserID(ctx context.Context, tx *sql.Tx, userID int64) ([]*models.Expression, error, int)

	// ReadExpressionTasks получает все задачи, связанные с указанным выражением.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных (не используется, создается новая).
	//	id: int64 - ID выражения.
	//
	// Returns:
	//
	//	[]*models.Task - Список задач выражения.
	//	error - Ошибка выполнения операции.
	//	int - HTTP статус код:
	//	    - 200 OK при успешном получении
	//		- 404 Not Found если задачи не найдены
	//	    - 500 Internal Server Error при ошибках
	ReadExpressionTasks(ctx context.Context, tx *sql.Tx, id int64) ([]*models.Task, error, int)

	// UpdateExpressionStatus обновляет статус выражения.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	id: int64 - ID выражения.
	//	status: string - Новый статус.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции.
	//	int - HTTP статус код:
	//	    - 200 OK при успешном обновлении
	//	    - 500 Internal Server Error при ошибках
	UpdateExpressionStatus(ctx context.Context, tx *sql.Tx, id int64, status string) (error, int)

	// UpdateExpressionError обновляет сообщение об ошибке выражения.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	id: int64 - ID выражения.
	//	errContent: string - Текст ошибки.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции.
	//	int - HTTP статус код:
	//	    - 200 OK при успешном обновлении
	//	    - 500 Internal Server Error при ошибках
	UpdateExpressionError(ctx context.Context, tx *sql.Tx, id int64, errContent string) (error, int)

	// UpdateExpressionResult обновляет результат вычисления выражения.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	id: int64 - ID выражения.
	//	result: float64 - Результат вычисления.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции
	//	int - HTTP статус код:
	//	    - 200 OK при успешном обновлении
	//	    - 500 Internal Server Error при ошибках
	UpdateExpressionResult(ctx context.Context, tx *sql.Tx, id int64, result float64) (error, int)
}

type TasksRepositoryInterface interface {
	// CreateTask создает новую задачу в базе данных вместе с ее аргументами и зависимостями.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	task: *models.Task - Задача для создания.
	//
	// Returns:
	//
	//	int64 - ID созданной задачи.
	//	error - Ошибка выполнения операции.
	//	int - HTTP статус код:
	//	    - 201 StatusCreated при успешном создании
	//	    - 500 Internal Server Error при ошибках
	CreateTask(ctx context.Context, tx *sql.Tx, task *models.Task) (int64, error, int)

	// ReadTaskByID получает задачу из базы данных по ее ID вместе с аргументами и зависимостями.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	id: int64 - ID задачи.
	//
	// Returns:
	//
	//	*models.Task - Найденная задача.
	//	error - Ошибка выполнения операции.
	//	int - HTTP статус код:
	//	    - 200 OK при успешном получении
	//	    - 404 Not Found если задача не найдена
	//	    - 500 Internal Server Error при ошибках ДБ
	ReadTaskByID(ctx context.Context, tx *sql.Tx, id int64) (*models.Task, error, int)

	// ReadTasksByExpressionID получает все задачи, связанные с указанным выражением.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	expressionID: int64 - ID выражения.
	//
	// Returns:
	//
	//	[]*models.Task - Список найденных задач.
	//	error - Ошибка выполнения операции.
	//	int - HTTP статус код:
	//	    - 200 OK при успешном получении
	//	    - 404 Not Found если задачи не найдены
	//	    - 500 Internal Server Error при ошибках
	ReadTasksByExpressionID(ctx context.Context, tx *sql.Tx, expressionID int64) ([]*models.Task, error, int)

	// ReadUncompletedTasks получает все невыполненные задачи (со статусом 'pending').
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//
	// Returns:
	//
	//	[]*models.Task - Список невыполненных задач.
	//	error - Ошибка выполнения операции.
	//	int - HTTP статус код:
	//	    - 200 OK при успешном получении
	//	    - 404 Not Found если задачи не найдены
	//	    - 500 Internal Server Error при ошибках
	ReadUncompletedTasks(ctx context.Context, tx *sql.Tx) ([]*models.Task, error, int)

	// UpdateTaskDependencies обновляет зависимости задачи.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	task: *models.Task - Задача с обновленными зависимостями.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции.
	//	int - HTTP статус код:
	//	    - 200 OK при успешном обновлении
	//	    - 500 Internal Server Error при ошибках
	UpdateTaskDependencies(ctx context.Context, tx *sql.Tx, task *models.Task) (error, int)

	// UpdateTaskArguments обновляет один из аргументов задачи.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	id: int64 - ID задачи
	//	index: int - Индекс аргумента (0 или 1).
	//	value: *float64 - Новое значение аргумента.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции.
	//	int - HTTP статус код:
	//	    - 200 OK при успешном обновлении
	//	    - 500 Internal Server Error при ошибках
	UpdateTaskArguments(ctx context.Context, tx *sql.Tx, id int64, index int, value *float64) (error, int)

	// UpdateTaskStatus обновляет статус задачи.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	id: int64 - ID задачи.
	//	status: string - Новый статус задачи.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции.
	//	int - HTTP статус код:
	//	    - 200 OK при успешном обновлении
	//	    - 500 Internal Server Error при ошибках
	UpdateTaskStatus(ctx context.Context, tx *sql.Tx, id int64, status string) (error, int)

	// UpdateTaskExpressionID обновляет ID выражения для задачи.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	id: int64 - ID задачи.
	//	exprId: int64 - Новый ID выражения.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции
	//	int - HTTP статус код:
	//	    - 200 OK при успешном обновлении
	//	    - 500 Internal Server Error при ошибках
	UpdateTaskExpressionID(ctx context.Context, tx *sql.Tx, id, exprId int64) (error, int)

	// UpdateTaskResult обновляет результат выполнения задачи.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	result: float64 - Новый результат.
	//	id: int64 - ID задачи.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции
	//	int - HTTP статус код:
	//	    - 200 OK при успешном обновлении
	//	    - 500 Internal Server Error при ошибках
	UpdateTaskResult(ctx context.Context, tx *sql.Tx, result float64, id int64) (error, int)

	// DeleteTasks удаляет все задачи, связанные с указанным выражением.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	id: int64 - ID выражения.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции
	//	int - HTTP статус код:
	//	    - 200 OK при успешном удалении
	//	    - 500 Internal Server Error при ошибках
	DeleteTasks(ctx context.Context, tx *sql.Tx, id int64) (error, int)
}

type TasksDepsRepositoryInterface interface {
	// CreateTaskDeps создает записи зависимостей задачи в базе данных.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	task: *models.Task - Задача, содержащая зависимости для сохранения.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции.
	CreateTaskDeps(ctx context.Context, tx *sql.Tx, task *models.Task) error

	// ReadTaskDeps получает зависимости задачи из базы данных.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	id: int64 - Идентификатор задачи.
	//
	// Returns:
	//
	//	[]int64 - Список идентификаторов зависимых задач.
	//	error - Ошибка выполнения операции.
	ReadTaskDeps(ctx context.Context, tx *sql.Tx, id int64) ([]int64, error)

	// UpdateTaskDeps обновляет зависимости задачи в базе данных.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	id: int64 - Идентификатор задачи.
	//	deps: []int64 - Новые зависимости задачи.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции.
	UpdateTaskDeps(ctx context.Context, tx *sql.Tx, id int64, deps []int64) error
}

type TasksArgsRepositoryInterface interface {
	// CreateTaskArgs создает запись аргументов задачи в базе данных.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	task: *models.Task - Задача, содержащая аргументы для сохранения.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции.
	CreateTaskArgs(ctx context.Context, tx *sql.Tx, task *models.Task) error

	// ReadTaskArgs получает аргументы задачи из базы данных.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	id: int64 - Идентификатор задачи.
	//
	// Returns:
	//
	//	[]*float64 - Срез из двух элементов: [первый аргумент, второй аргумент].
	//	error - Ошибка выполнения операции.
	ReadTaskArgs(ctx context.Context, tx *sql.Tx, id int64) ([]*float64, error)

	// UpdateTaskArgs обновляет один из аргументов задачи в базе данных.
	//
	// Args:
	//
	//	ctx: context.Context - Контекст выполнения запроса.
	//	tx: *sql.Tx - Транзакция базы данных.
	//	id: int64 - Идентификатор задачи.
	//	index: int - Индекс аргумента (0 - первый, 1 - второй).
	//	value: *float64 - Новое значение аргумента.
	//
	// Returns:
	//
	//	error - Ошибка выполнения операции.
	UpdateTaskArgs(ctx context.Context, tx *sql.Tx, id int64, index int, value *float64) error
}
