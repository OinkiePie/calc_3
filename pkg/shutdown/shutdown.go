package shutdown

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/OinkiePie/calc_3/pkg/logger"
)

// Shutdownable определяет интерфейс для сервисов, поддерживающих корректное завершение работы.
type Shutdownable interface {
	Stop() // Stop выполняет остановку сервиса.
}

// WaitForShutdown ожидает сигнал завершения (os.Interrupt или syscall.SIGTERM)
// и затем корректно завершает работу указанного сервиса.
//
// Args:
//
//	errChan: chan error - Канал, из которого читаются ошибки, возникшие во время работы сервиса.
//	serviceName: string - Название сервиса (для логирования).
//	managers: Shutdownable - Интерфейс Shutdownable, предоставляющий метод Stop() для остановки сервиса.
func WaitForShutdown(errChan <-chan error, serviceName string, service Shutdownable) {
	// Создаем канал для обработки сигналов завершения (Ctrl+C, SIGTERM)
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)

	// Ожидаем завершения работы серверов или возникновения ошибки
	select {
	// Ожидаем получения ошибки из канала errChan.
	case err := <-errChan:
		logger.Log.Fatalf("Фатальная ошибка в %s: %v", serviceName, err)
	// Ожидаем получения сигнала из канала sigint (сигнал завершения).
	case <-sigint:
		logger.Log.Debugf("Получен сигнал завершения для %s, начинаем остановку...", serviceName)
	}

	// Корректное завершение работы сервиса.
	logger.Log.Infof("Остановка сервиса %s", serviceName)

	service.Stop()

	logger.Log.Infof("Сервис %s завершил работу", serviceName)
}
