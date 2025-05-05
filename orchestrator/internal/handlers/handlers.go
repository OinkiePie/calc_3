package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/OinkiePie/calc_3/orchestrator/internal/managers"
	"github.com/OinkiePie/calc_3/pkg/jwt_manager"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

// Handlers представляет структуру обработчиков HTTP-запросов оркестратора.
// Содержит менеджеры для работы с пользователями, выражениями и JWT-токенами.
type Handlers struct {
	userManager managers.UserManagerInterface
	exprManager managers.ExpressionManagerInterface
	jwtManager  jwt_manager.JWTManagerInterface
}

// NewOrchestratorHandlers создает новый экземпляр Handlers с заданными менеджерами.
//
// Args:
//
//	um: managers.UserManagerInterface - Менеджер для работы с пользователями.
//	em: managers.ExpressionManagerInterface - Менеджер для работы с выражениями.
//	jwtm: jwt_manager.JWTManagerInterface - Менеджер для работы с JWT-токенами.
//
// Returns:
//
//	*Handlers - Указатель на новый экземпляр Handlers.
func NewOrchestratorHandlers(um managers.UserManagerInterface, em managers.ExpressionManagerInterface, jwtm jwt_manager.JWTManagerInterface) *Handlers {
	return &Handlers{userManager: um, exprManager: em, jwtManager: jwtm}
}

// RegisterUserHandler обрабатывает HTTP-запрос на регистрацию нового пользователя.
//
// Args:
//
//	w: http.ResponseWriter - Интерфейс для записи HTTP-ответа.
//	r: *http.Request - Входящий HTTP-запрос.
//
// Требования:
//   - Метод: POST
//
// Ожидаемые поля в теле запроса (JSON):
//   - login: string - Логин пользователя.
//   - password: string - Пароль пользователя.
//
// Возможные HTTP-статусы ответа:
//   - 201 Created - при успешной регистрации
//   - 400 Bad Request - при некорректном теле запроса
//   - 405 Method Not Allowed - при неправильном методе запроса
//   - 409 Conflict - если пользователь с таким логином уже существует
//   - 422 Unprocessable Entity - при ошибке парсинга JSON
//   - 500 Internal Server Error - при внутренних ошибках сервера
func (h *Handlers) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	if r.ContentLength == 0 {
		http.Error(w, "пустое тело запроса", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "некорректный запрос", http.StatusUnprocessableEntity)
		return
	}

	if requestBody.Login == "" || requestBody.Password == "" {
		http.Error(w, "некорректный запрос", http.StatusBadRequest)
		return
	}

	id, err, code := h.userManager.Register(r.Context(), requestBody.Login, requestBody.Password)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusCreated)

	logger.Log.Debugf("Пользователь №%d (%s) создан", id, requestBody.Login)
}

// LoginUserHandler обрабатывает HTTP-запрос на аутентификацию пользователя.
//
// Args:
//
//	w: http.ResponseWriter - Интерфейс для записи HTTP-ответа
//	r: *http.Request - Входящий HTTP-запрос
//
// Требования:
//   - Метод: POST
//
// Ожидаемые поля в теле запроса (JSON):
//   - login: string - Логин пользователя
//   - password: string - Пароль пользователя
//
// Ответ (JSON):
//   - token: string - JWT-токен для аутентификации
//
// Возможные HTTP-статусы ответа:
//   - 200 OK - при успешной аутентификации
//   - 400 Bad Request - при некорректном теле запроса
//   - 401 Unauthorized - при неверных учетных данных
//   - 405 Method Not Allowed - при неправильном методе запроса
//   - 422 Unprocessable Entity - при ошибке парсинга JSON
//   - 500 Internal Server Error - при внутренних ошибках сервера
func (h *Handlers) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	if r.ContentLength == 0 {
		http.Error(w, "пустое тело запроса", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "некорректный запрос", http.StatusUnprocessableEntity)
		return
	}

	token, id, err, code := h.userManager.Login(r.Context(), requestBody.Login, requestBody.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка при входе: %s", err.Error()), code)
		return
	}

	if err = json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		http.Error(w, "ошибка при кодировании ответа в JSON", http.StatusInternalServerError)
		return
	}

	logger.Log.Debugf("Пользователь №%d (%s) вошел", id, requestBody.Login)
}

// LogoutUserHandler обрабатывает HTTP-запрос на завершение сессии пользователя.
//
// Args:
//
//	w: http.ResponseWriter - Интерфейс для записи HTTP-ответа
//	r: *http.Request - Входящий HTTP-запрос с заголовком Authorization
//
// Требования:
//   - Метод: GET
//   - Заголовок Authorization: Bearer <token> - JWT-токен текущей сессии
//
// Возможные HTTP-статусы ответа:
//   - 200 OK - при успешном завершении сессии
//   - 405 Method Not Allowed - при неправильном методе запроса
//   - 500 Internal Server Error - при внутренних ошибках сервера
func (h *Handlers) LogoutUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, _ := h.jwtManager.Validate(token)

	if err, code := h.userManager.Logout(r.Context(), claims.JWTID); err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	logger.Log.Debugf("Сессия %s пользователя №%d остановлена", claims.JWTID, claims.Subject)
}

// DeleteUserHandler обрабатывает HTTP-запрос на удаление пользователя.
//
// Args:
//
//	w: http.ResponseWriter - Интерфейс для записи HTTP-ответа
//	r: *http.Request - Входящий HTTP-запрос
//
// Требования:
//   - Метод: POST
//
// Ожидаемые поля в теле запроса (JSON):
//   - login: string - Логин пользователя
//   - password: string - Пароль пользователя
//
// Возможные HTTP-статусы ответа:
//   - 200 OK - при успешном удалении
//   - 400 Bad Request - при некорректном теле запроса
//   - 401 Unauthorized - при неверных учетных данных
//   - 405 Method Not Allowed - при неправильном методе запроса
//   - 422 Unprocessable Entity - при ошибке парсинга JSON
//   - 500 Internal Server Error - при внутренних ошибках сервера
func (h *Handlers) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	if r.ContentLength == 0 {
		http.Error(w, "пустое тело запроса", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "некорректный запрос", http.StatusUnprocessableEntity)
		return
	}

	if requestBody.Login == "" || requestBody.Password == "" {
		http.Error(w, "некорректный запрос", http.StatusBadRequest)
		return
	}

	id, err, code := h.userManager.Delete(r.Context(), requestBody.Login, requestBody.Password)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	logger.Log.Debugf("Пользователь №%d (%s) удален", id, requestBody.Login)
}

// AddExpressionHandler обрабатывает HTTP-запрос на добавление нового выражения.
//
// Args:
//
//	w: http.ResponseWriter - Интерфейс для записи HTTP-ответа
//	r: *http.Request - Входящий HTTP-запрос
//
// Требования:
//   - Метод: POST
//   - Заголовок Authorization: Bearer <token>
//
// Ожидаемые поля в теле запроса (JSON):
//   - expression: string - Математическое выражение для вычисления
//
// Ответ (JSON):
//   - id: int64 - ID созданного выражения
//
// Возможные HTTP-статусы ответа:
//   - 200 OK - при успешном создании выражения
//   - 400 Bad Request - при пустом выражении
//   - 405 Method Not Allowed - при неправильном методе запроса
//   - 422 Unprocessable Entity - при ошибке парсинга JSON
//   - 500 Internal Server Error - при внутренних ошибках сервера
func (h *Handlers) AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	if r.ContentLength == 0 {
		http.Error(w, "пустое тело запроса", http.StatusBadRequest)
		return
	}

	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, _ := h.jwtManager.Validate(token)

	var requestBody models.ExpressionAdd

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "некорректный запрос", http.StatusUnprocessableEntity)
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

	// Пришлось отказаться от статуса Created (201) по следующей причине:
	// Если ошибка произойдет во время кодирования вернет статус 500.
	// Вернет 201 при успешном создании или ошибке в io.Copy.
	// К сожалению ее нельзя перехватить из-за невозможности
	// перезаписать статус тк это нарушает HTTP-спецификацию.
	// По этой же причине невозможно поймать ошибку при записывании
	// ответа в заголовок при помощи http.ResponseWriter.
	// Ниже представлен пример наиболее подходящего кода:
	//buf := new(bytes.Buffer)
	//if err := json.NewEncoder(buf).Encode(response); err != nil {
	//	http.Error(w, "ошибка при кодировании ответа в JSON", http.StatusInternalServerError)
	//	return
	//}
	//w.WriteHeader(http.StatusCreated)
	//if _, err := io.Copy(w, buf); err != nil {
	//	// Заголовки уже отправлены - можно только залогировать
	//	logger.Log.Errorf("Не удалось записать выражение в ответ: %v", err)
	//}

	response := map[string]int64{"id": id}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "ошибка при кодировании ответа в JSON", http.StatusInternalServerError)
		return
	}

	logger.Log.Debugf("Выражение №%d пользователя №%d создано", id, claims.Subject)
}

// GetExpressionsHandler обрабатывает HTTP-запрос на получение списка выражений пользователя.
//
// Args:
//
//	w: http.ResponseWriter - Интерфейс для записи HTTP-ответа
//	r: *http.Request - Входящий HTTP-запрос
//
// Требования:
//   - Метод: GET
//   - Заголовок Authorization: Bearer <token> - JWT-токен аутентификации
//
// Ответ (JSON):
//   - expressions: []models.ExpressionResponse - Массив выражений пользователя
//
// Возможные HTTP-статусы ответа:
//   - 200 OK - при успешном получении списка
//   - 404 Not Found если выражения не найдены
//   - 405 Method Not Allowed - при неправильном методе запроса
//   - 500 Internal Server Error - при внутренних ошибках сервера
func (h *Handlers) GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

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
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "ошибка при кодировании ответа в JSON", http.StatusInternalServerError)
		return
	}

	logger.Log.Debugf("Список выражений пользователя №%d отправлен", claims.Subject)
}

// GetExpressionHandler обрабатывает HTTP-запрос на получение конкретного выражения по ID.
//
// Args:
//
//	w: http.ResponseWriter - Интерфейс для записи HTTP-ответа
//	r: *http.Request - Входящий HTTP-запрос с параметром ID в URL
//
// Требования:
//   - Метод: GET
//   - Заголовок Authorization: Bearer <token> - JWT-токен аутентификации
//   - Параметр URL: id - числовой идентификатор выражения
//
// Ответ (JSON):
//   - expression: models.ExpressionResponse - Данные запрошенного выражения
//
// Возможные HTTP-статусы ответа:
//   - 200 OK - при успешном получении выражения
//   - 400 Bad Request - при некорректном ID выражения
//   - 403 Forbidden - при попытке доступа к чужому выражению
//   - 404 Not Found - если выражение не найдено
//   - 405 Method Not Allowed - при неправильном методе запроса
//   - 500 Internal Server Error - при внутренних ошибках сервера
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
		http.Error(w, "не удалось перевести выражение в число", http.StatusBadRequest)
		return
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

	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "ошибка при кодировании ответа в JSON", http.StatusInternalServerError)
		return
	}

	logger.Log.Debugf("Выражение №%d пользователя №%d отправлено", expression.ID, id)
}
