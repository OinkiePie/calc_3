package jwt_manager

type JWTManagerInterface interface {
	// Generate создает новый JWT токен для указанного пользователя.
	//
	// Args:
	//
	//	userID: int64 - ID пользователя, для которого генерируется токен.
	//
	// Returns:
	//
	//	string - Сгенерированный токен.
	//	string - Уникальный идентификатор токена (jti).
	//	int64 - Время истечения токена в Unix timestamp.
	//	error - Ошибка, если генерация не удалась.
	Generate(userID int64) (string, string, int64, error)

	// Validate проверяет валидность JWT токена и извлекает claims.
	//
	// Args:
	//
	//	tokenString: string - JWT токен для валидации.
	//
	// Returns:
	//
	//	Claims - Извлеченные claims из токена.
	//	error - Ошибка валидации или парсинга токена.
	Validate(tokenString string) (Claims, error)
}
