package config_test

//
//import (
//	"fmt"
//	"io"
//	"log"
//	"os"
//	"path/filepath"
//	"testing"
//
//	"github.com/OinkiePie/calc_3/config"
//	"github.com/OinkiePie/calc_3/pkg/logger"
//	"github.com/stretchr/testify/assert"
//)
//
//func init() {
//	// Отключаем выводы и инициализируем конфиг
//	log.SetOutput(io.Discard)
//	config.InitConfig()
//	logger.InitLogger(logger.Options{Level: 6})
//}
//
//// Создаем фальшивые директории и удаляем только если они не существовали ранее
//// чтобы случайно не уничтожить важные файлы.
//// В реальном случае запуска отсутствие директории с конфигом по умолчанию
//// не вызовет проблем т.к. инициализатор лоавит и обрабатывает ошибки.
//func InFakeDevDir(t assert.TestingT, function func() error) error {
//
//	_, err := os.Stat("config")
//	if os.IsNotExist(err) {
//		err := CreateDir("config")
//		defer RemoveDir("config")
//		if err != nil {
//			return err
//		}
//	}
//
//	_, err = os.Stat("config/configs")
//	if os.IsNotExist(err) {
//		err := CreateDir("config/configs")
//		defer RemoveDir("config/configs")
//		if err != nil {
//			return err
//		}
//	}
//
//	_, err = os.Stat("config/configs/dev.yml")
//	if os.IsNotExist(err) {
//		err = CreateAndWrite("config/configs/dev.yml", `empty: "value"`)
//		defer RemoveDir("config/configs/dev.yml")
//		if err != nil {
//			return err
//		}
//	}
//
//	err = function()
//	return err
//}
//
//func CreateDir(path string) error {
//	err := os.MkdirAll(path, 0755)
//	if err != nil {
//		return fmt.Errorf("Ошибка при создании директории: %v", err)
//	}
//
//	return nil
//}
//
//func RemoveDir(path string) error {
//	err := os.RemoveAll(path)
//	if err != nil {
//		return fmt.Errorf("Ошибка при удалении директории: %v", err)
//	}
//	return nil
//}
//
//func CreateAndWrite(filename, value string) error {
//	// Создаем файл
//	file, err := os.Create(filename)
//	if err != nil {
//		return fmt.Errorf("Ошибка при создании файла: %v", err)
//	}
//	defer file.Close()
//
//	// Записываем что-то в файл
//	_, err = file.WriteString(value)
//	if err != nil {
//		return fmt.Errorf("Ошибка при записи в файл: %v", err)
//	}
//
//	return nil
//}
//
//func Remove(filename string) error {
//	// Удаляем файл
//	err := os.Remove(filename)
//	if err != nil {
//		return fmt.Errorf("Ошибка при удалении файла: %v", err)
//	}
//
//	return nil
//}
//
//func TestConfig_Yml(t *testing.T) {
//	tempDir := t.TempDir()
//	path := filepath.Join(tempDir, "good.yml")
//
//	err := os.Setenv("APP_CFG", path)
//	assert.NoError(t, err)
//
//	yamlContent := `
//server:
//  orchestrator:
//    ADDR_ORCHESTRATOR: "newadres"
//math:
//  TIME_ADDITION_MS: 50
//  TIME_SUBTRACTION_MS: 100
//logger:
//  level: 1
//`
//
//	err = CreateAndWrite(path, yamlContent)
//	assert.NoError(t, err)
//
//	err = config.InitConfig()
//	assert.NoError(t, err)
//
//	assert.Equal(t, "newadres", config.Cfg.Server.Orchestrator.ADDR_ORCHESTRATOR)
//	assert.Equal(t, 50, config.Cfg.Math.TIME_ADDITION_MS)
//	assert.Equal(t, 100, config.Cfg.Math.TIME_SUBTRACTION_MS)
//	assert.Equal(t, 1, config.Cfg.Logger.Level)
//}
//
//func TestConfig_Undefined(t *testing.T) {
//	err := os.Setenv("APP_CFG", "thisconfigfileisnotexists.yml")
//	assert.NoError(t, err)
//
//	err = config.InitConfig()
//	assert.Error(t, err)
//}
//
//func TestConfig_EmptyYml(t *testing.T) {
//	tempDir := t.TempDir()
//	path := filepath.Join(tempDir, "empty.yml")
//
//	err := os.Setenv("APP_CFG", path)
//	assert.NoError(t, err)
//
//	err = CreateAndWrite(path, "")
//	assert.NoError(t, err)
//
//	err = config.InitConfig()
//	assert.Error(t, err)
//}
//
//func TestConfig_BadYml(t *testing.T) {
//	tempDir := t.TempDir()
//	path := filepath.Join(tempDir, "bad.yml")
//
//	err := os.Setenv("APP_CFG", path)
//	assert.NoError(t, err)
//
//	err = CreateAndWrite(path, "aboba")
//	assert.NoError(t, err)
//
//	err = config.InitConfig()
//	assert.Error(t, err)
//}
//
//func TestConfig_Env(t *testing.T) {
//	err := os.Setenv("APP_CFG", "")
//	assert.NoError(t, err)
//	err = os.Setenv("ADDR_ORCHESTRATOR", "newadres")
//	assert.NoError(t, err)
//	err = os.Setenv("TIME_ADDITION_MS", "100")
//	assert.NoError(t, err)
//
//	// Отключаем конфиг
//	err = os.Setenv("APP_CFG", "CFG_FALSE")
//	assert.NoError(t, err)
//
//	err = config.InitConfig()
//	assert.NoError(t, err)
//
//	assert.Equal(t, "newadres", config.Cfg.Server.Orchestrator.ADDR_ORCHESTRATOR)
//	assert.Equal(t, 100, config.Cfg.Math.TIME_ADDITION_MS)
//}
//
//func TestConfig_NoEnvCfg(t *testing.T) {
//	err := os.Setenv("APP_CFG", "")
//	assert.NoError(t, err)
//
//	/// Проверяем загрузку конфига по умолчанию (config/configs/dev.yml)
//	err = InFakeDevDir(t, config.InitConfig)
//	assert.NoError(t, err)
//}
