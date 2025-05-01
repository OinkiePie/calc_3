package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/OinkiePie/calc_3/orchestrator/internal/grpcservice"
	"github.com/OinkiePie/calc_3/orchestrator/internal/providers"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"time"

	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/orchestrator/internal/router"
	"github.com/OinkiePie/calc_3/pkg/initializer"
	"github.com/OinkiePie/calc_3/pkg/logger"
	pb "github.com/OinkiePie/calc_3/pkg/proto"
	"github.com/OinkiePie/calc_3/pkg/shutdown"
	"github.com/rs/cors"
)

// Orchestrator представляет собой сервис оркестратора.
type Orchestrator struct {
	errChan    chan error   // Канал для отправки ошибок, возникающих в сервисе.
	serverHTTP *http.Server // Указатель на структуру http.Server, управляющую HTTP-сервером.
	serverGRPC *grpc.Server
	addrGRPC   string // Адрес, на котором прослушивает gRPC-сервер.
	addrHTTP   string // Адрес, на котором прослушивает HTTP-сервер.
	provider   *providers.Providers
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
	addrHTTP := fmt.Sprintf("%s:%d", config.Cfg.Services.Orchestrator.ORCHESTRATOR_ADDR, config.Cfg.Services.Orchestrator.ORCHESTRATOR_HTTP_PORT)

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
	serverHTTP := &http.Server{
		Addr:    addrHTTP,
		Handler: routerHttp,
	}

	addrGRPC := fmt.Sprintf("%s:%d", config.Cfg.Services.Orchestrator.ORCHESTRATOR_ADDR, config.Cfg.Services.Orchestrator.ORCHESTRATOR_GRPC_PORT)

	serverGRPC := grpc.NewServer()
	serviceGRPC := grpcservice.NewOrchestratorGRPCServer(provider)
	pb.RegisterOrchestratorServiceServer(serverGRPC, serviceGRPC)

	return &Orchestrator{
		errChan:    errChan,
		serverHTTP: serverHTTP,
		serverGRPC: serverGRPC,
		addrGRPC:   addrGRPC,
		addrHTTP:   addrHTTP,
		provider:   provider,
	}, nil
}

// Start запускает HTTP-сервер в отдельной горутине. Если во время запуска
// возникает ошибка, она отправляется в канал ошибок.
func (o *Orchestrator) Start() {
	go func() {
		if err := o.serverHTTP.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			err = fmt.Errorf("ошибка запуска http сервера: %s", err)
			o.errChan <- err
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", o.addrGRPC)
		if err != nil {
			err = fmt.Errorf("ошибка запуска слушателя tcp: %s", err)
			o.errChan <- err
		}
		if err = o.serverGRPC.Serve(lis); err != nil {
			err = fmt.Errorf("ошибка обслуживания gRPC: %s", err)
			o.errChan <- err
		}
	}()
}

// Stop останавливает сервис. Он использует контекст с таймаутом, чтобы
// гарантировать, что остановка не займет слишком много времени.
func (o *Orchestrator) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := o.serverHTTP.Shutdown(ctx); err != nil {
		logger.Log.Errorf("ошибка при отключении HTTP сервера: %v", err)
	} else {
		logger.Log.Debugf("HTTP сервер успешно остановлен")
	}

	stopped := make(chan struct{})
	go func() {
		o.serverGRPC.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		logger.Log.Debugf("gRPC сервер успешно остановлен")
	case <-ctx.Done():
		o.serverGRPC.Stop() // Принудительная остановка по таймауту
		logger.Log.Warnf("gRPC сервер принудительно остановлен из-за тайм-аута")
	}

	if err := o.provider.DB.DB.Close(); err != nil {
		logger.Log.Errorf("Ошибка при закрытии соединения с базой данных: %v", err)
	} else {
		logger.Log.Debugf("Соединение с базой данных успешно закрыто")
	}
}

func main() {
	// Инициализация конфига и логгера
	initializer.Init()

	errChan := make(chan error, 1)
	logger.Log.Debugf("Начало создания сервиса Оркестратор...")
	orchestratorService, err := NewOrchestrator(errChan)
	if err != nil {
		logger.Log.Fatalf("Ошибка при создании оркестратора: %v", err)
	}

	go func() {
		logger.Log.Debugf("Запуск сервиса Оркестратор...")
		orchestratorService.Start()
		logger.Log.Infof("HTTP Cервис Оркестратора запущен на %s", orchestratorService.addrHTTP)
	}()

	shutdown.WaitForShutdown(errChan, "Оркестратор", orchestratorService)
}
