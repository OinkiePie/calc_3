package initializer

import (
	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/pkg/logger"
)

// Init инициализирует конфигурацию приложения и настраивает логгер.
//
// Функция загружает конфигурацию из файла (если он указан) и инициализирует логгер
// с параметрами, указанными в конфигурации. В случае ошибки при загрузке
// конфигурации используется конфигурация по умолчанию (dev) и выводится предупреждение.
func Init() {
	errConfig := config.InitConfig()

	logger.InitLogger(logger.Options{
		Level:        logger.Level(config.Cfg.Logger.Level),
		TimeFormat:   config.Cfg.Logger.TimeFormat,
		CallDepth:    config.Cfg.Logger.CallDepth,
		DisableCall:  config.Cfg.Logger.DisableCall,
		DisableTime:  config.Cfg.Logger.DisableTime,
		DisableColor: config.Cfg.Logger.DisableColor,
	})

	if errConfig != nil {
		logger.Log.Errorf(errConfig.Error())
		logger.Log.Warnf("Ошибка при загрузке конфигурации, используется конфигурация dev, если она существует")
	} else {
		logger.Log.Infof("Загружена конфигурация: %s", config.Filename)
	}
}
