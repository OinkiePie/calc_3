package main

import (
	"context"
	"fmt"
	"github.com/OinkiePie/calc_3/orchestrator/internal/providers"
	"net/http"
	"time"

	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/orchestrator/internal/router"
	"github.com/OinkiePie/calc_3/pkg/initializer"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"github.com/OinkiePie/calc_3/pkg/shutdown"
	"github.com/rs/cors"
)

// Orchestrator представляет собой сервис оркестратора.
type Orchestrator struct {
	errChan  chan error   // Канал для отправки ошибок, возникающих в сервисе.
	server   *http.Server // Указатель на структуру http.Server, управляющую HTTP-сервером.
	addr     string       // Адрес, на котором прослушивает HTTP-сервер.
	provider *providers.Providers
}

// NewOrchestrator создает новый экземпляр сервиса оркестратора.
//
// Args:
//
//	errChan: chan error - Канал для отправки ошибок, возникающих при инициализации или работе сервиса.
//
// Returns:
//
//	*Orchestrator - Указатель на новый экземпляр структуры Orchestrator.
func NewOrchestrator(errChan chan error) (*Orchestrator, error) {
	addr := fmt.Sprintf("%s:%d", config.Cfg.Services.Orchestrator.ORCHESTRATOR_ADDR, config.Cfg.Services.Orchestrator.ORCHESTRATOR_PORT)

	ctx := context.TODO()
	provider, err := providers.NewProviders(ctx, config.Cfg.Services.Orchestrator.DATABASE, config.Cfg.Middleware.SECRET_KEY)
	if err != nil {
		return nil, err
	}
	muxRouter := router.NewOrchestratorRouter(provider)

	c := cors.New(cors.Options{
		AllowedOrigins:   config.Cfg.Middleware.AllowOrigin,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	routerHttp := c.Handler(muxRouter)
	// Создаем экземпляр структуры http.Server, указывая адрес и обработчик
	srv := &http.Server{
		Addr:    addr,
		Handler: routerHttp,
	}

	return &Orchestrator{errChan: errChan, server: srv, addr: addr, provider: provider}, nil
}

// Start запускает HTTP-сервер в отдельной горутине. Если во время запуска
// возникает ошибка, она отправляется в канал ошибок.
func (o *Orchestrator) Start() {
	// Запускаем сервер в отдельной горутине, чтобы не блокировать основной поток выполнения.
	go func() {
		// Запускаем прослушивание входящих соединений на указанном адресе.
		if err := o.server.ListenAndServe(); err != http.ErrServerClosed {
			// Если при запуске сервера произошла ошибка, отправляем её в канал ошибок.
			o.errChan <- err
		}
	}()
}

// Stop останавливает HTTP-сервер. Он использует контекст с таймаутом, чтобы
// гарантировать, что остановка не займет слишком много времени.
func (o *Orchestrator) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := o.server.Shutdown(ctx)
	if err != nil {
		logger.Log.Errorf("Ошибка при остановке сервиса Оркестратор")
	}
	err = o.provider.DB.DB.Close()
	if err != nil {
		logger.Log.Errorf("Ошибка при отключении базы данных")
	}
}

// Запуск сервиса оркестратора
func main() {
	// Инициализация конфига и логгера
	initializer.Init()

	errChan := make(chan error, 1)

	orchestratorService, err := NewOrchestrator(errChan)
	if err != nil {
		logger.Log.Fatalf("Ошибка при создании оркестратораЖ %v", err)
	}

	go func() {
		logger.Log.Debugf("Запуск сервиса Оркестратор...")
		orchestratorService.Start()
		logger.Log.Infof("Cервис Оркестратор запущен на %s", orchestratorService.addr)
	}()

	shutdown.WaitForShutdown(errChan, "Orchestrator", orchestratorService)
}
