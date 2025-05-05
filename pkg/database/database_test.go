package database_test

import (
	"context"
	"fmt"
	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/pkg/database"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"os"
	"testing"
	"time"
)

func init() {
	// Отключаем выводы и инициализируем конфиг
	log.SetOutput(io.Discard)
	_ = config.InitConfig()
	logger.InitLogger(logger.Options{Level: 6})
	config.Cfg.Middleware.SESSION_CLEAR_MIN = 10
}

func TestDatabaseIntegration(t *testing.T) {

	ctx := context.Background()

	t.Run("NewDB success", func(t *testing.T) {
		db, err := database.NewDB(ctx, ":memory:")
		require.NoError(t, err)
		defer db.CloseDB()

		assert.NotNil(t, db.DB)
		assert.NoError(t, db.DB.Ping())
	})

	t.Run("CreateTables", func(t *testing.T) {
		db, err := database.NewDB(ctx, ":memory:")
		require.NoError(t, err)
		defer db.CloseDB()

		tables := []string{"users", "sessions", "expressions", "tasks"}
		for _, table := range tables {
			_, err := db.DB.ExecContext(ctx, fmt.Sprintf("SELECT 1 FROM %s LIMIT 1", table))
			assert.NoError(t, err, "table %s should exist", table)
		}
	})

	t.Run("ClearDB", func(t *testing.T) {
		db, err := database.NewDB(ctx, ":memory:")
		require.NoError(t, err)
		defer db.CloseDB()

		_, err = db.DB.ExecContext(ctx, "INSERT INTO users(login, pas) VALUES('test', 'pass')")
		require.NoError(t, err)

		err = db.ClearDB()
		assert.NoError(t, err)

		var count int
		err = db.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("DeleteDB", func(t *testing.T) {
		dbPath := fmt.Sprintf("test_db_%s.db", time.Now().Format("20060102150405"))
		db, err := database.NewDB(ctx, dbPath)
		require.NoError(t, err)

		err = db.DeleteDB()
		assert.NoError(t, err)

		_, err = os.Stat(dbPath)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestNewDBErrors(t *testing.T) {
	t.Run("Invalid DSN", func(t *testing.T) {
		_, err := database.NewDB(context.Background(), "invalid://dsn")
		assert.Error(t, err)
	})
}
