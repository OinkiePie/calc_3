package task_splitter

import (
	"errors"
	"strconv"
	"strings"
	"unicode"

	"github.com/OinkiePie/calc_3/pkg/models"
	"github.com/OinkiePie/calc_3/pkg/operators"
)

var (
	errOneOperand        = errors.New("минимум два операнда требуются для расчета")
	errUnopenedParen     = errors.New("неоткрытая скобка")
	errUnclosedParen     = errors.New("незакрытая скобка")
	errInvalidSyntax     = errors.New("неверный синтаксис")
	errNotEnoughOperands = errors.New("недостаточно операндов")
	errUnaryMinus        = errors.New("недостаточно операндов для унарного минуса")
	errRPN               = errors.New("не удалось преобразовать RPN")
)

// ParseExpression разбирает математическое выражение и преобразует его в набор вычислительных задач.
//
// Args:
//
//	expression: string - Математическое выражение в инфиксной нотации
//
// Returns:
//
//	[]*models.Task - Список задач для вычисления выражения
//	error - Ошибка парсинга:
//	    - ошибки из infixToRPN при невалидном выражении
//	    - ошибки из rpnToTasks при создании задач
func ParseExpression(expression string) ([]*models.Task, error) {

	expression = strings.ReplaceAll(expression, " ", "")

	rpn, err := infixToRPN(expression)
	if err != nil {
		return nil, err
	}

	var tasks []*models.Task
	tasks, err = rpnToTasks(rpn)

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// precedence определяет приоритет оператора для правильной вложенности при разбиении на задачи.
//
// Args:
//
//	op: string - Строка, представляющая оператор (+, -, *, /, ^, u-).
//
// Returns:
//
//	int - Целое число, представляющее приоритет оператора. Чем больше число, тем выше приоритет.
//	     Возвращает 0 для неопознанных операторов.
func precedence(op string) int {
	switch op {
	case operators.OpAdd, operators.OpSubtract:
		return 1
	case operators.OpMultiply, operators.OpDivide:
		return 2
	case operators.OpUnaryMinus:
		return 3
	case operators.OpPower:
		return 4
	default:
		return 0
	}
}

// isOperator проверяет, является ли токен строкой, представляющей математический оператор.
//
// Args:
//
//	token: string - Строка, которую необходимо проверить.
//
// Returns:
//
//	bool - true, если токен является одним из допустимых операторов (+, -, *, /, ^, u-), иначе false.
func isOperator(token string) bool {
	switch token {
	case operators.OpAdd, operators.OpSubtract, operators.OpMultiply, operators.OpDivide, operators.OpPower:
		return true
	default:
		return false
	}
}

// isUnaryMinus определяет, следует ли обрабатывать знак минус как унарный (например, "-5") или бинарный (например, "3 - 5").
//
// Args:
//
//	tokens: []string - Срез строк, представляющий токены выражения.
//	i: int - Индекс текущего токена в срезе.
//
// Returns:
//
//	bool - true, если минус должен быть обработан как унарный, иначе false.
func isUnaryMinus(tokens []string, i int) bool {
	if i == 0 {
		return true // Минус в начале выражения - унарный
	}
	prevToken := tokens[i-1]
	return prevToken == operators.ParenLeft || isOperator(prevToken)
}

// infixToRPN преобразует математическое выражение в инфиксной нотации (обычная запись) в обратную польскую нотацию (RPN).
// RPN упрощает вычисление выражений с помощью стека.
//
// Args:
//
//	expression: string - Математическое выражение в инфиксной нотации.
//
// Returns:
//
//	[]string - Срез строк, представляющий выражение в обратной польской нотации (RPN).
//	error - Ошибка, если выражение не может быть преобразовано.
func infixToRPN(expression string) ([]string, error) {

	tokens := tokenize(expression) // Сначала разбиваем на токены
	var output []string            // Выходная очередь
	var stack []string             // Стек операторов

	for i, token := range tokens {
		switch {
		case isNumber(token): // Если число, добавляем в выходную очередь
			output = append(output, token)
		case token == operators.ParenLeft: // Если открывающая скобка, помещаем в стек
			stack = append(stack, token)
		case token == operators.ParenRight: // Если закрывающая скобка
			for len(stack) > 0 && stack[len(stack)-1] != operators.ParenLeft {
				// Переносим операторы из стека в выходную очередь,
				// пока он не опустеет, или мы не встретим открывающую скобку
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				// Если в стеке не осталось открывающей скобки, это означает, что у нас была
				// закрывающая скобка, но не было соответствующей открывающей скобки в выражении
				return nil, errUnopenedParen
			}
			stack = stack[:len(stack)-1] // Удаляем открывающую скобку из стека
		case isOperator(token): // Если оператор
			if token == "-" && isUnaryMinus(tokens, i) {
				token = operators.OpUnaryMinus // Помечаем как унарный минус
			}
			for len(stack) > 0 && precedence(token) <= precedence(stack[len(stack)-1]) {
				// Переносим операторы из стека в выходную очередь, пока приоритет текущего оператора
				// меньше или равен приоритету оператора на вершине стека
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token) // Помещаем текущий оператор в стек
		default:
			return nil, errInvalidSyntax
		}
	}

	// Переносим все оставшиеся операторы из стека в выходную очередь
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		if top == operators.ParenLeft || top == operators.ParenRight {
			return nil, errUnclosedParen
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	if len(output) == 1 {
		return nil, errOneOperand
	}

	return output, nil
}

// tokenize разбивает входную строку математического выражения на отдельные токены (числа, операторы, скобки).
// Токены используются для дальнейшей обработки выражения.
//
// Args:
//
//	expression: string - Строка, содержащая математическое выражение.
//
// Returns:
//
//	[]string - Срез строк, представляющий токены выражения.
func tokenize(expression string) []string {
	var tokens []string
	var currentNumber string

	for i, r := range expression {
		s := string(r)
		// Если символ является цифрой, точкой или знаком "+" в начале числа
		if unicode.IsDigit(r) || s == operators.Point || (s == "+" && (i == 0 || isOperator(string(expression[i-1])) || expression[i-1] == '(')) {
			currentNumber += s
		} else {
			// Если накопилось число, добавляем его в токены
			if currentNumber != "" {
				tokens = append(tokens, currentNumber)
				currentNumber = ""
			}
			// Добавляем текущий символ (оператор или скобку) в токены
			if s != "" {
				tokens = append(tokens, s)
			}
		}
	}

	// Если осталось число, добавляем его в токены
	if currentNumber != "" {
		tokens = append(tokens, currentNumber)
	}

	return tokens
}

// isNumber проверяет, является ли переданный токен числом.
//
// Args:
//
//	token: string - Строка, которую необходимо проверить.
//
// Returns:
//
//	bool - true, если токен может быть преобразован в число, иначе false.
func isNumber(token string) bool {
	_, err := strconv.ParseFloat(token, 64)
	return err == nil
}

// rpnToTasks преобразует выражение в обратной польской записи (RPN) в список задач для вычисления.// Использует стековый алгоритм для построения графа зависимостей между операциями.
//
// Args:
//
//	rpn: []string - Выражение в формате RPN (массив токенов)
//
// Returns:
//
//	[]*models.Task - Список задач, представляющих операции выражения
//	error - Ошибка преобразования:
//	    - errNotEnoughOperands: недостаточно операндов для операции
//	    - errUnaryMinus: отсутствует операнд для унарного минуса
//	    - errRPN: неверный формат RPN или числового значения
//
// Функция использует стек для отслеживания операндов и операций.
// При обнаружении оператора, функция извлекает два операнда (или один для унарного минуса) из стека,
// создает новую задачу с этим оператором и зависимостями, и помещает задачу в стек.
// Числа преобразуются в задачи с предопределенным статусом "completed" и результатом.
// Функция возвращает nil и ошибку, если входная строка RPN некорректна.
func rpnToTasks(rpn []string) ([]*models.Task, error) {
	var tasks []*models.Task // Срез для хранения созданных задач
	var stack []*models.Task // Стек для хранения операндов и промежуточных результатов
	var indexer int          // Счетчик для индексации зависимостей

	// Вспомогательная функция для создания новой задачи
	newTask := func(operator string) models.Task {
		return models.Task{
			Operation:         operator,            // Операция
			Args:              make([]*float64, 2), //  Аргументы
			Dependencies:      []int64{-1, -1},     //  ID зависимостей
			DependencyIndexes: make([]int, 2),      //  Индексы зависимостей в срезе tasks
		}
	}

	//  Цикл по токенам RPN
	for _, token := range rpn {
		switch token {
		case operators.OpAdd, operators.OpSubtract, operators.OpMultiply, operators.OpDivide, operators.OpPower:
			//  Обработка бинарных операторов (+, -, *, /)
			if len(stack) < 2 {
				//  Недостаточно операндов на стеке
				return nil, errNotEnoughOperands
			}

			//  Извлекаем операнды из стека
			operand2 := stack[len(stack)-1]
			operand1 := stack[len(stack)-2]
			stack = stack[:len(stack)-2] //  Удаляем операнды из стека

			task := newTask(token) // Создаем новую задачу для оператора

			//  Заполняем аргументы задачи (либо значениями, либо индексами зависимостей)
			if operand1.Result != nil {
				val := *operand1.Result
				task.Args[0] = &val // Используем значение
			} else {
				task.DependencyIndexes[0] = int(operand1.ID) //  Устанавливаем индекс зависимости
			}

			if operand2.Result != nil {
				val := *operand2.Result
				task.Args[1] = &val //  Используем значение
			} else {
				task.DependencyIndexes[1] = int(operand2.ID) //  Устанавливаем индекс зависимости
			}

			indexer++ //  Увеличиваем счетчик
			task.ID = int64(indexer)
			tasks = append(tasks, &task)               //  Добавляем задачу в срез
			stack = append(stack, tasks[len(tasks)-1]) //  Помещаем задачу в стек

		case operators.OpUnaryMinus:
			//  Обработка унарного минуса
			if len(stack) < 1 {
				//  Недостаточно операндов на стеке
				return nil, errUnaryMinus
			}

			operand := stack[len(stack)-1]
			stack = stack[:len(stack)-1] //  Удаляем операнд из стека

			task := newTask(operators.OpUnaryMinus) //  Создаем новую задачу для оператора

			//  Заполняем аргументы задачи (либо значениями, либо индексами зависимостей)
			if operand.Result != nil {
				val := *operand.Result
				task.Args[0] = &val //  Используем значение
			} else {
				task.DependencyIndexes[0] = int(operand.ID) // Устанавливаем индекс зависимости
			}

			indexer++ //  Увеличиваем счетчик
			task.ID = int64(indexer)
			tasks = append(tasks, &task)               //  Добавляем задачу в срез
			stack = append(stack, tasks[len(tasks)-1]) //  Помещаем задачу в стек

		default:
			// Обработка чисел (операндов)
			num, err := strconv.ParseFloat(token, 64)
			if err != nil {
				//  Ошибка при преобразовании токена в число
				return nil, errRPN
			}

			//  Создаем задачу для числа со статусом "completed"
			stack = append(stack, &models.Task{
				Status: "completed",
				Result: &num,
			})
		}
	}

	//  Проверка, что в стеке остался только один элемент (корень выражения)
	if len(stack) != 1 {
		return nil, errRPN
	}

	return tasks, nil // Возвращаем созданные задачи
}
