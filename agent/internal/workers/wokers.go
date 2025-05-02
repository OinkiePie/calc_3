package workers

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/OinkiePie/calc_3/pkg/operators"
	pb "github.com/OinkiePie/calc_3/pkg/proto"
)

var (
	errDivisionByZero = errors.New("деление на ноль")
	errFirstNil       = errors.New("первый оператор не может быть nil")
)

// Worker представляет собой рабочего, выполняющего задачи.
type Worker struct {
	errChan  chan error                   // Канал для отправки ошибок, возникающих при выполнении задач.
	workerID int                          // Уникальный идентификатор рабочего.
	client   pb.OrchestratorServiceClient //  Клиент gRPC для получения и отправки задач.
	wg       *sync.WaitGroup              // WaitGroup для сигнализации о завершении работы.
}

// NewWorker создает нового воркера.
//
// Args:
//
//	workerID: int - Уникальный идентификатор воркера.
//	client:   pb.OrchestratorServiceClient - Клиент gRPC для взаимодействия с оркестратором.
//	wg:       *sync.WaitGroup - Указатель на WaitGroup для синхронизации работы воркеров.
//	errChan:  chan error - Канал для передачи ошибок из воркеров.
//
// Returns:
//
//	*Worker - Указатель на созданный экземпляр воркера.
func NewWorker(workerID int, client pb.OrchestratorServiceClient, wg *sync.WaitGroup, errChan chan error) *Worker {
	return &Worker{
		workerID: workerID,
		client:   client,
		wg:       wg,
		errChan:  errChan,
	}
}

// Start запускает воркер и начинает получать и обрабатывать задачи.
//
// Args:
//
//	ctx context.Context - Контекст для управления жизненным циклом воркера.  При отмене контекста,
//	воркер завершает свою работу
//
// Воркер постоянно запрашивает задачи у сервиса, выполняет их и отправляет результаты.
// Функция завершается, когда контекст ctx отменяется.
func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	defer w.wg.Done()

	prevErr := errors.New("") //  Сохраняем предыдущую ошибку (для логирования)
	waiting := true           //  Флаг, указывающий, что воркер ждет задачи.

	for {
		select {
		case <-ctx.Done():
			// Контекст отменен - завершаем работу воркера.
			logger.Log.Debugf("Рабочий %d отключен", w.workerID)
			return
		default:
			//  Основной цикл обработки задач.
			resp, err := w.client.GetTask(context.TODO(), &pb.Empty{}) // Запрашиваем задачу

			if err != nil {
				// Обработка ошибок при получении задачи:
				if prevErr.Error() != err.Error() {
					// Логируем только новые ошибки (чтобы не засорять логи)
					logger.Log.Errorf("Рабочий %d: Ошибка при получении задачи: %v. Повторные запросы каждые %d мс.",
						w.workerID, err, config.Cfg.Services.Agent.AGENT_REPEAT_ERR)
				}
				prevErr = err // Сохраняем текущую ошибку для сравнения со следующей
				time.Sleep(time.Duration(config.Cfg.Services.Agent.AGENT_REPEAT_ERR) * time.Millisecond)
				continue // Переходим к следующей итерации цикла (повторный запрос)
			}

			if resp.GetId() == 0 {
				// Нет доступных задач:
				if waiting {
					// Логируем только один раз, когда воркер переходит в состояние ожидания
					logger.Log.Debugf("Рабочий %d: Нет доступных задач. Повторные запросы каждые %d мс.",
						w.workerID, config.Cfg.Services.Agent.AGENT_REPEAT)
					waiting = false //  Устанавливаем флаг, что мы уже логировали состояние ожидания
				}
				time.Sleep(time.Duration(config.Cfg.Services.Agent.AGENT_REPEAT) * time.Millisecond)
				continue // Переходим к следующей итерации цикла (повторный запрос)
			}
			// Получена задача:
			task := &models.TaskResponse{
				ID:         resp.GetId(),
				Args:       convertArgs(resp.GetArgs()),
				Operation:  resp.GetOperation(),
				Expression: resp.GetExpression(),
				Error:      resp.GetError(),
			}
			logger.Log.Debugf("Рабочий %d: Получена задача %d", w.workerID, task.ID)
			waiting = true //  Устанавливаем флаг, что воркер снова готов к выполнению задач

			//  Устанавливаем таймаут на выполнение задачи
			taskCtx, cancel := context.WithTimeout(context.Background(), calcOperationTime(task.Operation))
			defer cancel()

			// Запускаем вычисление в горутине
			resultChan := make(chan float64, 1) // Канал для результата
			errorChan := make(chan error, 1)    // Канал для ошибок

			go func(t *models.TaskResponse) {
				//  Обеспечиваем, что если возникла паника, ее можно было перехватить (recover)
				defer func() {
					if r := recover(); r != nil {
						close(resultChan) // Закрываем каналы, если произошла паника
						close(errorChan)
						w.errChan <- fmt.Errorf("ошибка во время вычисления: %v", r) // Отправляем ошибку в канал ошибок
					}
				}()

				result, err := Calculate(t) // Вычисляем задачу
				if err != nil {
					errorChan <- err // Отправляем ошибку в канал ошибок
					return
				}
				resultChan <- result // Отправляем результат в канал
			}(task)

			var result float64 // Переменная для хранения результата
			select {
			case result = <-resultChan:
				// Успешное завершение вычисления
				<-taskCtx.Done() //  Ждем, пока истечет таймаут (если задача выполнилась слишком быстро)
				logger.Log.Debugf("Рабочий %d: Задача %d успешно выполнена", w.workerID, task.ID)
			case err = <-errorChan:
				//  Ошибка при вычислении
				logger.Log.Debugf("Рабочий %d: Задача %d невыполнима: %v", w.workerID, task.ID, err)
				task.Error = err.Error() // Устанавливаем сообщение об ошибке
				result = 0               // Устанавливаем результат в 0 при ошибке
			}

			//  Проверка на значения +Inf и -Inf
			if math.IsInf(result, 1) {
				result = 0
				task.Error = "Результат - +Inf"
			}

			if math.IsInf(result, -1) {
				result = 0
				task.Error = "Результат - -Inf"
			}

			// Формируем сообщение с результатом для отправки
			completedTask := &pb.TaskCompleted{
				Expression: task.Expression,
				Id:         task.ID,
				Result:     result,
				Error:      task.Error,
			}

			//  Отправляем результат в оркестратор
			_, err = w.client.SubmitResult(context.TODO(), completedTask)
			if err != nil {
				// Обработка ошибок при отправке результата
				logger.Log.Errorf("Рабочий %d: Ошибка при отправлении задачи %d: %v", w.workerID, task.ID, err)
				time.Sleep(time.Duration(config.Cfg.Services.Agent.AGENT_REPEAT_ERR) * time.Second) //  Задержка при ошибке отправки
			} else {
				logger.Log.Debugf("Рабочий %d: Задача %d успешно отправлена", w.workerID, task.ID)
			}
		}
	}
}

// Calculate выполняет математическую операцию над двумя аргументами, указанными в задаче.
// Если второй аргумент равен nil, выполняется унарный минус для первого аргумента.
//
// Args:
//
//	task: (*models.TaskResponse) - Задача, содержащая аргументы и операцию.
//
// Returns:
//
//	float64 - Результат выполнения операции.
//	error - Ошибка, если операция не может быть выполнена (например, деление на ноль или неизвестная операция).
func Calculate(task *models.TaskResponse) (float64, error) {
	var arg1, arg2 float64

	// Nil попадает в операнд только если оператором является унарный минус.
	// При этом число которое необходимо обратить всегда первый операнд.
	if task.Args[0] == nil {
		// Первый оператор никогда не может быть nil
		return 0, errFirstNil
	}
	arg1 = *task.Args[0]
	// Если 2-й операнд - nil, то операнд всегда унарный минус
	if task.Args[1] == nil {
		return -*task.Args[0], nil
	}
	arg2 = *task.Args[1]

	switch task.Operation {

	case operators.OpAdd:
		return arg1 + arg2, nil

	case operators.OpSubtract:
		return arg1 - arg2, nil

	case operators.OpMultiply:
		return arg1 * arg2, nil

	case operators.OpDivide:
		if arg2 == 0 {
			return 0, errDivisionByZero
		}
		return arg1 / arg2, nil

	case operators.OpPower:
		return math.Pow(arg1, arg2), nil
	}

	return 0, fmt.Errorf("неизвестный оператор: %s", task.Operation)
}

// calcOperationTime возвращает длительность выполнения для указанной математической операции.
// Время выполнения берется из конфигурации приложения.
//
// Args:
//
//	operation: string - Строковый идентификатор операции
//
// Returns:
//
//	time.Duration - Длительность выполнения операции в миллисекундах
func calcOperationTime(operation string) time.Duration {
	var timeMs int

	switch operation {
	case operators.OpAdd:
		timeMs = config.Cfg.Math.TIME_ADDITION_MS
		break
	case operators.OpSubtract:
		timeMs = config.Cfg.Math.TIME_SUBTRACTION_MS
		break
	case operators.OpMultiply:
		timeMs = config.Cfg.Math.TIME_MULTIPLICATION_MS
		break
	case operators.OpDivide:
		timeMs = config.Cfg.Math.TIME_DIVISION_MS
		break
	case operators.OpPower:
		timeMs = config.Cfg.Math.TIME_POWER_MS
		break
	case operators.OpUnaryMinus:
		timeMs = config.Cfg.Math.TIME_UNARY_MINUS_MS
		break
	default:
		logger.Log.Warnf("Оператор %s не найден", operation)
		return 0
	}

	return time.Duration(timeMs) * time.Millisecond
}

// convertArgs преобразует срез указателей на  pb.WrappedDouble в срез указателей на float64.
//
// Args:
//
//	pbArgs: []*pb.WrappedDouble - Срез указателей на WrappedDouble из proto-файла.
//
// Returns:
//
//	[]*float64 - Срез указателей на float64, соответствующий переданному pbArgs.
func convertArgs(pbArgs []*pb.WrappedDouble) []*float64 {
	goArgs := make([]*float64, len(pbArgs))
	for i, arg := range pbArgs {
		if arg != nil {
			goArgs[i] = arg.Value
		}
	}
	return goArgs
}
