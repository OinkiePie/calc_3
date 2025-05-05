package managers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories"
	"github.com/OinkiePie/calc_3/pkg/jwt_manager"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// UserManager предоставляет методы для управления пользователями, включая регистрацию, аутентификацию и удаление.
type UserManager struct {
	db          *sql.DB // Подключение к базе данных
	sessionRepo repositories.SessionRepositoryInterface
	userRepo    repositories.UserRepositoryInterface // Репозиторий для работы с данными пользователей
	jwtManager  jwt_manager.JWTManagerInterface      // Менеджер для работы с JWT-токенами
}

// NewUserManager создает новый экземпляр UserManager.
//
// Args:
//
//	userRepo: *repositories.UserRepository - Репозиторий пользователей
//	jwtManager: *jwt.JWTManager - Менеджер JWT-токенов
//
// Returns:
//
//	*UserManager - Новый экземпляр менеджера пользователей
func NewUserManager(
	db *sql.DB,
	sessionRepo repositories.SessionRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	jwtManager jwt_manager.JWTManagerInterface,
) *UserManager {
	return &UserManager{
		db:          db,
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		jwtManager:  jwtManager,
	}
}

// Register регистрирует нового пользователя в системе.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения.
//	login: string - Логин пользователя.
//	password: string - Пароль пользователя (в открытом виде).
//
// Returns:
//
//	int64 - ID зарегистрированного пользователя.
//	error - Ошибка выполнения.
//	int - HTTP статус код:
//		- 201 Created при успешном создании
//		- 409 Conflict при дубликате логина
//		- 500 Internal Server Error при других ошибках
func (m *UserManager) Register(ctx context.Context, login, password string) (int64, error, int) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("не удалось хешировать пароль: %w", err), http.StatusInternalServerError
	}

	user, err, code := m.userRepo.CreateUser(ctx, login, string(hashedPassword))
	if err != nil {
		return 0, err, code
	}

	return user.ID, nil, http.StatusCreated
}

// Login выполняет аутентификацию пользователя и генерирует JWT-токен.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения.
//	login: string - Логин пользователя.
//	password: string - Пароль пользователя (в открытом виде).
//
// Returns:
//
//	string - Сгенерированный JWT-токен.
//	int64 - ID аутентифицированного пользователя.
//	error - Ошибка выполнения.
//	int - HTTP статус код:
//		- 200 OK при успешной аутентификации
//		- 401 Unauthorized если пользователь не найден
//		- 500 Internal Server Error при ошибках
func (m *UserManager) Login(ctx context.Context, login, password string) (string, int64, error, int) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return "", 0, fmt.Errorf("не удалось начать вход пользователя: %w", err), http.StatusInternalServerError
	}
	defer tx.Rollback()

	user, err, _ := m.userRepo.ReadUserByLogin(ctx, login)
	if err != nil {
		return "", 0, errors.New("неверный логин или пароль"), http.StatusUnauthorized
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", 0, errors.New("неверный логин или пароль"), http.StatusUnauthorized
	}

	token, jti, exp, err := m.jwtManager.Generate(user.ID)
	if err != nil {
		return "", 0, fmt.Errorf("не удалось сгенерировать токен: %w", err), http.StatusInternalServerError
	}

	err, _ = m.sessionRepo.CreateSession(ctx, tx, jti, user.ID, exp)
	if err != nil {
		return "", 0, err, http.StatusUnauthorized
	}

	if err = tx.Commit(); err != nil {
		return "", 0, fmt.Errorf("вход пользователя не удался: %w", err), http.StatusInternalServerError
	}

	return token, user.ID, nil, http.StatusOK
}

// Logout удаляет сессию из базы данных.
//
// Args:
//
//	ctx: context.Context - Контекст для контроля выполнения запроса.
//	jti: string - Идентификатор сессии.
//
// Returns:
//
//	*models.Session - Указатель на сессию.
//	error - Ошибка выполнения операции.
//	int - HTTP-статус код результата операции:
//		- 200 OK при успешном получении
//		- 500 Internal Server Error при ошибках
func (m *UserManager) Logout(ctx context.Context, jti string) (error, int) {
	return m.sessionRepo.DeleteSession(ctx, jti)
}

// Delete удаляет учетную запись пользователя после проверки пароля.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения.
//	login: string - Логин пользователя.
//	password: string - Пароль пользователя (в открытом виде).
//
// Returns:
//
//	int64 - ID удаленного пользователя.
//	error - Ошибка выполнения.
//	int - HTTP статус код:
//		- 200 OK при успешном удалении
//		- 401 Unauthorized при несовпадении паролей
//		- 500 Internal Server Error при ошибках
func (m *UserManager) Delete(ctx context.Context, login, password string) (int64, error, int) {
	user, err, code := m.userRepo.ReadUserByLogin(ctx, login)
	if err != nil {
		return 0, err, http.StatusUnauthorized
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return 0, errors.New("пароли не совпадают"), http.StatusUnauthorized
	}

	if err, code = m.userRepo.DeleteUser(ctx, user.ID); err != nil {
		return 0, err, code
	}

	return user.ID, nil, http.StatusOK
}

// SessionExists проверяет существование сессии.
//
// Args:
//
//	ctx: context.Context - Контекст выполнения.
//	jti: string - Идентификатор сессии.
//
// Returns:
//
//	error - Ошибка выполнения.
//	bool - Индикатор существования.
func (m *UserManager) SessionExists(ctx context.Context, jti string) (error, bool) {
	session, err, _ := m.sessionRepo.ReadSession(ctx, jti)
	if session == nil {
		if err != nil {
			return err, false
		}
		return nil, false
	}
	return nil, true
}
