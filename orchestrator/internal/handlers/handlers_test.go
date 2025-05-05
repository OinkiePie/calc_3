package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/orchestrator/internal/handlers"
	mm "github.com/OinkiePie/calc_3/orchestrator/internal/managers"
	mj "github.com/OinkiePie/calc_3/pkg/jwt_manager"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	// Отключаем выводы и инициализируем конфиг
	log.SetOutput(io.Discard)
	_ = config.InitConfig()
	logger.InitLogger(logger.Options{Level: 6})
}

// Рекордер использующийся для вызывания
// ошибки при кодировании JSON ответа
type failingRecorder struct {
	*httptest.ResponseRecorder
}

func (fr *failingRecorder) Write(_ []byte) (int, error) {
	return 0, errors.New("error")
}

func TestRegisterUserHandler_CorrectUser_StatusCreated(t *testing.T) {
	mockUM := new(mm.MockUserManager)

	h := handlers.NewOrchestratorHandlers(mockUM, nil, nil)

	mockUM.On("Register", mock.Anything, "user", "pass").
		Return(int64(1), nil, http.StatusCreated)

	requestBody := map[string]string{
		"login":    "user",
		"password": "pass",
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.RegisterUserHandler(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Empty(t, w.Body.String())
	mockUM.AssertExpectations(t)
}

func TestRegisterUserHandler_InvalidMethod_StatusMethodNotAllowed(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	tests := []struct {
		method string
	}{
		{http.MethodGet},
		{http.MethodPut},
		{http.MethodDelete},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/register", nil)
			w := httptest.NewRecorder()

			h.RegisterUserHandler(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
			assert.Equal(t, "метод не поддерживается\n", w.Body.String())
		})
	}
}

func TestRegisterUserHandler_EmptyBody_StatusBadRequest(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/register", nil)
	w := httptest.NewRecorder()

	h.RegisterUserHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "пустое тело запроса\n", w.Body.String())
}

func TestRegisterUserHandler_MissingFields_StatusBadRequest(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	testCases := []struct {
		name string
		body map[string]string
	}{
		{"Missing Login", map[string]string{"password": "pass"}},
		{"Missing Password", map[string]string{"login": "user"}},
		{"Empty Login", map[string]string{"login": "", "password": "pass"}},
		{"Empty Password", map[string]string{"login": "user", "password": ""}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.body)
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			w := httptest.NewRecorder()

			h.RegisterUserHandler(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, "некорректный запрос\n", w.Body.String())
		})
	}
}

func TestRegisterUserHandler_InvalidJSON_StatusUnprocessableEntity(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte("{invalid}")))
	w := httptest.NewRecorder()

	h.RegisterUserHandler(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, "некорректный запрос\n", w.Body.String())
}

func TestRegisterUserHandler_UserExists_StatusConflict(t *testing.T) {
	mockUM := new(mm.MockUserManager)
	h := handlers.NewOrchestratorHandlers(mockUM, nil, nil)

	mockUM.On("Register", mock.Anything, "existing", "pass").
		Return(int64(0), errors.New("логин testexisting уже существует"), http.StatusConflict)

	requestBody := map[string]string{
		"login":    "existing",
		"password": "pass",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.RegisterUserHandler(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Equal(t, "логин testexisting уже существует\n", w.Body.String())
	mockUM.AssertExpectations(t)
}

func TestRegisterUserHandler_InternalError_StatusInternalServerError(t *testing.T) {
	mockUM := new(mm.MockUserManager)
	h := handlers.NewOrchestratorHandlers(mockUM, nil, nil)

	mockUM.On("Register", mock.Anything, "erroneous", "pass").
		Return(int64(0), assert.AnError, http.StatusInternalServerError)

	requestBody := map[string]string{
		"login":    "erroneous",
		"password": "pass",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.RegisterUserHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), assert.AnError.Error())
	mockUM.AssertExpectations(t)
}

func TestLoginUserHandler_CorrectUser_StatusOK(t *testing.T) {
	mockUM := new(mm.MockUserManager)
	h := handlers.NewOrchestratorHandlers(mockUM, nil, nil)

	expectedToken := "test.jwt.token"
	mockUM.On("Login", mock.Anything, "valid", "valid").
		Return(expectedToken, int64(1), nil, http.StatusOK)

	requestBody := map[string]string{
		"login":    "valid",
		"password": "valid",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.LoginUserHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, response["token"])
	mockUM.AssertExpectations(t)
}

func TestLoginUserHandler_InvalidMethod_StatusMethodNotAllowed(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	testCases := []struct {
		method string
	}{
		{http.MethodGet},
		{http.MethodPut},
		{http.MethodDelete},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/login", nil)
			w := httptest.NewRecorder()

			h.LoginUserHandler(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
			assert.Equal(t, "метод не поддерживается\n", w.Body.String())
		})
	}
}

func TestLoginUserHandler_EmptyBody_StatusBadRequest(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	w := httptest.NewRecorder()

	h.LoginUserHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "пустое тело запроса\n", w.Body.String())
}

func TestLoginUserHandler_InvalidJSON_StatusUnprocessableEntity(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte("{invalid}")))
	w := httptest.NewRecorder()

	h.LoginUserHandler(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, "некорректный запрос\n", w.Body.String())
}

func TestLoginUserHandler_InvalidCredentials_StatusUnauthorized(t *testing.T) {
	mockUM := new(mm.MockUserManager)
	h := handlers.NewOrchestratorHandlers(mockUM, nil, nil)

	expectedErr := errors.New("error")
	mockUM.On("Login", mock.Anything, "invalid", "creds").
		Return("", int64(0), expectedErr, http.StatusUnauthorized)

	requestBody := map[string]string{
		"login":    "invalid",
		"password": "creds",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.LoginUserHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), expectedErr.Error())
	mockUM.AssertExpectations(t)
}

func TestLoginUserHandler_InternalError_StatusInternalServerError(t *testing.T) {
	mockUM := new(mm.MockUserManager)
	h := handlers.NewOrchestratorHandlers(mockUM, nil, nil)

	expectedErr := errors.New("error")
	mockUM.On("Login", mock.Anything, "erroneous", "erroneous").
		Return("", int64(0), expectedErr, http.StatusInternalServerError)

	requestBody := map[string]string{
		"login":    "erroneous",
		"password": "erroneous",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.LoginUserHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), expectedErr.Error())
	mockUM.AssertExpectations(t)
}

func TestLoginUserHandler_JSONEncodeError_StatusInternalServerError(t *testing.T) {
	mockUM := new(mm.MockUserManager)
	h := handlers.NewOrchestratorHandlers(mockUM, nil, nil)

	mockUM.On("Login", mock.Anything, "erroneous", "erroneous").
		Return("token", int64(1), nil, http.StatusOK)

	requestBody := map[string]string{
		"login":    "erroneous",
		"password": "erroneous",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := &failingRecorder{ResponseRecorder: httptest.NewRecorder()}

	h.LoginUserHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUM.AssertExpectations(t)
}

func TestLogoutUserHandler_CorrectToken_StatusOK(t *testing.T) {
	mockUM := new(mm.MockUserManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(mockUM, nil, mockJWT)

	testClaims := mj.Claims{
		JWTID:   "session123",
		Subject: 1,
	}
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)
	mockUM.On("Logout", mock.Anything, "session123").Return(nil, http.StatusOK)

	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	req.Header.Set("Authorization", "Bearer valid.token")
	w := httptest.NewRecorder()

	h.LogoutUserHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
	mockUM.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestLogoutUserHandler_InvalidMethod_StatusMethodNotAllowed(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	testCases := []struct {
		method string
	}{
		{http.MethodPost},
		{http.MethodPut},
		{http.MethodDelete},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/logout", nil)
			w := httptest.NewRecorder()

			h.LogoutUserHandler(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
			assert.Equal(t, "метод не поддерживается\n", w.Body.String())
		})
	}
}

func TestLogoutUserHandler_LogoutError_StatusInternalServerError(t *testing.T) {
	mockUM := new(mm.MockUserManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(mockUM, nil, mockJWT)

	testClaims := mj.Claims{
		JWTID:   "session123",
		Subject: 1,
	}
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)
	mockUM.On("Logout", mock.Anything, "session123").Return(errors.New("error"), http.StatusInternalServerError)

	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	req.Header.Set("Authorization", "Bearer valid.token")
	w := httptest.NewRecorder()

	h.LogoutUserHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "error\n", w.Body.String())
	mockUM.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestDeleteUserHandler_CorrectUser_StatusOK(t *testing.T) {
	mockUM := new(mm.MockUserManager)
	h := handlers.NewOrchestratorHandlers(mockUM, nil, nil)

	mockUM.On("Delete", mock.Anything, "user", "pass").
		Return(int64(1), nil, http.StatusOK)

	requestBody := map[string]string{
		"login":    "user",
		"password": "pass",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/delete", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.DeleteUserHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
	mockUM.AssertExpectations(t)
}

func TestDeleteUserHandler_InvalidMethod_StatusMethodNotAllowed(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	testCases := []struct {
		name   string
		method string
	}{
		{"GET", http.MethodGet},
		{"PUT", http.MethodPut},
		{"DELETE", http.MethodDelete},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/delete", nil)
			w := httptest.NewRecorder()

			h.DeleteUserHandler(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
			assert.Equal(t, "метод не поддерживается\n", w.Body.String())
		})
	}
}

func TestDeleteUserHandler_EmptyBody_StatusBadRequest(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/delete", nil)
	w := httptest.NewRecorder()

	h.DeleteUserHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "пустое тело запроса\n", w.Body.String())
}

func TestDeleteUserHandler_InvalidJSON_StatusUnprocessableEntity(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/delete", bytes.NewReader([]byte("{invalid}")))
	w := httptest.NewRecorder()

	h.DeleteUserHandler(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, "некорректный запрос\n", w.Body.String())
}

func TestDeleteUserHandler_MissingFields_StatusBadRequest(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	testCases := []struct {
		name string
		body map[string]string
	}{
		{"Missing Login", map[string]string{"password": "pass"}},
		{"Missing Password", map[string]string{"login": "user"}},
		{"Empty Login", map[string]string{"login": "", "password": "pass"}},
		{"Empty Password", map[string]string{"login": "user", "password": ""}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/delete", bytes.NewReader(body))
			w := httptest.NewRecorder()

			h.DeleteUserHandler(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, "некорректный запрос\n", w.Body.String())
		})
	}
}

func TestDeleteUserHandler_InvalidCredentials_StatusUnauthorized(t *testing.T) {
	mockUM := new(mm.MockUserManager)
	h := handlers.NewOrchestratorHandlers(mockUM, nil, nil)

	mockUM.On("Delete", mock.Anything, "erroneous", "erroneous").
		Return(int64(0), errors.New("error"), http.StatusUnauthorized)

	reqBody := map[string]string{
		"login":    "erroneous",
		"password": "erroneous",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/delete", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.DeleteUserHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "error\n", w.Body.String())
	mockUM.AssertExpectations(t)
}

func TestDeleteUserHandler_InternalError_StatusInternalServerError(t *testing.T) {
	mockUM := new(mm.MockUserManager)
	h := handlers.NewOrchestratorHandlers(mockUM, nil, nil)

	mockUM.On("Delete", mock.Anything, "erroneous", "erroneous").
		Return(int64(0), errors.New("error"), http.StatusInternalServerError)

	reqBody := map[string]string{
		"login":    "erroneous",
		"password": "erroneous",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/delete", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.DeleteUserHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "error\n", w.Body.String())
	mockUM.AssertExpectations(t)
}

func TestAddExpressionHandler_CorrectExpression_StatusOK(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, mockEM, mockJWT)

	testClaims := mj.Claims{Subject: 1}
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)
	mockEM.On("AddExpression", mock.Anything, "2+2", testClaims.Subject).
		Return(int64(1), nil, http.StatusCreated)

	reqBody := map[string]string{"expression": "2+2"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/expressions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid.token")
	w := httptest.NewRecorder()

	h.AddExpressionHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]int64
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), response["id"])
	mockEM.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestAddExpressionHandler_InvalidMethod_StatusMethodNotAllowed(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	testCases := []struct {
		method string
	}{
		{http.MethodGet},
		{http.MethodPut},
		{http.MethodDelete},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/expressions", nil)
			w := httptest.NewRecorder()

			h.AddExpressionHandler(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
			assert.Equal(t, "метод не поддерживается\n", w.Body.String())
		})
	}
}

func TestAddExpressionHandler_EmptyBody_StatusBadRequest(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/expressions", nil)
	w := httptest.NewRecorder()

	h.AddExpressionHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "пустое тело запроса\n", w.Body.String())
}

func TestAddExpressionHandler_InvalidJSON_StatusUnprocessableEntity(t *testing.T) {
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, nil, mockJWT)
	mockJWT.On("Validate", "valid.token").Return(mj.Claims{Subject: 1}, nil)

	req := httptest.NewRequest(http.MethodPost, "/expressions", bytes.NewReader([]byte("{invalid}")))
	req.Header.Set("Authorization", "Bearer valid.token")
	w := httptest.NewRecorder()

	h.AddExpressionHandler(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, "некорректный запрос\n", w.Body.String())
	mockJWT.AssertExpectations(t)
}

func TestAddExpressionHandler_EmptyExpression_StatusBadRequest(t *testing.T) {
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, nil, mockJWT)

	mockJWT.On("Validate", "valid.token").Return(mj.Claims{Subject: 1}, nil)

	testCases := []struct {
		name string
		body string
	}{
		{"Empty", `{"expression": ""}`},
		{"Spaces", `{"expression": "   "}`},
		{"Missing", `{}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/expressions", strings.NewReader(tc.body))
			req.Header.Set("Authorization", "Bearer valid.token")
			w := httptest.NewRecorder()

			h.AddExpressionHandler(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, "выражения обязательно\n", w.Body.String())
		})
	}
}

func TestAddExpressionHandler_InternalError_StatusInternalServerError(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, mockEM, mockJWT)

	testClaims := mj.Claims{Subject: 1}
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)
	mockEM.On("AddExpression", mock.Anything, "error", int64(1)).
		Return(int64(0), errors.New("error"), http.StatusInternalServerError)

	reqBody := map[string]string{"expression": "error"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/expressions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid.token")
	w := httptest.NewRecorder()

	h.AddExpressionHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "error\n", w.Body.String())
	mockJWT.AssertExpectations(t)
	mockEM.AssertExpectations(t)
}

func TestAddExpressionHandler_JSONEncodeError_StatusInternalServerError(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, mockEM, mockJWT)

	testClaims := mj.Claims{Subject: 1}
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)
	mockEM.On("AddExpression", mock.Anything, "2+2", int64(1)).
		Return(int64(1), nil, http.StatusCreated)

	reqBody := map[string]string{"expression": "2+2"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/expressions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid.token")
	w := &failingRecorder{ResponseRecorder: httptest.NewRecorder()}

	h.AddExpressionHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockJWT.AssertExpectations(t)
	mockEM.AssertExpectations(t)
}

func TestGetExpressionsHandler_CorrectToken_StatusOK(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, mockEM, mockJWT)

	testClaims := mj.Claims{Subject: 1}
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)

	expectedExpressions := []*models.Expression{
		{
			ID:               1,
			ExpressionString: "2+2",
		},
	}
	mockEM.On("ReadExpressions", mock.Anything, int64(1)).
		Return(expectedExpressions, nil, http.StatusOK)

	req := httptest.NewRequest(http.MethodGet, "/expressions", nil)
	req.Header.Set("Authorization", "Bearer valid.token")
	w := httptest.NewRecorder()

	h.GetExpressionsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string][]models.ExpressionResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response["expressions"], 1)
	assert.Equal(t, "2+2", response["expressions"][0].ExpressionString)
	mockEM.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestGetExpressionsHandler_InvalidMethod_StatusMethodNotAllowed(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	testCases := []struct {
		method string
	}{
		{http.MethodPost},
		{http.MethodPut},
		{http.MethodDelete},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/expressions", nil)
			w := httptest.NewRecorder()

			h.GetExpressionsHandler(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
			assert.Equal(t, "метод не поддерживается\n", w.Body.String())
		})
	}
}

func TestGetExpressionsHandler_ReadExpressionsError_StatusInternalServerError(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, mockEM, mockJWT)

	testClaims := mj.Claims{Subject: 1}
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)
	mockEM.On("ReadExpressions", mock.Anything, int64(1)).
		Return(([]*models.Expression)(nil), errors.New("error"), http.StatusInternalServerError)

	req := httptest.NewRequest(http.MethodGet, "/expressions", nil)
	req.Header.Set("Authorization", "Bearer valid.token")
	w := httptest.NewRecorder()

	h.GetExpressionsHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "error\n", w.Body.String())
	mockEM.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestGetExpressionsHandler_EmptyExpressions_StatusOK(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, mockEM, mockJWT)

	testClaims := mj.Claims{Subject: 1}
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)
	mockEM.On("ReadExpressions", mock.Anything, int64(1)).
		Return([]*models.Expression{}, nil, http.StatusOK)

	req := httptest.NewRequest(http.MethodGet, "/expressions", nil)
	req.Header.Set("Authorization", "Bearer valid.token")
	w := httptest.NewRecorder()

	h.GetExpressionsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string][]models.ExpressionResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Empty(t, response["expressions"])
	mockEM.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestGetExpressionsHandler_JSONEncodeError_StatusInternalServerError(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, mockEM, mockJWT)

	testClaims := mj.Claims{Subject: 1}
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)

	mockEM.On("ReadExpressions", mock.Anything, int64(1)).
		Return([]*models.Expression{{ID: 1}}, nil, http.StatusOK)

	req := httptest.NewRequest(http.MethodGet, "/expressions", nil)
	req.Header.Set("Authorization", "Bearer valid.token")
	w := &failingRecorder{ResponseRecorder: httptest.NewRecorder()}

	h.GetExpressionsHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockEM.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestGetExpressionHandler_CorrectID_StatusOK(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, mockEM, mockJWT)

	testClaims := mj.Claims{Subject: 1}
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)

	expectedExpression := &models.Expression{
		ID:               1,
		UserID:           1,
		Status:           "completed",
		ExpressionString: "2+2",
	}
	mockEM.On("ReadExpression", mock.Anything, int64(1)).
		Return(expectedExpression, nil, http.StatusOK)

	req := httptest.NewRequest(http.MethodGet, "/expressions/1", nil)
	req.Header.Set("Authorization", "Bearer valid.token")

	vars := map[string]string{"id": "1"}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	h.GetExpressionHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]models.ExpressionResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "2+2", response["expression"].ExpressionString)
	mockEM.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestGetExpressionHandler_InvalidMethod_StatusMethodNotAllowed(t *testing.T) {
	h := handlers.NewOrchestratorHandlers(nil, nil, nil)

	testCases := []struct {
		method string
	}{
		{http.MethodPost},
		{http.MethodPut},
		{http.MethodDelete},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/expressions/1", nil)
			w := httptest.NewRecorder()

			h.GetExpressionHandler(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
			assert.Equal(t, "метод не поддерживается\n", w.Body.String())

		})
	}
}

func TestGetExpressionHandler_InvalidID_StatusBadRequest(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, mockEM, mockJWT)
	mockJWT.On("Validate", "valid.token").Return(mj.Claims{}, nil)

	testCases := []struct {
		name string
		id   string
	}{
		{"NotNumber", "abc"},
		{"Empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/expressions/"+tc.id, nil)
			req.Header.Set("Authorization", "Bearer valid.token")
			req = mux.SetURLVars(req, map[string]string{"id": tc.id})
			w := httptest.NewRecorder()

			h.GetExpressionHandler(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, "не удалось перевести выражение в число\n", w.Body.String())
		})
	}
}

func TestGetExpressionHandler_NotFound_StatusNotFound(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, mockEM, mockJWT)

	testClaims := mj.Claims{Subject: 1}
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)
	mockEM.On("ReadExpression", mock.Anything, int64(999)).
		Return((*models.Expression)(nil), errors.New("error"), http.StatusNotFound)

	req := httptest.NewRequest(http.MethodGet, "/expressions/999", nil)
	req.Header.Set("Authorization", "Bearer valid.token")
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	w := httptest.NewRecorder()

	h.GetExpressionHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "error\n", w.Body.String())
	mockEM.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestGetExpressionHandler_ForbiddenExpression_StatusForbidden(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, mockEM, mockJWT)

	testClaims := mj.Claims{Subject: 1}
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)

	expression := &models.Expression{
		ID:     1,
		UserID: 2,
	}
	mockEM.On("ReadExpression", mock.Anything, int64(1)).
		Return(expression, nil, http.StatusOK)

	req := httptest.NewRequest(http.MethodGet, "/expressions/1", nil)
	req.Header.Set("Authorization", "Bearer valid.token")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.GetExpressionHandler(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "невозможно получить выражение другого пользователя\n", w.Body.String())
	mockEM.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestGetExpressionHandler_InternalError_StatusInternalServerError(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, mockEM, mockJWT)

	testClaims := mj.Claims{Subject: 1}
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)
	mockEM.On("ReadExpression", mock.Anything, int64(1)).
		Return((*models.Expression)(nil), errors.New("error"), http.StatusInternalServerError)

	req := httptest.NewRequest(http.MethodGet, "/expressions/1", nil)
	req.Header.Set("Authorization", "Bearer valid.token")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.GetExpressionHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "error\n", w.Body.String())
	mockEM.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}

func TestGetExpressionHandler_JSONEncodeError_StatusInternalServerError(t *testing.T) {
	mockEM := new(mm.MockExpressionManager)
	mockJWT := new(mj.MockJWTManager)
	h := handlers.NewOrchestratorHandlers(nil, mockEM, mockJWT)

	testClaims := mj.Claims{Subject: 1}
	mockJWT.On("Validate", "valid.token").Return(testClaims, nil)

	mockEM.On("ReadExpression", mock.Anything, int64(1)).
		Return(&models.Expression{UserID: 1}, nil, http.StatusOK)

	req := httptest.NewRequest(http.MethodGet, "/expressions/1", nil)
	req.Header.Set("Authorization", "Bearer valid.token")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := &failingRecorder{ResponseRecorder: httptest.NewRecorder()}

	h.GetExpressionHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockEM.AssertExpectations(t)
	mockJWT.AssertExpectations(t)
}
