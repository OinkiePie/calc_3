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

// Agent представляет собой gRPC агент, который взаимодействует с оркестратором для получения и обработки задач.
type Agent struct {
	addr        string             // Адрес оркестратора (host:port).
	errChan     chan error         // Канал для передачи ошибок от воркеров.
	stopWorkers context.CancelFunc // Функция для отмены контекста воркеров (остановка работы).
	connection  *grpc.ClientConn   // gRPC подключение к оркестратору.
	workersCtx  context.Context    // Контекст для управления воркерами.
	workers     []*workers.Worker  // Срез указателей на воркеры.
	power       int                // Количество воркеров (вычислительная мощность).
	wg          *sync.WaitGroup    // WaitGroup для ожидания завершения всех воркеров.
}

// NewAgent создает и инициализирует новый экземпляр агента.
//
// Args:
//
//	errChan:  chan error - Канал для передачи ошибок от воркеров.
//
// Returns:
//
//	*Agent - Указатель на созданный экземпляр агента.
//
//	Функция устанавливает соединение с gRPC-сервером оркестратора,
//	инициализирует контекст для воркеров, создает и настраивает агента.
//	При возникновении ошибки подключения к gRPC-серверу, функция завершает работу (fatal)
func NewAgent(errChan chan error) *Agent {
	// Формируем адрес оркестратора.
	addr := fmt.Sprintf("%s:%d",
		config.Cfg.Services.Orchestrator.ORCHESTRATOR_ADDR,
		config.Cfg.Services.Orchestrator.ORCHESTRATOR_GRPC_PORT)

	// Устанавливаем gRPC-соединение с оркестратором.
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		//  В случае ошибки подключения - фатальная ошибка.
		logger.Log.Fatalf("Не удалось подключиться к gRPC серверу: %v", err)
	}

	// Создаем контекст и функцию отмены для управления воркерами.
	ctx, cancel := context.WithCancel(context.Background())
	//  Получаем количество воркеров из конфигурации.
	computingPower := config.Cfg.Services.Agent.COMPUTING_POWER
	//  Инициализируем слайс воркеров.
	workersGroup := make([]*workers.Worker, computingPower)

	// Создаем и возвращаем новый агент.
	a := &Agent{
		addr:        addr,              //  Устанавливаем адрес оркестратора
		errChan:     errChan,           //  Устанавливаем канал для ошибок
		workersCtx:  ctx,               //  Устанавливаем контекст воркеров
		stopWorkers: cancel,            //  Устанавливаем функцию для отмены контекста
		connection:  conn,              //  Устанавливаем gRPC соединение
		workers:     workersGroup,      //  Устанавливаем слайс воркеров
		power:       computingPower,    //  Устанавливаем вычислительную мощность
		wg:          &sync.WaitGroup{}, //  Устанавливаем WaitGroup для синхронизации
	}
	return a
}

// initWorkers инициализирует воркеров агента.
//
// Args:
//
//	client: pb.OrchestratorServiceClient - gRPC клиент для взаимодействия с оркестратором.
//
// Функция создает и инициализирует воркеров агента, используя предоставленный gRPC клиент.
// Количество воркеров определяется полем power агента.
func (a *Agent) initWorkers(client pb.OrchestratorServiceClient) {
	logger.Log.Debugf("Инициализация %d работников", a.power)
	for i := range a.power {
		a.workers[i] = workers.NewWorker(i+1, client, a.wg, a.errChan)
	}
}

// Start запускает работу агента.
//
// Функция устанавливает gRPC соединение с оркестратором,
// инициализирует воркеров и запускает их в отдельных горутинах.
// Если соединение с gRPC сервисом не удалось, функция завершает работу (fatal).
func (a *Agent) Start() {
	grpcClient := pb.NewOrchestratorServiceClient(a.connection) //  Создаем gRPC клиент
	a.initWorkers(grpcClient)                                   //  Инициализируем воркеров
	logger.Log.Debugf("Запуск %d работников", a.power)

	//  Запускаем воркеров в горутинах
	for i := 1; i <= a.power; i++ {
		go a.workers[i-1].Start(a.workersCtx)
	}
}

// Stop останавливает работу агента.
//
//	Функция отменяет контекст воркеров, закрывает gRPC соединение и ожидает завершения всех воркеров.
func (a *Agent) Stop() {
	a.stopWorkers()       // Отменяем контекст воркеров (остановка)
	a.connection.Close()  // Закрываем gRPC соединение
	if a.workers != nil { // Проверяем, был ли слайс воркеров инициализирован
		a.wg.Wait() // Ожидаем завершения всех воркеров
	}
}

// main - запуск сервиса агента.
//
//	Функция инициализирует конфигурацию и логгер, создает экземпляр агента,
//	запускает агент в отдельной горутине и ожидает завершения работы.
//	В случае возникновения ошибки, ошибка передается в errChan и сервис отключается.
func main() {
	// Инициализируем конфигурацию и логгер.
	initializer.Init()

	errChan := make(chan error, 1) // Создаем канал для ошибок

	// Запускаем сервис агента в отдельной горутине, чтобы можно было поймать завершение.
	agentService := NewAgent(errChan)
	go func() {
		logger.Log.Debugf("Запуск сервиса Агент...")
		agentService.Start() //  Запускаем агент
		logger.Log.Infof("Сервис Агент запущен")
	}()

	shutdown.WaitForShutdown(errChan, "Агент", agentService) //  Ожидаем завершение работы
}
