package router_test

import (
	"github.com/OinkiePie/calc_3/config"
	mm "github.com/OinkiePie/calc_3/orchestrator/internal/managers"
	"github.com/OinkiePie/calc_3/orchestrator/internal/providers"
	"github.com/OinkiePie/calc_3/orchestrator/internal/router"
	mj "github.com/OinkiePie/calc_3/pkg/jwt_manager"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	// Отключаем выводы и инициализируем конфиг
	log.SetOutput(io.Discard)
	_ = config.InitConfig()
	logger.InitLogger(logger.Options{Level: 4})
}

func TestRoutes_Reachability(t *testing.T) {
	mockPr := &providers.Providers{}
	r := router.NewOrchestratorRouter(mockPr)

	tests := []struct {
		method     string
		path       string
		wantStatus int
	}{
		{http.MethodPost, "/api/register", http.StatusBadRequest},
		{http.MethodPost, "/api/login", http.StatusBadRequest},
		{http.MethodGet, "/api/p/logout", http.StatusUnauthorized},
		{http.MethodPost, "/api/p/delete", http.StatusUnauthorized},
		{http.MethodPost, "/api/p/calculate", http.StatusUnauthorized},
		{http.MethodGet, "/api/p/expressions", http.StatusUnauthorized},
		{http.MethodGet, "/api/p/expressions/1", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, tt.wantStatus, rr.Code, "Путь: %s", tt.path)
	}
}

func TestRouter_CorsMiddleware(t *testing.T) {
	mockPr := &providers.Providers{}
	r := router.NewOrchestratorRouter(mockPr)

	req := httptest.NewRequest(http.MethodOptions, "/api/register", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestRouter_AuthMiddleware(t *testing.T) {
	mockPr := &providers.Providers{}
	r := router.NewOrchestratorRouter(mockPr)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/p/logout"},
		{http.MethodPost, "/api/p/delete"},
		{http.MethodPost, "/api/p/calculate"},
		{http.MethodGet, "/api/p/expressions"},
		{http.MethodGet, "/api/p/expressions/1"},
	}

	for _, tt := range tests {
		method := http.MethodPost
		if tt.method == http.MethodPost {
			method = http.MethodGet
		}
		req := httptest.NewRequest(method, tt.path, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code, "Путь: %s", tt.path)
	}
}

func TestRoute_MethodValidation(t *testing.T) {
	mockJWT := new(mj.MockJWTManager)
	mockUM := new(mm.MockUserManager)
	mockPr := &providers.Providers{JWTManager: mockJWT, UserManager: mockUM}
	r := router.NewOrchestratorRouter(mockPr)

	testClaims := mj.Claims{JWTID: "id"}
	// Обманываем middleware, чтобы добраться до проверки метода
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)
	mockUM.On("SessionExists", mock.Anything, testClaims.JWTID).Return(nil, true)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/register"},
		{http.MethodPost, "/api/login"},
		{http.MethodGet, "/api/p/logout"},
		{http.MethodPost, "/api/p/delete"},
		{http.MethodPost, "/api/p/calculate"},
		{http.MethodGet, "/api/p/expressions"},
		{http.MethodGet, "/api/p/expressions/1"},
	}

	for _, tt := range tests {
		method := http.MethodPost
		if tt.method == http.MethodPost {
			method = http.MethodGet
		}
		req := httptest.NewRequest(method, tt.path, nil)
		req.Header.Set("Authorization", "Bearer valid.token")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code, "Путь: %s", tt.path)
	}
}
