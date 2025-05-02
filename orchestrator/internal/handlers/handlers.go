package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/OinkiePie/calc_3/orchestrator/internal/managers"
	"github.com/OinkiePie/calc_3/pkg/jwt"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Handlers struct {
	userManager managers.UserManagerInterface
	exprManager managers.ExpressionManagerInterface
	jwtManager  *jwt.JWTManager
}

func NewOrchestratorHandlers(um managers.UserManagerInterface, em managers.ExpressionManagerInterface, jwtm *jwt.JWTManager) *Handlers {
	return &Handlers{userManager: um, exprManager: em, jwtManager: jwtm}
}

func (h *Handlers) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "некорректный запрос", http.StatusBadRequest)
		return
	}

	id, err, code := h.userManager.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusCreated)

	logger.Log.Debugf("Пользователь №%d (%s) создан", id, req.Login)
}

func (h *Handlers) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "некорректный запрос", http.StatusBadRequest)
		return
	}

	token, id, err, code := h.userManager.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка при входе: %s", err.Error()), code)
		return
	}

	err = json.NewEncoder(w).Encode(map[string]string{"token": token})
	if err != nil {
		http.Error(w, "ошибка при кодировании ответа в JSON", http.StatusInternalServerError)
		return
	}

	logger.Log.Debugf("Пользователь №%d (%s) вошел", id, req.Login)
}

func (h *Handlers) LogoutUserHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, _ := h.jwtManager.Validate(token)

	if err, code := h.userManager.Logout(r.Context(), claims.JWTID); err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	logger.Log.Debugf("Сессия %s пользователя №%d остановлена", claims.JWTID, claims.Subject)
}

func (h *Handlers) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "некорректный запрос", http.StatusBadRequest)
		return
	}

	id, err, code := h.userManager.Delete(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, "не удалось удалить", code)
		return
	}

	w.WriteHeader(http.StatusOK)

	logger.Log.Debugf("Пользователь №%d (%s) удален", id, req.Login)
}

func (h *Handlers) AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	if r.Body == nil {
		http.Error(w, "пустое тело запроса", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "не удалось прочитать запрос", http.StatusInternalServerError)
		return
	}

	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, _ := h.jwtManager.Validate(token)

	var requestBody models.ExpressionAdd

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		http.Error(w, "не удалось декодировать JSON", http.StatusUnprocessableEntity)
		return
	}

	trimmedBody := strings.TrimSpace(requestBody.Expression)
	if trimmedBody == "" {
		http.Error(w, "выражения обязательно", http.StatusBadRequest)
		return
	}

	id, err, code := h.exprManager.AddExpression(r.Context(), trimmedBody, claims.Subject)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	response := map[string]int64{"id": id}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "ошибка при кодировании ответа в JSON", http.StatusInternalServerError)
		return
	}

	logger.Log.Debugf("Выражение №%d пользователя №%d создано", id, claims.Subject)
}

func (h *Handlers) GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	logger.Log.Fatalf("ss")

	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, _ := h.jwtManager.Validate(token)

	expressions, err, code := h.exprManager.ReadExpressions(r.Context(), claims.Subject)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	var expressionResponses []models.ExpressionResponse

	for _, expression := range expressions {
		expressionResponse := models.ExpressionResponse{
			ID:               expression.ID,
			Status:           expression.Status,
			ExpressionString: expression.ExpressionString,
			Result:           expression.Result,
			Error:            expression.Error,
		}
		expressionResponses = append(expressionResponses, expressionResponse)
	}

	response := map[string][]models.ExpressionResponse{"expressions": expressionResponses}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "ошибка при кодировании ответа в JSON", http.StatusInternalServerError)
		return
	}

	logger.Log.Debugf("Список выражений пользователя №%d отправлен", claims.Subject)
}

func (h *Handlers) GetExpressionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, _ := h.jwtManager.Validate(token)

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "не удалось выражение в число", http.StatusBadRequest)
	}

	expression, err, code := h.exprManager.ReadExpression(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	if claims.Subject != expression.UserID {
		http.Error(w, "невозможно получить выражение другого пользователя", http.StatusForbidden)
		return
	}

	expressionResponse := models.ExpressionResponse{
		ID:               expression.ID,
		Status:           expression.Status,
		ExpressionString: expression.ExpressionString,
		Result:           expression.Result,
		Error:            expression.Error,
	}

	response := map[string]models.ExpressionResponse{"expression": expressionResponse}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "ошибка при кодировании ответа в JSON", http.StatusInternalServerError)
		return
	}

	logger.Log.Debugf("Выражение №%d пользователя №%d отправлено", expression.ID, id)
}
