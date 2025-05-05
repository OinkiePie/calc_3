package jwt_manager_test

import (
	"encoding/base64"
	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/pkg/jwt_manager"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Отключаем выводы и инициализируем конфиг
	log.SetOutput(io.Discard)
	_ = config.InitConfig()
	logger.InitLogger(logger.Options{Level: 6})
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		secretKey string
		name      string
		userID    int64
		wantErr   bool
		checkRes  func(t *testing.T, token, jti string, exp int64, err error)
	}{
		{
			secretKey: "key",
			name:      "successful generation",
			userID:    1,
			wantErr:   false,
			checkRes: func(t *testing.T, token, jti string, exp int64, err error) {
				assert.NotEmpty(t, token)
				assert.NotEmpty(t, jti)
				assert.True(t, exp > time.Now().Unix())
				assert.NoError(t, err)
			},
		},
		// Действующий JWT менеджер допускает создание токена с пустым секретным
		// ключом выведя предупреждение при инициализации менеджера
		{
			secretKey: "",
			name:      "empty secret key",
			userID:    1,
			wantErr:   false,
			checkRes: func(t *testing.T, token, jti string, exp int64, err error) {
				assert.NotEmpty(t, token)
				assert.NotEmpty(t, jti)
				assert.True(t, exp > time.Now().Unix())
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := jwt_manager.NewJWTManager(tt.secretKey)
			token, jti, exp, err := manager.Generate(tt.userID)

			tt.checkRes(t, token, jti, exp, err)
		})
	}
}

func TestValidate(t *testing.T) {
	manager := jwt_manager.NewJWTManager("key")
	userID := int64(1)
	validToken, _, _, err := manager.Generate(userID)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		token    string
		wantErr  bool
		errText  string
		checkRes func(t *testing.T, claims jwt_manager.Claims)
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
			checkRes: func(t *testing.T, claims jwt_manager.Claims) {
				assert.Equal(t, userID, claims.Subject)
			},
		},
		{
			name:    "invalid token",
			token:   "invalid.token.string",
			wantErr: true,
			errText: "не удалось разобрать токен",
		},
		{
			name:    "wrong signing method",
			token:   generateWrongMethodToken(t),
			wantErr: true,
			errText: "неожиданный метод подписи",
		},
		{
			name:    "expired token",
			token:   generateExpiredToken(t, userID),
			wantErr: true,
			errText: "сессия истекла",
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
			errText: "не удалось разобрать токен",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := manager.Validate(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errText != "" {
					assert.Contains(t, err.Error(), tt.errText)
				}
			} else {
				assert.NoError(t, err)
				if tt.checkRes != nil {
					tt.checkRes(t, claims)
				}
			}
		})
	}
}

func generateWrongMethodToken(t *testing.T) string {
	token := jwt.New(jwt.SigningMethodHS256)
	tokenString, err := token.SignedString([]byte("wrong-key"))
	assert.NoError(t, err)

	parts := strings.Split(tokenString, ".")
	parts[0] = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	return strings.Join(parts, ".")
}

func generateExpiredToken(t *testing.T, userID int64) string {
	jti := uuid.New().String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(-time.Hour).Unix(), // Время в прошлом
		"jti": jti,
	})
	tokenString, err := token.SignedString([]byte("key"))
	assert.NoError(t, err)
	return tokenString
}
