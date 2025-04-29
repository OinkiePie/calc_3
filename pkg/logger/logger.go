package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level определяет уровень логирования.
type Level int

const (
	// DebugLevel логирует все сообщения.
	DebugLevel Level = iota
	// InfoLevel логирует информационные сообщения, предупреждения и ошибки.
	InfoLevel
	// WarningLevel логирует предупреждения и ошибки.
	WarningLevel
	// ErrorLevel логирует ошибки.
	ErrorLevel
	// FatalLevel логирует ошибки и затем вызывает os.Exit(1).
	FatalLevel
	// Disabled отключает все логирование.
	Disabled
)

var levelStrings = map[Level]string{
	DebugLevel:   "DEBUG",
	InfoLevel:    "INFO",
	WarningLevel: "WARN",
	ErrorLevel:   "ERROR",
	FatalLevel:   "FATAL",
}

const (
	levelPadding    = 6
	timePadding     = 0
	fileInfoPadding = 0
)

// String возвращает строковое представление Level.
func (l Level) String() string {
	if s, ok := levelStrings[l]; ok {
		return s
	}
	return "UNKNOWN"
}

// Logger представляет экземпляр логгера.
type Logger struct {
	mu           sync.Mutex  // Обеспечивает потокобезопасное логирование
	level        Level       // Текущий уровень логирования
	log          *log.Logger // Стандартный Go логгер
	timeFormat   string      // Формат для временных меток
	callDepth    int         // Глубина вызова для определения файла/строки источника
	disableCall  bool        // Отключить вывод источника
	disableTime  bool        // Отключает временные метки в логах
	disableColor bool        // Отключает цветной вывод
}

// Options представляет параметры конфигурации для логгера.
type Options struct {
	Level        Level  // Уровень логирования (по умолчанию: InfoLevel)
	TimeFormat   string // Формат временной метки (по умолчанию: "2006-01-02 15:04:05")
	CallDepth    int    // Глубина вызова для определения файла/строки источника (по умолчанию: 2)
	DisableCall  bool   // Отключить глубину вызова
	DisableTime  bool   // Отключить вывод времени
	DisableColor bool   // Отключить цветной вывод
}

// SetLevel устанавливает уровень логирования.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()         // Блокируем для обеспечения потокобезопасности
	defer l.mu.Unlock() // Не забываем разблокировать после завершения
	l.level = level     // Устанавливаем новый уровень
}

// getColorCode возвращает код цвета для заданного уровня
func (l *Logger) getColorCode(level Level) string {
	if l.disableColor {
		return ""
	}
	switch level {
	case DebugLevel:
		return "\033[36m" // Голубой
	case InfoLevel:
		return "\033[32m" // Зеленый
	case WarningLevel:
		return "\033[33m" // Желтый
	case ErrorLevel:
		return "\033[31m" // Красный
	case FatalLevel:
		return "\033[35m" // Пурпурный
	default:
		return ""
	}
}

// resetColorCode сбрасывает код цвета
func (l *Logger) resetColorCode() string {
	if l.disableColor {
		return ""
	}
	return "\033[0m"
}

// padRight добавляет пробелы справа от строки до заданной длины.
func padRight(s string, padding int) string {
	if len(s) >= padding {
		return s
	}
	return s + strings.Repeat(" ", padding-len(s))
}

// logf форматирует и печатает лог-сообщение с заданным уровнем.
func (l *Logger) logf(level Level, format string, v ...interface{}) {
	if format == "" {
		return
	}

	l.mu.Lock()         // Блокируем для обеспечения потокобезопасности
	defer l.mu.Unlock() // Не забываем разблокировать после завершения

	// Если уровень сообщения ниже текущего уровня логгера, ничего не делаем
	if level < l.level {
		return
	}

	colorCode := l.getColorCode(level) // Получаем код цвета для уровня
	resetColor := l.resetColorCode()   // Получаем код сброса цвета

	var timeString string
	if !l.disableTime {
		timeString = time.Now().Format(l.timeFormat)
		timeString = padRight(timeString, timePadding)        // Выравниваем время
		timeString = strings.TrimRight(timeString, " ") + " " // Удаляем лишние пробелы и добавляем один
	}

	levelString := padRight(level.String(), levelPadding) // Выравниваем уровень
	var fileInfo string
	if level <= WarningLevel {
		// Получаем информацию о файле и строке
		if !l.disableCall {
			_, file, line, ok := runtime.Caller(l.callDepth)
			if ok {
				file = filepath.Base(file)
				fileInfo = fmt.Sprintf("%s:%d", file, line)
				fileInfo = padRight(fileInfo, fileInfoPadding)    // Выравниваем информацию о файле
				fileInfo = strings.TrimRight(fileInfo, " ") + " " // Удаляем лишние пробелы и добавляем один
			} else {
				fileInfo = padRight("", fileInfoPadding) // Выравниваем пустую строку
			}
		}
	} else {
		fileInfo = padRight("", fileInfoPadding) // Выравниваем пустую строку
	}

	prefix := fmt.Sprintf("%s%s%s%s%s", colorCode, levelString, resetColor, timeString, fileInfo)
	message := fmt.Sprintf(format, v...)  // Форматируем сообщение
	l.log.Printf("%s%s", prefix, message) // Печатаем сообщение с префиксом
}

// Debugf логирует сообщение на уровне Debug.
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logf(DebugLevel, format, v...)
}

// Infof логирует сообщение на уровне Info.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.logf(InfoLevel, format, v...)
}

// Warnf логирует сообщение на уровне Warning.
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.logf(WarningLevel, format, v...)
}

// Errorf логирует сообщение на уровне Error.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logf(ErrorLevel, format, v...)
}

// Fatalf логирует сообщение на уровне Fatal и затем вызывает os.Exit(1).
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logf(FatalLevel, format, v...)
	os.Exit(1)
}

// SetTimeFormat устанавливает формат для временных меток.
func (l *Logger) SetTimeFormat(format string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timeFormat = format // Устанавливаем новый формат времени
}

// DisableTimestamp отключает отображение временных меток
func (l *Logger) DisableTimestamp(disable bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.disableTime = disable // Отключаем отображение временных меток
}

var (
	// Глобальная переменная для общего использования
	Log *Logger
)

// Инициализируем глоабльный логгер
func InitLogger(options Options) {
	if options.TimeFormat == "" {
		options.TimeFormat = "2006-01-02 15:04:05"
	}
	if options.CallDepth == 0 {
		options.CallDepth = 2
	}

	Log = &Logger{
		level:        options.Level,
		log:          log.New(os.Stdout, "", 0),
		timeFormat:   options.TimeFormat,
		callDepth:    options.CallDepth,
		disableCall:  options.DisableCall,
		disableTime:  options.DisableTime,
		disableColor: options.DisableColor,
	}
}
