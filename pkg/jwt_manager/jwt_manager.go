package jwt_manager

import (
	"errors"
	"fmt"
	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

// JWTManager предоставляет функциональность для работы с JWT токенами.
type JWTManager struct {
	secretKey string // Секретный ключ для подписи токенов
}

// Claims представляет кастомные claims JWT токена.
type Claims struct {
	Subject   int64  `json:"sub"` // ID пользователя
	ExpiresAt int64  `json:"exp"` // Время истечения токена в Unix timestamp
	JWTID     string `json:"jti"` // Уникальный идентификатор токена
	jwt.RegisteredClaims
}

// NewJWTManager создает новый экземпляр JWTManager.
//
// Args:
//
//	secretKey: string - Секретный ключ для подписи токенов.
//
// Returns:
//
//	*JWTManager - Указатель на созданный JWTManager.
func NewJWTManager(secretKey string) *JWTManager {
	if secretKey == "" {
		logger.Log.Warnf("Секретный ключ пуст")
	}
	return &JWTManager{
		secretKey: secretKey,
	}
}

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
func (m *JWTManager) Generate(userID int64) (string, string, int64, error) {

	jti := uuid.New().String()
	exp := time.Now().Add(time.Minute * time.Duration(config.Cfg.Middleware.TOKEN_TTL_MIN)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": exp,
		"jti": jti,
	})
	signedToken, err := token.SignedString([]byte(m.secretKey))
	return signedToken, jti, exp, err
}

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
func (m *JWTManager) Validate(tokenString string) (Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return Claims{}, fmt.Errorf("не удалось разобрать токен: %w", err)
	}

	if !token.Valid && !errors.Is(err, jwt.ErrTokenExpired) {
		return Claims{}, errors.New("некорректный токен")
	}

	claimsRaw, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, errors.New("некорректная структура токена")
	}

	claims, err := m.parseClaims(claimsRaw)
	if err != nil {
		return Claims{}, err
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return claims, errors.New("сессия истекла")
	}

	return claims, nil
}

// parseClaims преобразует raw claims из JWT токена в структуру Claims.
//
// Args:
//
//	claims: jwt.MapClaims - Сырые claims из JWT токена.
//
// Returns:
//
//	Claims - Преобразованная структура claims.
//	error - Ошибка, если преобразование не удалось.
func (m *JWTManager) parseClaims(claims jwt.MapClaims) (Claims, error) {
	subInterface, ok := claims["sub"]
	if !ok {
		return Claims{}, errors.New("не удалось извлечь пользователя")
	}

	subFloat, ok := subInterface.(float64)
	if !ok {
		return Claims{}, errors.New("не удалось преобразовать пользователя в число")
	}
	subInt := int64(subFloat)

	expInterface, ok := claims["exp"]
	if !ok {
		return Claims{}, errors.New("не удалось извлечь время истечения")
	}

	expFloat, ok := expInterface.(float64)
	if !ok {
		return Claims{}, errors.New("не удалось преобразовать время истечения в число")
	}
	exp := int64(expFloat)

	jtiInterface, ok := claims["jti"]
	jti, ok := jtiInterface.(string)
	if !ok {
		return Claims{}, errors.New("не удалось преобразовать идентификатор сессии в число")
	}

	return Claims{
		Subject:   subInt,
		ExpiresAt: exp,
		JWTID:     jti,
	}, nil
}
