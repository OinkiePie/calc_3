package jwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type JWTManager struct {
	secretKey string
}

type Claims struct {
	Subject   int64 `json:"sub"`
	ExpiresAt int64 `json:"exp"`
	jwt.RegisteredClaims
}

func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey: secretKey,
	}
}

func (m *JWTManager) Generate(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) Validate(tokenString string) (Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return Claims{}, fmt.Errorf("не удалось разобрать токен: %w", err)
	}

	if !token.Valid {
		return Claims{}, errors.New("некорректный токен")
	}

	claimsRaw, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, errors.New("некорректная структура токена")
	}

	claims, err := m.parse(claimsRaw)
	if err != nil {
		return Claims{}, err
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return Claims{}, errors.New("токен просрочен")
	}

	return claims, nil
}

func (m *JWTManager) parse(claims jwt.MapClaims) (Claims, error) {
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

	return Claims{
		Subject:   subInt,
		ExpiresAt: exp,
	}, nil
}
