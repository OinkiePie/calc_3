package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"

	"github.com/OinkiePie/calc_3/agent/internal/workers"
	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/pkg/initializer"
	"github.com/OinkiePie/calc_3/pkg/logger"
	pb "github.com/OinkiePie/calc_3/pkg/proto"
	"github.com/OinkiePie/calc_3/pkg/shutdown"
)

// Agent представляет собой сервис агента, отвечающий за выполнение задач.
type Agent struct {
	addr        string
	errChan     chan error         // Канал для отправки ошибок, возникающих в сервисе.
	stopWorkers context.CancelFunc // Функция для остановки всех воркеров.
	connection  *grpc.ClientConn
	workersCtx  context.Context   // Контекст, используемый воркерами для выполнения задач.
	workers     []*workers.Worker // Список воркеров, выполняющих задачи.
	power       int               // Вычислительная мощность агента (количество воркеров).
	wg          *sync.WaitGroup   // WaitGroup для ожидания завершения всех воркеров.
}

// NewAgent создает новый экземпляр сервиса агента.
//
// Args:
//
//	errChan: chan error - Канал для отправки ошибок, возникающих при работе сервиса.
//
// Returns:
//
//	*Agent - Указатель на новый экземпляр структуры Agent.
func NewAgent(errChan chan error) *Agent {
	addr := fmt.Sprintf("%s:%d",
		config.Cfg.Services.Orchestrator.ORCHESTRATOR_ADDR,
		config.Cfg.Services.Orchestrator.ORCHESTRATOR_GRPC_PORT)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Fatalf("could not connect to grpc server: ", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	computingPower := config.Cfg.Services.Agent.COMPUTING_POWER
	workersGroup := make([]*workers.Worker, computingPower)

	a := &Agent{
		addr:        addr,
		errChan:     errChan, // Канал для ошибок
		workersCtx:  ctx,     // Контекст для воркеров
		stopWorkers: cancel,  // Функция для отмены контекста
		connection:  conn,
		workers:     workersGroup,      // Слайс воркеров
		power:       computingPower,    // Вычислительная мощность
		wg:          &sync.WaitGroup{}, // WaitGroup для ожидания завершения воркеров
	}
	return a
}

// initWorkers инициализирует воркеров, создавая новые экземпляры Worker
// и добавляя их в слайс workers.
func (a *Agent) initWorkers(client pb.OrchestratorServiceClient) {
	logger.Log.Debugf("Инициализация %d работников", a.power)
	for i := range a.power {
		a.workers[i] = workers.NewWorker(i+1, client, a.wg, a.errChan)
	}
}

// Start запускает воркеров, запуская для каждого из них отдельную горутину.
func (a *Agent) Start() {
	conn, err := grpc.NewClient(a.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Fatalf("Не удалось подключитсья к gRPC сервису: %v", err)
	}
	grpcClient := pb.NewOrchestratorServiceClient(conn)
	a.initWorkers(grpcClient)
	logger.Log.Debugf("Запуск %d работников", a.power)
	for i := 1; i <= a.power; i++ {
		go a.workers[i-1].Start(a.workersCtx)
	}
}

// Stop останавливает воркеров, отменяя контекст и дожидаясь завершения
// всех горутин воркеров.
func (a *Agent) Stop() {
	a.stopWorkers() //  cancel
	a.connection.Close()
	if a.workers != nil {
		a.wg.Wait()
	}
}

// Запуск сервиса агента
func main() {
	// Инициализация конфига и логгера
	initializer.Init()

	errChan := make(chan error, 1)

	// Запуск сервиса агента в отдельной горутине чтобы можно было поймать завершение
	agentService := NewAgent(errChan)
	go func() {
		logger.Log.Debugf("Запуск сервиса Агент...")
		agentService.Start()
		logger.Log.Infof("Сервис Агент запущен")
	}()

	shutdown.WaitForShutdown(errChan, "Агент", agentService)
}
