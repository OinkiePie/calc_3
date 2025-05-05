package jwt_manager

type JWTManagerInterface interface {
	Generate(userID int64) (string, string, int64, error)
	Validate(tokenString string) (Claims, error)
}
