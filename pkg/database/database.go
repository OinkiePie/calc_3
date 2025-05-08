package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/pkg/logger"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"time"
)

type DataBase struct {
	DB  *sql.DB
	ctx context.Context
	dsn string
}

// NewDB создаёт подключение и таблицы (если их нет).
func NewDB(ctx context.Context, dsn string) (*DataBase, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть базу данных: %w", err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, fmt.Errorf("не удалось включить foregin ключи: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось пингануть базу данных: %w", err)
	}

	database := &DataBase{
		DB:  db,
		ctx: ctx,
		dsn: dsn,
	}

	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("не удалось создать таблицы: %w", err)
	}

	startSessionCleaner(db)

	return database, nil
}

// createTables — приватный метод для создания таблиц.
func (db *DataBase) createTables() error {
	const (
		usersTable = `
		CREATE TABLE IF NOT EXISTS users(
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			login TEXT UNIQUE NOT NULL,
			pas TEXT NOT NULL
		);`

		sessionsTable = `
		CREATE TABLE IF NOT EXISTS sessions(
			id VARCHAR(36) PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires INTEGER NOT NULL,
		                                
		    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);`

		expressionsTable = `
		CREATE TABLE IF NOT EXISTS expressions(
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			user_id INTEGER NOT NULL,
			expression_string TEXT NOT NULL,
			status TEXT CHECK(status IN ('pending', 'processing', 'completed', 'error')) DEFAULT 'pending',
			result REAL,
			error TEXT DEFAULT '',	
		    
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);`

		tasksTable = `
		CREATE TABLE IF NOT EXISTS tasks(
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			expression_id INTEGER NOT NULL,
			operation TEXT NOT NULL CHECK(operation IN ('+', '-', '*', '/', '^', 'u-')),
		    result REAL,
			status TEXT CHECK(status IN ('pending', 'processing', 'completed', 'error')) DEFAULT 'pending',
		    
			FOREIGN KEY (expression_id) REFERENCES expressions(id) ON DELETE CASCADE
		);`

		tasksArgsTable = `
		CREATE TABLE IF NOT EXISTS task_args (
			task_id INTEGER PRIMARY KEY NOT NULL,
			first REAL,
			second REAL,
			
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
		);`

		tasksDependenciesTable = `
		CREATE TABLE IF NOT EXISTS task_deps (
			task_id INTEGER PRIMARY KEY NOT NULL,
			first INTEGER,
			second INTEGER,
			
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
		);`
	)

	if _, err := db.DB.ExecContext(db.ctx, usersTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	if _, err := db.DB.ExecContext(db.ctx, sessionsTable); err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	if _, err := db.DB.ExecContext(db.ctx, expressionsTable); err != nil {
		return fmt.Errorf("failed to create expressions table: %w", err)
	}

	if _, err := db.DB.ExecContext(db.ctx, tasksTable); err != nil {
		return fmt.Errorf("failed to create tasks table: %w", err)
	}

	if _, err := db.DB.ExecContext(db.ctx, tasksArgsTable); err != nil {
		return fmt.Errorf("failed to create tasks args table: %w", err)
	}

	if _, err := db.DB.ExecContext(db.ctx, tasksDependenciesTable); err != nil {
		return fmt.Errorf("failed to create tasks deps table: %w", err)
	}

	return nil
}

func (db *DataBase) CloseDB() error {
	if err := db.DB.Close(); err != nil {
		return fmt.Errorf("не удалось отключить базу данных: %w", err)
	}
	return nil
}

func (db *DataBase) DeleteDB() error {
	if db.dsn == "" {
		return fmt.Errorf("пустое название базы данных")
	}
	if err := db.DB.Close(); err != nil {
		return fmt.Errorf("не удалось отключить базу данных: %w", err)
	}

	err := os.Remove(db.dsn)
	if err != nil {
		return fmt.Errorf("не удалось удалить файл базы данных: %w", err)
	}
	return nil
}

func (db *DataBase) ClearDB() error {
	tables := []string{"users", "expressions", "tasks", "task_args", "task_deps", "sessions"}

	// Временное отключение внешних ключей
	_, err := db.DB.ExecContext(db.ctx, "PRAGMA foreign_keys = OFF")
	if err != nil {
		return fmt.Errorf("не удолось отключить внешние ключи: %w", err)
	}
	defer func() {
		// Повторное включение внешних ключей
		_, err := db.DB.ExecContext(db.ctx, "PRAGMA foreign_keys = ON")
		if err != nil {
			logger.Log.Warnf("Не удалось повторно включитить внешние ключи: %v", err)
		}
	}()

	// Очистка всех таблиц
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		_, err := db.DB.ExecContext(db.ctx, query)
		if err != nil {
			return fmt.Errorf("не удалось очистить таблицу %s: %w", table, err)
		}
	}

	// Сброс авто-прибавления всех таблиц
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name = '%s'", table)
		_, err := db.DB.ExecContext(db.ctx, query)
		if err != nil {
			logger.Log.Warnf("Не удалось сбросить авто-прибавления таблицы %s: %v\n", table, err)
		}
	}

	return nil
}

// startSessionCleaner запускает фоновую очистку сессий
func startSessionCleaner(db *sql.DB) {
	go func() {
		ticker := time.NewTicker(time.Duration(config.Cfg.Middleware.SESSION_CLEAR_MIN) * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			cleanExpiredSessions(db)
		}
	}()
}

// cleanExpiredSessions удаляет истекшие сессии
func cleanExpiredSessions(db *sql.DB) {
	_, err := db.Exec(`
		DELETE FROM sessions 
		WHERE expires < strftime('%s', 'now')`)

	if err != nil {
		log.Printf("Ошибка при очистке сессий: %v", err)
		return
	}
}
