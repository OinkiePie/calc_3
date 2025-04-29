package models

// Expression представляет структуру арифметического выражения.
type Expression struct {
	UserID int64
	// ID - Уникальный идентификатор выражения.
	ID int64
	// Status - Статус выражения ("pending", "processing", "completed", "error").
	Status string
	// Result - Указатель на результат вычисления выражения. Может быть nil, если вычисление ещё не завершено или ошибочно.
	Result *float64
	// Tasks - Список задач, составляющих выражение.
	Tasks []*Task
	// ExpressionString - Исходное выражение в виде строки.
	ExpressionString string
	// Error - Описание ошибки если выражение невозможно выполнить.
	Error string
}

// ExpressionResponse представляет структуру для отправки информации о выражении в HTTP-ответе.
type ExpressionResponse struct {
	// ID - Уникальный идентификатор выражения.
	ID int64 `json:"id"`
	// Status - Статус выражения.
	Status string `json:"status"`
	// Status - Статус выражения.
	ExpressionString string `json:"expression"`
	// Result - Указатель на результат вычисления выражения. Если nil, то поле не включается в JSON-ответ (omitempty).
	Result *float64 `json:"result,omitempty"` //omitempty - если result nil, то не выводить его
	// Error - Описание ошибки если выражение невозможно выполнить. Если nil, то поле не включается в JSON-ответ (omitempty).
	Error string `json:"error,omitempty"` //omitempty - если result nil, то не выводить его
}

// ExpressionAdd представляет структуру для получения математического выражения из HTTP-запроса.
// Используется для декодирования вырожения из тела запроса.
type ExpressionAdd struct {
	// Expression - Математическое выражение в виде строки.
	Expression string `json:"expression"`
}
