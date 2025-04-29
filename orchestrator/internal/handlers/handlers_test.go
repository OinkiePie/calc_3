package handlers_test

//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/OinkiePie/calc_3/orchestrator/internal/managers"
//	"github.com/OinkiePie/calc_3/pkg/jwt"
//	"github.com/OinkiePie/calc_3/pkg/logger"
//	"github.com/OinkiePie/calc_3/pkg/models"
//	"github.com/gorilla/mux"
//	"io"
//	"net/http"
//	"strconv"
//	"strings"
//)
//
//// Handlers - структура для обработчиков запросов, зависит от TaskManager
//type Handlers struct {
//	userManager managers.UserManagerInterface
//	exprManager managers.ExpressionManagerInterface
//	jwtManager  *jwt.JWTManager
//}
//
//func NewOrchestratorHandlers(um managers.UserManagerInterface, em managers.ExpressionManagerInterface, jwtm *jwt.JWTManager) *Handlers {
//	return &Handlers{userManager: um, exprManager: em, jwtManager: jwtm}
//}
//
//func (h *Handlers) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
//	var req struct {
//		Login    string `json:"login"`
//		Password string `json:"password"`
//	}
//	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//		http.Error(w, "Invalid request", http.StatusBadRequest)
//		return
//	}
//
//	id, err, code := h.userManager.Register(r.Context(), req.Login, req.Password)
//	if err != nil {
//		http.Error(w, err.Error(), code)
//		return
//	}
//
//	w.WriteHeader(http.StatusCreated)
//
//	logger.Log.Debugf("Пользователь №%d (%s) создан", id, req.Login)
//
//}
//
//func (h *Handlers) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
//	var req struct {
//		Login    string `json:"login"`
//		Password string `json:"password"`
//	}
//	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//		http.Error(w, "Invalid request", http.StatusBadRequest)
//		return
//	}
//
//	token, id, err, code := h.userManager.Login(r.Context(), req.Login, req.Password)
//	if err != nil {
//		http.Error(w, fmt.Sprintf("ошибка при входе: %s", err.Error()), code)
//		return
//	}
//
//	err = json.NewEncoder(w).Encode(map[string]string{"token": token})
//	if err != nil {
//		http.Error(w, "ошибка при кодировании ответа в JSON", http.StatusInternalServerError)
//		return
//	}
//
//	logger.Log.Debugf("Пользователь №%d (%s) вошел", id, req.Login)
//
//}
//
//func (h *Handlers) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
//	var req struct {
//		Login    string `json:"login"`
//		Password string `json:"password"`
//	}
//	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//		http.Error(w, "Invalid request", http.StatusBadRequest)
//		return
//	}
//
//	id, err, code := h.userManager.Delete(r.Context(), req.Login, req.Password)
//	if err != nil {
//		http.Error(w, "не удалось удалить", code)
//		return
//	}
//
//	w.WriteHeader(http.StatusOK)
//
//	logger.Log.Debugf("Пользователь №%d (%s) удален", id, req.Login)
//}
//
//// AddExpressionHandler обрабатывает POST-запросы на эндпоинт /api/v1/calculate.
////
//// Функция принимает JSON-запрос, содержащий математическое выражение в строковом формате,
//// передает выражение в TaskManager для обработки и сохранения, и возвращает ID созданного выражения.
////
//// Args:
////
////	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
////	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
////
//// Request body (JSON):
////
////	{
////		"expression": "строка с математическим выражением"
////	}
////
//// Responses:
////
////	201 Created:
////	{
////		"id": "уникальный ID созданного выражения"
////	}
////
////	400 Bad Request:
////	{
////		"error": "выражения обязательно"
////	}
////
////	{
////		"error": "пустое тело запроса"
////	}
////
////	405 Method Not Allowed:
////	{
////		"error": "метод не поддерживается"
////	}
////
////	422 Unprocessable Entity:
////	{
////		"error": "не удалось декодировать JSON"
////	}
////
////	{
////		"error": "Содержание ошибки при добавлении выражения в TaskManager"
////	}
////
////	500 Internal Server Error:
////	{
////		"error": "не удалось прочитать запрос"
////	}
////	{
////		"error": "ошибка при кодировании ответа в JSON."
////	}
//func (h *Handlers) AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
//	if r.Method != http.MethodPost {
//		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
//		return
//	}
//
//	if r.Body == nil {
//		http.Error(w, "пустое тело запроса", http.StatusBadRequest)
//		return
//	}
//
//	body, err := io.ReadAll(r.Body)
//	if err != nil {
//		http.Error(w, "не удалось прочитать запрос", http.StatusInternalServerError)
//		return
//	}
//
//	authHeader := r.Header.Get("Authorization")
//	token := strings.TrimPrefix(authHeader, "Bearer ")
//	claims, err := h.jwtManager.Validate(token)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusUnauthorized)
//		return
//	}
//
//	var requestBody models.ExpressionAdd
//
//	err = json.Unmarshal(body, &requestBody)
//	if err != nil {
//		http.Error(w, "не удалось декодировать JSON", http.StatusUnprocessableEntity)
//		return
//	}
//
//	trimmedBody := strings.TrimSpace(requestBody.Expression)
//	if trimmedBody == "" {
//		http.Error(w, "выражения обязательно", http.StatusBadRequest)
//		return
//	}
//
//	id, err, code := h.exprManager.AddExpression(r.Context(), trimmedBody, claims.Subject)
//	if err != nil {
//		http.Error(w, err.Error(), code)
//		return
//	}
//
//	response := map[string]int64{"id": id}
//	w.Header().Set("Content-Type", "application/json")
//	err = json.NewEncoder(w).Encode(response)
//	if err != nil {
//		http.Error(w, "ошибка при кодировании ответа в JSON", http.StatusInternalServerError)
//		return
//	}
//
//	logger.Log.Debugf("Выражение №%d пользователя №%d создано", id, claims.Subject)
//}
//
//// GetExpressionsHandler обрабатывает GET-запросы на эндпоинт /api/v1/expressions.
////
//// Функция получает список всех выражений из TaskManager, преобразует их в формат ExpressionResponse
//// и возвращает JSON-ответ со списком выражений.
////
//// Args:
////
////	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
////	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
////
//// Responses:
////
////	200 OK:
////	{
////	  "expressions": [
////	    {
////				"id": "уникальный ID выражения",
////				"status": "статус выражения (pending, processing, completed, error)",
////				"result": "результат выражения (может отсутствовать, если вычисления не завершены)",
////				"error": "ошибка при вычислении (может отсутствовать, если ошибки нет)"
////	    },
////	    ...
////	  ]
////	}
////
////	405 Method Not Allowed:
////	{
////		"error": "метод не поддерживается"
////	}
////
////	500 Internal Server Error:
////	{
////		"error": "ошибка при кодировании ответа в JSON."
////	}
//func (h *Handlers) GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
//	if r.Method != http.MethodGet {
//		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
//		return
//	}
//
//	authHeader := r.Header.Get("Authorization")
//	token := strings.TrimPrefix(authHeader, "Bearer ")
//	claims, err := h.jwtManager.Validate(token)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusUnauthorized)
//		return
//	}
//
//	expressions, err, code := h.exprManager.ReadExpressions(r.Context(), claims.Subject)
//	if err != nil {
//		http.Error(w, err.Error(), code)
//		return
//	}
//
//	var expressionResponses []models.ExpressionResponse
//
//	for _, expression := range expressions {
//		expressionResponse := models.ExpressionResponse{
//			ID:               expression.ID,
//			Status:           expression.Status,
//			ExpressionString: expression.ExpressionString,
//			Result:           expression.Result,
//			Error:            expression.Error,
//		}
//		expressionResponses = append(expressionResponses, expressionResponse)
//	}
//
//	response := map[string][]models.ExpressionResponse{"expressions": expressionResponses}
//
//	w.Header().Set("Content-Type", "application/json")
//	err = json.NewEncoder(w).Encode(response)
//	if err != nil {
//		http.Error(w, "ошибка при кодировании ответа в JSON", http.StatusInternalServerError)
//		return
//	}
//
//	logger.Log.Debugf("Список выражений пользователя №%d отправлен", claims.Subject)
//}
//
//// GetExpressionHandler обрабатывает GET-запросы на эндпоинт /api/v1/expressions/{id}.
////
//// Функция получает выражение по указанному ID из TaskManager, преобразует его в формат ExpressionResponse
//// и возвращает JSON-ответ с информацией о выражении.
////
//// Args:
////
////	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
////	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
////
//// Path parameters:
////
////	id: ID выражения, которое нужно получить.
////
//// Responses:
////
////	200 OK:
////	{
////		"expression": {
////			"id": "уникальный ID выражения",
////			"status": "статус выражения (pending, processing, completed, error)",
////			"result": "результат выражения (может отсутствовать, если вычисления не завершены)",
////			"error": "ошибка при вычислении (может отсутствовать, если ошибки нет)"
////		}
////	}
////
////	404 Not Found:
////	{
////	  "error": "выражение не найдено"
////	}
////
////	405 Method Not Allowed:
////	{
////		"error": "метод не поддерживается"
////	}
////
////	500 Internal Server Error:
////	{
////	  "error": "ошибка при кодировании ответа в JSON"
////	}
//func (h *Handlers) GetExpressionHandler(w http.ResponseWriter, r *http.Request) {
//	if r.Method != http.MethodGet {
//		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
//		return
//	}
//
//	authHeader := r.Header.Get("Authorization")
//	token := strings.TrimPrefix(authHeader, "Bearer ")
//	if _, err := h.jwtManager.Validate(token); err != nil {
//		http.Error(w, err.Error(), http.StatusUnauthorized)
//		return
//	}
//
//	vars := mux.Vars(r)
//	idStr := vars["id"]
//	id, err := strconv.ParseInt(idStr, 10, 64)
//	if err != nil {
//		http.Error(w, "не удалось выражение в число", http.StatusBadRequest)
//	}
//
//	expression, err, code := h.exprManager.ReadExpression(r.Context(), id)
//	if err != nil {
//		http.Error(w, err.Error(), code)
//		return
//	}
//
//	expressionResponse := models.ExpressionResponse{
//		ID:               expression.ID,
//		Status:           expression.Status,
//		ExpressionString: expression.ExpressionString,
//		Result:           expression.Result,
//		Error:            expression.Error,
//	}
//
//	response := map[string]models.ExpressionResponse{"expression": expressionResponse}
//
//	w.Header().Set("Content-Type", "application/json")
//
//	err = json.NewEncoder(w).Encode(response)
//	if err != nil {
//		http.Error(w, "ошибка при кодировании ответа в JSON", http.StatusInternalServerError)
//		return
//	}
//
//	logger.Log.Debugf("Выражение №%d пользователя №%d отправлено", expression.ID, id)
//}
//
//// GetTaskHandler обрабатывает GET-запросы на эндпоинт /internal/task.
////
//// Функция получает задачу для выполнения из TaskManager и возвращает JSON-ответ с информацией о задаче.
//// Этот эндпоинт предназначен для внутреннего использования агентом.
////
//// Args:
////
////	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
////	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
////
//// Responses:
////
////	200 OK:
////	{
////		"id": "уникальный ID задачи",
////		"operation": "операция, которую нужно выполнить (+, -, *, /, ^, u-)",
////		"args": [], // 2 числа
////		"operation_time": "время выполнения задачи",
////		"expression": "ID выражения, составной частью которого является задача"
////	}
////
////	404 Not Found:
////		(пустой ответ) - Если нет доступных задач для выполнения
////
////	405 Method Not Allowed:
////	{
////		"error": "метод не поддерживается"
////	}
//func (h *Handlers) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
//	if r.Method != http.MethodGet {
//		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
//		return
//	}
//
//	task, err, code := h.exprManager.ReadTask(r.Context())
//	if err != nil {
//		http.Error(w, err.Error(), code)
//		return
//	}
//	if task == nil {
//		http.Error(w, "нет доступных задач", http.StatusNotFound)
//		return
//	}
//
//	response := models.TaskResponse{
//		ID:             task.ID,
//		Args:           task.Args,
//		Operation:      task.Operation,
//		Operation_time: task.OperationTime,
//		Expression:     task.Expression,
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//
//	err = json.NewEncoder(w).Encode(response)
//	if err != nil {
//		http.Error(w, "ошибка при кодировании ответа в JSON", http.StatusInternalServerError)
//		return
//	}
//
//	logger.Log.Debugf("Задача №%d отправлена", task.ID)
//}
//
//// CompleteTaskHandler обрабатывает POST-запросы на эндпоинт /internal/task.
////
//// Функция принимает JSON-запрос с ID выполненной задачи и результатом ее выполнения,
//// обновляет информацию о задаче в TaskManager. Этот эндпоинт предназначен для внутреннего использования агентами.
////
//// Args:
////
////	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
////	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
////
//// Request body (JSON):
////
////	{
////		"expression": "ID выражения, частью которого являетя задача"
////		"id": "ID выполненной задачи",
////		"result": "результат выполнения задачи (число)",
////		"error": "ошибка, возикшая при выполнении задачи" (может отсутсвовать)
////	}
////
//// Responses:
////
////	200 OK:
////	(пустой ответ) - В случае успешного завершения.
////
////
////	400 Bad Request:
////	{
////		"error": "пустое тело запроса"
////	}
////
////	404 Not Found:
////	{
////		"error": "задача не найдена"
////	}
////
////	405 Method Not Allowed:
////	{
////		"error": "метод не поддерживается"
////	}
////
////	422 Unprocessable Entity:
////	{
////		"error": "не удалось декодировать JSON"
////	}
////
////	500 Internal Server Error:
////	{
////		"error": "не удалось прочитать тело запроса"
////	}
//func (h *Handlers) CompleteTaskHandler(w http.ResponseWriter, r *http.Request) {
//	if r.Method != http.MethodPost {
//		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
//		return
//	}
//
//	if r.Body == nil {
//		http.Error(w, "пустое тело запроса", http.StatusBadRequest)
//		return
//	}
//
//	body, err := io.ReadAll(r.Body)
//	if err != nil {
//		http.Error(w, "не удалось прочитать запрос", http.StatusInternalServerError)
//		return
//	}
//
//	var requestBody models.TaskCompleted
//
//	err = json.Unmarshal(body, &requestBody)
//	if err != nil {
//		http.Error(w, "не удалось декодировать JSON", http.StatusUnprocessableEntity)
//		return
//	}
//
//	if err, code := h.exprManager.CompleteTask(r.Context(), &requestBody); err != nil {
//		http.Error(w, err.Error(), code)
//		return
//	}
//
//	w.WriteHeader(http.StatusOK)
//
//	logger.Log.Debugf("Задача %d успешно завершена", requestBody.ID)
//}
