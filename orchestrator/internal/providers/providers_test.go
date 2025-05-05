package providers_test

import (
	"context"
	"fmt"
	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/OinkiePie/calc_3/orchestrator/internal/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Отключаем выводы и инициализируем конфиг
	log.SetOutput(io.Discard)
	_ = config.InitConfig()
	logger.InitLogger(logger.Options{Level: 6})
}

func TestNewProviders_Success(t *testing.T) {
	ctx := context.Background()
	p, err := providers.NewProviders(ctx, ":memory:", "key")

	require.NoError(t, err)
	require.NotNil(t, p)

	assert.NotNil(t, p.SessionRepo)
	assert.NotNil(t, p.UserRepo)
	assert.NotNil(t, p.UserManager)
	assert.NotNil(t, p.ArgsRepo)
	assert.NotNil(t, p.DepsRepo)
	assert.NotNil(t, p.TaskRepo)
	assert.NotNil(t, p.ExprRepo)
	assert.NotNil(t, p.ExprManager)
	assert.NotNil(t, p.JWTManager)
	assert.NotNil(t, p.DB)

	err = p.DB.CloseDB()
	assert.NoError(t, err)
}

func TestNewProviders_DBError(t *testing.T) {
	ctx := context.Background()

	_, err := providers.NewProviders(ctx, "/invalid/path/to/db", "key")

	require.Error(t, err)
}

func TestProviders_Close(t *testing.T) {
	ctx := context.Background()
	p, err := providers.NewProviders(ctx, ":memory:", "key")
	require.NoError(t, err)

	err = p.DB.CloseDB()
	assert.NoError(t, err)

	err = p.DB.DB.Ping()
	assert.Error(t, err)
}

func TestNewProviders_EmptyJWTKey(t *testing.T) {
	ctx := context.Background()
	_, err := providers.NewProviders(ctx, ":memory:", "")

	require.NoError(t, err)
}

func TestProviders_Integration(t *testing.T) {
	dbPath := fmt.Sprintf("test_db_%s.db", time.Now().Format("20060102150405"))
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			t.Logf("failed to remove test db file: %v", err)
		}
	}()

	ctx := context.Background()
	p, err := providers.NewProviders(ctx, dbPath, "key")
	require.NoError(t, err)
	defer p.DB.CloseDB()

	_, err = p.DB.DB.Exec("SELECT 1")
	assert.NoError(t, err)

	token, _, _, err := p.JWTManager.Generate(1)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := p.JWTManager.Validate(token)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), claims.Subject)
}
