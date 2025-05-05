package middlewares_test

import (
	"errors"
	mu "github.com/OinkiePie/calc_3/orchestrator/internal/managers"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OinkiePie/calc_3/orchestrator/internal/middlewares"
	mj "github.com/OinkiePie/calc_3/pkg/jwt_manager"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestEnableAuthMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserManager := new(mu.MockUserManager)
	mockJWTManager := new(mj.MockJWTManager)

	mw := middlewares.NewOrchestratorMiddlewares(
		[]string{"*"},
		mockUserManager,
		mockJWTManager,
	)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		authHeader     string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "No Authorization header",
			authHeader:     "",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Отсутствует заголовок Authorization",
		},
		{
			name:           "Invalid Authorization format",
			authHeader:     "InvalidToken",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Неверный формат заголовка Authorization",
		},
		{
			name:           "Empty token",
			authHeader:     "Bearer ",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Пустой ключ авторизации",
		},
		{
			name:       "Invalid token",
			authHeader: "Bearer invalid_token",
			mockSetup: func() {
				mockJWTManager.On("Validate", "invalid_token").Return(mj.Claims{}, errors.New("error"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Сессия была завершена или истекла",
		},
		{
			name:       "Session does not exist",
			authHeader: "Bearer valid_token",
			mockSetup: func() {
				claims := mj.Claims{JWTID: "session123"}
				mockJWTManager.On("Validate", "valid_token").Return(claims, nil)
				mockUserManager.On("SessionExists", mock.Anything, "session123").Return(nil, false)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Сессия была завершена или истекла",
		},
		{
			name:       "Valid token and session",
			authHeader: "Bearer valid_token",
			mockSetup: func() {
				claims := mj.Claims{JWTID: "session123"}
				mockJWTManager.On("Validate", "valid_token").Return(claims, nil)
				mockUserManager.On("SessionExists", mock.Anything, "session123").Return(nil, true)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()

			handler := mw.EnableAuth(nextHandler)
			handler.ServeHTTP(rr, req)

			if tt.expectedBody != "" {
				assert.Contains(t, strings.TrimSpace(rr.Body.String()), tt.expectedBody)
			}
		})
	}
}

func TestEnableCORS(t *testing.T) {
	tests := []struct {
		name            string
		allowedOrigins  []string
		requestOrigin   string
		expectedOrigin  string
		expectedHeaders map[string]string
	}{
		{
			name:           "Allowed origin",
			allowedOrigins: []string{"https://example.com", "https://test.com"},
			requestOrigin:  "https://example.com",
			expectedOrigin: "https://example.com",
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Methods": "GET, POST, OPTIONS, PUT, DELETE",
				"Access-Control-Allow-Headers": "Accept, Content-Type, Content-Length, Authorization",
			},
		},
		{
			name:           "Wildcard origin",
			allowedOrigins: []string{"*"},
			requestOrigin:  "https://any-origin.com",
			expectedOrigin: "*",
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Methods": "GET, POST, OPTIONS, PUT, DELETE",
				"Access-Control-Allow-Headers": "Accept, Content-Type, Content-Length, Authorization",
			},
		},
		{
			name:            "Disallowed origin",
			allowedOrigins:  []string{"https://allowed.com"},
			requestOrigin:   "https://disallowed.com",
			expectedOrigin:  "",
			expectedHeaders: map[string]string{},
		},
		{
			name:           "Case insensitive match",
			allowedOrigins: []string{"https://Example.com"},
			requestOrigin:  "https://example.com",
			expectedOrigin: "https://example.com",
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Methods": "GET, POST, OPTIONS, PUT, DELETE",
				"Access-Control-Allow-Headers": "Accept, Content-Type, Content-Length, Authorization",
			},
		},
		{
			name:            "No origin header",
			allowedOrigins:  []string{"https://example.com"},
			requestOrigin:   "",
			expectedOrigin:  "",
			expectedHeaders: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := middlewares.NewOrchestratorMiddlewares(
				tt.allowedOrigins,
				nil,
				nil,
			)

			req := httptest.NewRequest("GET", "/", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}

			rr := httptest.NewRecorder()

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := mw.EnableCORS(testHandler)
			handler.ServeHTTP(rr, req)

			if tt.expectedOrigin != "" {
				assert.Equal(t, tt.expectedOrigin, rr.Header().Get("Access-Control-Allow-Origin"))
			} else {
				assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
			}

			for header, expectedValue := range tt.expectedHeaders {
				assert.Equal(t, expectedValue, rr.Header().Get(header))
			}
		})
	}
}

func TestEnableCORS_OptionsMethod(t *testing.T) {
	mw := middlewares.NewOrchestratorMiddlewares(
		[]string{"https://example.com"},
		nil,
		nil,
	)

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://example.com")

	rr := httptest.NewRecorder()

	called := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := mw.EnableCORS(testHandler)
	handler.ServeHTTP(rr, req)

	assert.False(t, called)
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, http.StatusOK, rr.Code)
}
