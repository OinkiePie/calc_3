package task_splitter_test

import (
	"github.com/OinkiePie/calc_3/config"
	"github.com/OinkiePie/calc_3/orchestrator/internal/task_splitter"
	"github.com/OinkiePie/calc_3/pkg/logger"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"testing"
)

func init() {
	// Отключаем выводы и инициализируем конфиг
	log.SetOutput(io.Discard)
	_ = config.InitConfig()
	logger.InitLogger(logger.Options{Level: 6})
}

func TestParseExpression(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectedLen int
		expectError bool
		err         string
	}{
		{
			name:        "Valid expression: simple addition",
			expression:  "2 + 2",
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "Valid expression: simple subtraction",
			expression:  "5 - 3",
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "Valid expression: simple multiplication",
			expression:  "2 * 3",
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "Valid expression: simple division",
			expression:  "6 / 2",
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "Valid expression: power operation",
			expression:  "2 ^ 3",
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "Valid expression: unary minus",
			expression:  "-5 + 3",
			expectedLen: 2,
			expectError: false,
		},
		{
			name:        "Valid expression: complex expression with parentheses",
			expression:  "2 + 3 * (4 - 1)",
			expectedLen: 3,
			expectError: false,
		},
		{
			name:        "Valid expression: nested parentheses",
			expression:  "2 * (3 + (4 - 1))",
			expectedLen: 3,
			expectError: false,
		},
		{
			name:        "Valid expression: multiple operations",
			expression:  "2 + 3 * 4 - 1 / 2",
			expectedLen: 4,
			expectError: false,
		},
		{
			name:        "Invalid expression: missing operand",
			expression:  "2 +",
			expectedLen: 0,
			expectError: true,
			err:         "недостаточно операндов",
		},
		{
			name:        "Invalid expression: unclosed parenthesis",
			expression:  "2 + (3 * 4",
			expectedLen: 0,
			expectError: true,
			err:         "незакрытая скобка",
		},
		{
			name:        "Invalid expression: unopened parenthesis",
			expression:  "2 + 3) * 4",
			expectedLen: 0,
			expectError: true,
			err:         "неоткрытая скобка",
		},
		{
			name:        "Invalid expression: invalid syntax",
			expression:  "2 + * 3",
			expectedLen: 0,
			expectError: true,
			err:         "неверный синтаксис",
		},
		{
			name:        "Invalid expression: single operand",
			expression:  "2",
			expectedLen: 0,
			expectError: true,
			err:         "минимум два операнда требуются для расчета",
		},
		{
			name:        "Invalid expression: empty expression",
			expression:  "",
			expectedLen: 0,
			expectError: true,
			err:         "неверный синтаксис",
		},
		{
			name:        "Invalid expression: unary minus without operand",
			expression:  "-",
			expectedLen: 0,
			expectError: true,
			err:         "недостаточно операндов для унарного минуса",
		},
		{
			name:        "Invalid expression: division by zero",
			expression:  "2 / 0",
			expectedLen: 1,
			expectError: false, // Парсинг успешен, но выполнение задачи вызовет ошибку
		},
		{
			name:        "Invalid expression: invalid characters",
			expression:  "2 + abc",
			expectedLen: 0,
			expectError: true,
			err:         "неверный синтаксис",
		},
		{
			name:        "Valid expression: decimal numbers",
			expression:  "2.5 + 3.7",
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "Valid expression: multiple unary minuses",
			expression:  "--5 + 3",
			expectedLen: 0,
			expectError: true,
			err:         "недостаточно операндов для унарного минуса",
		},
		{
			name:        "Valid expression: complex expression with multiple operations",
			expression:  "2 + 3 * 4 - 1 / 2 ^ 3",
			expectedLen: 5,
			expectError: false,
		},
		{
			name:        "Positive number: at the start",
			expression:  "+7 - 3",
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "Positive number: after operator",
			expression:  "2 + +7",
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "Positive number: after parenthesis",
			expression:  "(+7) - 3",
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "Positive number: mixed with negative numbers",
			expression:  "+7 - -3",
			expectedLen: 2,
			expectError: false,
		},
		{
			name:        "Positive number: Complex expression",
			expression:  "2 + (+7 * -3)",
			expectedLen: 3,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := task_splitter.ParseExpression(tt.expression)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedLen, len(tasks))
		})
	}
}
