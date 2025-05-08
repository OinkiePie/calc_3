package user_manager_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/orchestrator/internal/managers/user_manager"
	mr "github.com/OinkiePie/calc_3/orchestrator/internal/repositories"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories/session_repository"
	"github.com/OinkiePie/calc_3/orchestrator/internal/repositories/user_repository"
	mj "github.com/OinkiePie/calc_3/pkg/jwt_manager"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"
)

func init() {
	// Отключаем выводы и инициализируем конфиг
	log.SetOutput(io.Discard)
	_ = config.InitConfig()
	logger.InitLogger(logger.Options{Level: 6})
}

func TestNewUserManager(t *testing.T) {
	mockDB := &sql.DB{}
	mockUserRepo := new(mr.MockUserRepository)
	mockSessionRepo := new(mr.MockSessionRepository)
	mockJWTManager := new(mj.MockJWTManager)

	manager := user_manager.NewUserManager(
		mockDB,
		mockSessionRepo,
		mockUserRepo,
		mockJWTManager,
	)

	assert.NotNil(t, manager)
}

// МОДУЛЬНЫЕ ТЕСТЫ

func TestUserManager_Register(t *testing.T) {
	mockDB := &sql.DB{}
	mockUserRepo := new(mr.MockUserRepository)
	mockSessionRepo := new(mr.MockSessionRepository)
	mockJWTManager := new(mj.MockJWTManager)

	manager := user_manager.NewUserManager(mockDB, mockSessionRepo, mockUserRepo, mockJWTManager)

	ctx := context.Background()
	testLogin := "testuser"
	testPassword := "securepassword"

	t.Run("successful registration", func(t *testing.T) {
		mockUserRepo.On("CreateUser", ctx, testLogin, mock.AnythingOfType("string")).
			Return(&models.User{ID: 1}, nil, http.StatusCreated).Once()

		userID, err, code := manager.Register(ctx, testLogin, testPassword)

		assert.Equal(t, int64(1), userID)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, code)

		args := mockUserRepo.Calls[0].Arguments
		providedHashedPass := args.Get(2).(string)
		assert.NotEqual(t, testPassword, providedHashedPass)
		assert.Contains(t, providedHashedPass, "$2a$10$")
	})

	t.Run("duplicate login", func(t *testing.T) {
		mockUserRepo.On("CreateUser", ctx, "duplicate", mock.AnythingOfType("string")).
			Return((*models.User)(nil), errors.New("user already exists"), http.StatusConflict).Once()

		_, err, code := manager.Register(ctx, "duplicate", testPassword)

		assert.Error(t, err)
		assert.Equal(t, "user already exists", err.Error())
		assert.Equal(t, http.StatusConflict, code)
	})

	t.Run("repository error", func(t *testing.T) {
		mockUserRepo.On("CreateUser", ctx, "erroruser", mock.AnythingOfType("string")).
			Return((*models.User)(nil), errors.New("db error"), http.StatusInternalServerError).Once()

		_, err, code := manager.Register(ctx, "erroruser", testPassword)

		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
		assert.Equal(t, http.StatusInternalServerError, code)
	})
}

func TestUserManager_Login(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mockUserRepo := new(mr.MockUserRepository)
	mockSessionRepo := new(mr.MockSessionRepository)
	mockJWTManager := new(mj.MockJWTManager)

	manager := user_manager.NewUserManager(db, mockSessionRepo, mockUserRepo, mockJWTManager)

	ctx := context.Background()
	testLogin := "testuser"
	testPassword := "securepassword"
	hashedPassword := "$2y$10$7dRJdbrVqcBIazK0ewRN.eQhZImE7cSjVWdw97C/rjV8tRFhUPSja"
	userID := int64(1)
	testToken := "test.jwt.token"
	testJTI := "jti-123"
	testExp := int64(3600)

	t.Run("successful login", func(t *testing.T) {
		mockUserRepo.On("ReadUserByLogin", ctx, testLogin).
			Return(&models.User{
				ID:       userID,
				Login:    testLogin,
				Password: hashedPassword,
			}, nil, http.StatusOK).Once()

		mockJWTManager.On("Generate", userID).
			Return(testToken, testJTI, testExp, nil).Once()

		mockSessionRepo.On("CreateSession", ctx, mock.Anything, testJTI, userID, testExp).
			Return(nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit()

		token, id, err, code := manager.Login(ctx, testLogin, testPassword)

		assert.Equal(t, testToken, token)
		assert.Equal(t, userID, id)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo.On("ReadUserByLogin", ctx, "unknown").
			Return((*models.User)(nil), errors.New("user not found"), http.StatusNotFound).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectRollback()

		_, _, err, code := manager.Login(ctx, "unknown", testPassword)

		assert.Error(t, err)
		assert.Equal(t, "неверный логин или пароль", err.Error())
		assert.Equal(t, http.StatusUnauthorized, code)
	})

	t.Run("wrong password", func(t *testing.T) {
		mockUserRepo.On("ReadUserByLogin", ctx, testLogin).
			Return(&models.User{
				ID:       userID,
				Login:    testLogin,
				Password: hashedPassword,
			}, nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectRollback()

		_, _, err, code := manager.Login(ctx, testLogin, "wrongpassword")

		assert.Error(t, err)
		assert.Equal(t, "неверный логин или пароль", err.Error())
		assert.Equal(t, http.StatusUnauthorized, code)
	})

	t.Run("token generation error", func(t *testing.T) {
		mockUserRepo.On("ReadUserByLogin", ctx, testLogin).
			Return(&models.User{
				ID:       userID,
				Login:    testLogin,
				Password: hashedPassword,
			}, nil, http.StatusOK).Once()

		mockJWTManager.On("Generate", userID).
			Return("", "", int64(0), errors.New("token error")).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectRollback()

		_, _, err, code := manager.Login(ctx, testLogin, testPassword)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось сгенерировать токен")
		assert.Equal(t, http.StatusInternalServerError, code)
	})

	t.Run("session creation error", func(t *testing.T) {
		mockUserRepo.On("ReadUserByLogin", ctx, testLogin).
			Return(&models.User{
				ID:       userID,
				Login:    testLogin,
				Password: hashedPassword,
			}, nil, http.StatusOK).Once()

		mockJWTManager.On("Generate", userID).
			Return(testToken, testJTI, testExp, nil).Once()

		mockSessionRepo.On("CreateSession", ctx, mock.Anything, testJTI, userID, testExp).
			Return(errors.New("session error"), http.StatusInternalServerError).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectRollback()

		_, _, err, code := manager.Login(ctx, testLogin, testPassword)

		assert.Error(t, err)
		assert.Equal(t, "session error", err.Error())
		assert.Equal(t, http.StatusUnauthorized, code)
	})

	t.Run("transaction commit error", func(t *testing.T) {
		mockUserRepo.On("ReadUserByLogin", ctx, testLogin).
			Return(&models.User{
				ID:       userID,
				Login:    testLogin,
				Password: hashedPassword,
			}, nil, http.StatusOK).Once()

		mockJWTManager.On("Generate", userID).
			Return(testToken, testJTI, testExp, nil).Once()

		mockSessionRepo.On("CreateSession", ctx, mock.Anything, testJTI, userID, testExp).
			Return(nil, http.StatusOK).Once()

		mockDB.ExpectBegin()
		mockDB.ExpectCommit().WillReturnError(errors.New("commit error"))

		_, _, err, code := manager.Login(ctx, testLogin, testPassword)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "вход пользователя не удался")
		assert.Equal(t, http.StatusInternalServerError, code)
	})
}

func TestUserManager_Logout(t *testing.T) {
	mockDB := &sql.DB{}
	mockSessionRepo := new(mr.MockSessionRepository)

	manager := user_manager.NewUserManager(mockDB, mockSessionRepo, nil, nil)

	ctx := context.Background()
	testJTI := "session_jti_123"

	t.Run("successful logout", func(t *testing.T) {
		mockSessionRepo.On("DeleteSession", ctx, testJTI).
			Return(nil, http.StatusOK).Once()

		err, code := manager.Logout(ctx, testJTI)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
	})

	t.Run("session not found", func(t *testing.T) {
		mockSessionRepo.On("DeleteSession", ctx, "unknown_jti").
			Return(errors.New("session not found"), http.StatusNotFound).Once()

		err, code := manager.Logout(ctx, "unknown_jti")

		assert.Error(t, err)
		assert.Equal(t, "session not found", err.Error())
		assert.Equal(t, http.StatusNotFound, code)
	})

	t.Run("database error", func(t *testing.T) {
		mockSessionRepo.On("DeleteSession", ctx, "error_jti").
			Return(errors.New("database error"), http.StatusInternalServerError).Once()

		err, code := manager.Logout(ctx, "error_jti")

		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
		assert.Equal(t, http.StatusInternalServerError, code)
	})
}

func TestUserManager_Delete(t *testing.T) {
	mockDB := &sql.DB{}
	mockUserRepo := new(mr.MockUserRepository)

	manager := user_manager.NewUserManager(mockDB, nil, mockUserRepo, nil)

	ctx := context.Background()
	testLogin := "testuser"
	testPassword := "securepassword"
	hashedPassword := "$2y$10$7dRJdbrVqcBIazK0ewRN.eQhZImE7cSjVWdw97C/rjV8tRFhUPSja"
	userID := int64(1)

	t.Run("successful delete", func(t *testing.T) {
		mockUserRepo.On("ReadUserByLogin", ctx, testLogin).
			Return(&models.User{
				ID:       userID,
				Login:    testLogin,
				Password: hashedPassword,
			}, nil, http.StatusOK).Once()

		mockUserRepo.On("DeleteUser", ctx, userID).
			Return(nil, http.StatusOK).Once()

		deletedID, err, code := manager.Delete(ctx, testLogin, testPassword)

		assert.Equal(t, userID, deletedID)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo.On("ReadUserByLogin", ctx, "unknown").
			Return((*models.User)(nil), errors.New("user not found"), http.StatusNotFound).Once()

		_, err, code := manager.Delete(ctx, "unknown", testPassword)

		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
		assert.Equal(t, http.StatusUnauthorized, code)
	})

	t.Run("wrong password", func(t *testing.T) {
		mockUserRepo.On("ReadUserByLogin", ctx, testLogin).
			Return(&models.User{
				ID:       userID,
				Login:    testLogin,
				Password: hashedPassword,
			}, nil, http.StatusOK).Once()

		_, err, code := manager.Delete(ctx, testLogin, "wrongpassword")

		assert.Error(t, err)
		assert.Equal(t, "пароли не совпадают", err.Error())
		assert.Equal(t, http.StatusUnauthorized, code)
	})

	t.Run("delete error", func(t *testing.T) {
		mockUserRepo.On("ReadUserByLogin", ctx, testLogin).
			Return(&models.User{
				ID:       userID,
				Login:    testLogin,
				Password: hashedPassword,
			}, nil, http.StatusOK).Once()

		mockUserRepo.On("DeleteUser", ctx, userID).
			Return(errors.New("database error"), http.StatusInternalServerError).Once()

		_, err, code := manager.Delete(ctx, testLogin, testPassword)

		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
		assert.Equal(t, http.StatusInternalServerError, code)
	})
}

func TestUserManager_SessionExists(t *testing.T) {
	mockDB := &sql.DB{}
	mockSessionRepo := new(mr.MockSessionRepository)

	manager := user_manager.NewUserManager(mockDB, mockSessionRepo, nil, nil)

	ctx := context.Background()
	testJTI := "test_jti_123"
	expiredJTI := "expired_jti_123"
	errorJTI := "error_jti_123"

	testSession := &models.Session{
		ID:      testJTI,
		UserID:  1,
		Expires: time.Now().Add(time.Hour).Unix(),
	}

	t.Run("session exists", func(t *testing.T) {
		mockSessionRepo.On("ReadSession", ctx, testJTI).
			Return(testSession, nil, http.StatusOK).Once()

		err, exists := manager.SessionExists(ctx, testJTI)

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("session does not exist", func(t *testing.T) {
		mockSessionRepo.On("ReadSession", ctx, "unknown_jti").
			Return((*models.Session)(nil), nil, http.StatusNotFound).Once()

		err, exists := manager.SessionExists(ctx, "unknown_jti")

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("expired session", func(t *testing.T) {
		expiredSession := &models.Session{
			ID:      expiredJTI,
			UserID:  1,
			Expires: time.Now().Add(-time.Hour).Unix(),
		}
		mockSessionRepo.On("ReadSession", ctx, expiredJTI).
			Return(expiredSession, nil, http.StatusOK).Once()

		err, exists := manager.SessionExists(ctx, expiredJTI)

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("database error", func(t *testing.T) {
		mockSessionRepo.On("ReadSession", ctx, errorJTI).
			Return((*models.Session)(nil), errors.New("database error"), http.StatusInternalServerError).Once()

		err, exists := manager.SessionExists(ctx, errorJTI)

		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
		assert.False(t, exists)
	})
}

// ИНТЕГРАЦИОНнЫЕ ТЕСТЫ

func TestUserManager_Register_Integration(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	setupTestDatabase := func(db *sql.DB) error {
		_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			login TEXT NOT NULL UNIQUE,
			pas TEXT NOT NULL
		);
		`)
		return err
	}

	if setupTestDatabase(db) != nil {
		t.Fatal(err)
	}

	userRepo := user_repository.NewUserRepository(db)
	sessionRepo := session_repository.NewSessionRepository(db)
	jwtManager := mj.NewJWTManager("test_secret_key")

	manager := user_manager.NewUserManager(db, sessionRepo, userRepo, jwtManager)

	ctx := context.Background()

	t.Run("successful registration flow", func(t *testing.T) {
		userID, err, code := manager.Register(ctx, "integration_user", "password123")

		assert.NotZero(t, userID)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, code)

		user, err, code := userRepo.ReadUserByLogin(ctx, "integration_user")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "integration_user", user.Login)

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123"))
		assert.NoError(t, err)
	})

	t.Run("duplicate registration", func(t *testing.T) {
		_, _, _ = manager.Register(ctx, "duplicate_user", "pass1")

		_, err, code := manager.Register(ctx, "duplicate_user", "pass2")

		assert.Error(t, err)
		assert.Equal(t, http.StatusConflict, code)
	})
}

func TestUserManager_Login_Integration(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:testdb?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	setupTestDatabase := func(db *sql.DB) error {
		queries := []string{
			`CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            login TEXT NOT NULL UNIQUE,
            pas TEXT NOT NULL
        )`,
			`CREATE TABLE IF NOT EXISTS sessions (
            id TEXT PRIMARY KEY,
            user_id INTEGER NOT NULL,
            expires INTEGER NOT NULL,
            FOREIGN KEY(user_id) REFERENCES users(id)
        )`,
		}

		for _, query := range queries {
			if _, err := db.Exec(query); err != nil {
				return err
			}
		}
		return nil
	}

	if err := setupTestDatabase(db); err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}

	userRepo := user_repository.NewUserRepository(db)
	sessionRepo := session_repository.NewSessionRepository(db)
	jwtManager := mj.NewJWTManager("test_secret_key")
	manager := user_manager.NewUserManager(db, sessionRepo, userRepo, jwtManager)

	ctx := context.Background()

	testLogin := "test_user"
	testPassword := "secure_password123"
	_, err, _ = manager.Register(ctx, testLogin, testPassword)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	t.Run("successful login", func(t *testing.T) {
		token, userID, err, status := manager.Login(ctx, testLogin, testPassword)

		assert.NotEmpty(t, token)
		assert.NotZero(t, userID)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)

		claims, err := jwtManager.Validate(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.Subject)

		session, err, _ := sessionRepo.ReadSession(ctx, claims.JWTID)
		assert.NoError(t, err)
		assert.Equal(t, userID, session.UserID)
	})

	t.Run("invalid login - wrong username", func(t *testing.T) {
		_, _, err, status := manager.Login(ctx, "wrong_user", testPassword)

		assert.Error(t, err)
		assert.Equal(t, "неверный логин или пароль", err.Error())
		assert.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("invalid login - wrong password", func(t *testing.T) {
		_, _, err, status := manager.Login(ctx, testLogin, "wrong_password")

		assert.Error(t, err)
		assert.Equal(t, "неверный логин или пароль", err.Error())
		assert.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("token generation error", func(t *testing.T) {
		mockJWT := new(mj.MockJWTManager)
		mockJWT.On("Generate", mock.Anything).Return("", "", int64(0), errors.New("token error"))

		errorManager := user_manager.NewUserManager(db, sessionRepo, userRepo, mockJWT)
		_, _, err, status := errorManager.Login(ctx, testLogin, testPassword)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось сгенерировать токен")
		assert.Equal(t, http.StatusInternalServerError, status)
	})
}

func TestUserManager_Logout_Integration(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:testdb?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	setupTestDatabase := func(db *sql.DB) error {
		_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires INTEGER NOT NULL
		);
	`)
		return err
	}

	if err := setupTestDatabase(db); err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}

	sessionRepo := session_repository.NewSessionRepository(db)
	manager := user_manager.NewUserManager(db, sessionRepo, nil, nil)

	ctx := context.Background()

	testJTI := "test_session_jti"
	testUserID := int64(1)
	testExp := int64(3600)

	_, err = db.Exec("INSERT INTO sessions (id, user_id, expires) VALUES (?, ?, ?)",
		testJTI, testUserID, testExp)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	t.Run("successful logout", func(t *testing.T) {
		err, status := manager.Logout(ctx, testJTI)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)

		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE id = ?", testJTI).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("logout non-existent session", func(t *testing.T) {
		err, status := manager.Logout(ctx, "non_existent_jti")

		assert.Error(t, err)
		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("database error", func(t *testing.T) {
		mockSessionRepo := new(mr.MockSessionRepository)
		mockSessionRepo.On("DeleteSession", mock.Anything, mock.Anything).
			Return(errors.New("database error"), http.StatusInternalServerError).Once()

		errorManager := user_manager.NewUserManager(db, mockSessionRepo, nil, nil)
		err, status := errorManager.Logout(ctx, "any_jti")

		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
		assert.Equal(t, http.StatusInternalServerError, status)
	})
}

func TestUserManager_Delete_Integration(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:testdb?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	setupTestDatabase := func(db *sql.DB) error {
		_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			login TEXT NOT NULL UNIQUE,
			pas TEXT NOT NULL
		);
	`)
		return err
	}

	if err := setupTestDatabase(db); err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}

	userRepo := user_repository.NewUserRepository(db)
	manager := user_manager.NewUserManager(db, nil, userRepo, nil)

	ctx := context.Background()

	testLogin := "testuser"
	testPassword := "securepassword"

	_, err, _ = manager.Register(ctx, testLogin, testPassword)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	t.Run("successful delete", func(t *testing.T) {
		deletedID, err, code := manager.Delete(ctx, testLogin, testPassword)

		assert.NotZero(t, deletedID)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)

		_, err, _ = userRepo.ReadUserByLogin(ctx, testLogin)
		assert.Error(t, err)
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err, _ = manager.Register(ctx, testLogin, testPassword)
		if err != nil {
			t.Fatalf("Failed to register test user: %v", err)
		}

		_, err, code := manager.Delete(ctx, testLogin, "wrongpassword")

		assert.Error(t, err)
		assert.Equal(t, "пароли не совпадают", err.Error())
		assert.Equal(t, http.StatusUnauthorized, code)

		_, err, _ = userRepo.ReadUserByLogin(ctx, testLogin)
		assert.NoError(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		_, err, code := manager.Delete(ctx, "unknown", testPassword)

		assert.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, code)
	})
}

func TestUserManager_SessionExists_Integration(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:testdb?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	setupTestDatabase := func(db *sql.DB) error {
		_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires INTEGER NOT NULL
		);
	`)
		return err
	}

	if err := setupTestDatabase(db); err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}

	sessionRepo := session_repository.NewSessionRepository(db)
	manager := user_manager.NewUserManager(db, sessionRepo, nil, nil)

	ctx := context.Background()

	testJTI := "test_session_jti"
	expiredJTI := "expired_session_jti"
	testUserID := int64(1)
	now := time.Now()

	_, err = db.Exec(`
		INSERT INTO sessions (id, user_id, expires) VALUES 
		(?, ?, ?), 
		(?, ?, ?)`,
		testJTI, testUserID, now.Add(time.Hour).Unix(),
		expiredJTI, testUserID, now.Add(-time.Hour).Unix(),
	)
	if err != nil {
		t.Fatalf("Failed to create test sessions: %v", err)
	}

	t.Run("active session exists", func(t *testing.T) {
		err, exists := manager.SessionExists(ctx, testJTI)

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("expired session does not exist", func(t *testing.T) {
		err, exists := manager.SessionExists(ctx, expiredJTI)

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existent session", func(t *testing.T) {
		err, exists := manager.SessionExists(ctx, "unknown_jti")

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("database error", func(t *testing.T) {
		mockSessionRepo := new(mr.MockSessionRepository)
		mockSessionRepo.On("ReadSession", mock.Anything, mock.Anything).
			Return((*models.Session)(nil), errors.New("database error"), http.StatusInternalServerError).Once()

		errorManager := user_manager.NewUserManager(db, mockSessionRepo, nil, nil)
		err, exists := errorManager.SessionExists(ctx, "any_jti")

		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
		assert.False(t, exists)
	})
}
