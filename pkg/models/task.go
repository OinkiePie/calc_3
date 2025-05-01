package models

// Task представляет структуру для части арифметического выражения, которую нужно вычислить.
type Task struct {
	// ID - Уникальный идентификатор задачи.
	ID int64
	// Args - Срез указателей на аргументы задачи.
	Args []*float64
	// Operation - Операция, которую необходимо выполнить.
	Operation string
	// Dependencies - Срез ID задач-зависимостей.
	Dependencies []int64
	// Status - Статус выполнения выражения
	Status string
	// Result - Результат вычисления задачи.
	Result *float64
	// Expression - ID выражения, к которому принадлежит данная задача.
	Expression int64

	DependencyIndexes []int
}

// TaskResponse представляет структуру для отправки информации о задаче в HTTP-ответе.
type TaskResponse struct {
	// ID - Уникальный идентификатор задачи.
	ID int64 `json:"id"`
	// Args - Срез указателей на аргументы задачи.
	Args []*float64 `json:"args"`
	// Operation - Операция, которую необходимо выполнить.
	Operation string `json:"operation"`
	// Expression - ID выражения, к которому принадлежит данная задача.
	// Используется для опитмизации возвращения резульата агентом.
	Expression int64 `json:"expression"`
	// Error - Указывает на невыполниасть задачи
	Error string `json:"error,omitempty"`
}

// TaskCompleted представляет структуру для получения информации о завершенной задаче из HTTP-запроса.
// Используется для декодирования вырожения из тела запроса.
type TaskCompleted struct {
	// Expression - ID корневого выражения, к которому принадлежит задача.
	Expression int64 `json:"expression"`
	// ID - Уникальный идентификатор задачи.
	ID int64 `json:"id"`
	// Result - Результат вычисления задачи.
	Result float64 `json:"result"`
	// Error - Указывает на невыполнимость задачи
	Error string `json:"error,omitempty"`
}
