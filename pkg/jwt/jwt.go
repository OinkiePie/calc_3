package jwt

import (
	"errors"
	"fmt"
	"github.com/OinkiePie/calc_3/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

type JWTManager struct {
	secretKey string
}

type Claims struct {
	Subject   int64  `json:"sub"`
	ExpiresAt int64  `json:"exp"`
	JWTID     string `json:"jti"`
	jwt.RegisteredClaims
}

func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey: secretKey,
	}
}

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
