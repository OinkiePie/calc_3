package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/OinkiePie/calc_3/pkg/models"
)

// APIClient - структура для взаимодействия с API оркестратора.
type APIClient struct {
	// Адрес с которого получают и на который отравляют задачи.
	url string
	// HTTP клиент для выполнения запросов.
	httpClient *http.Client
}

// NewAPIClient - создает новый экземпляр APIClient.
//
// Args:
//
//	URL: string - Адресс для осуществления запросов.
//	authToken: string - Токен авторизации для доступа к API оркестратора.
//	httpClient: *http.Client - HTTP клиент для выполнения запросов.
//
//	Returns:
//	*APIClient - Новый экземпляр APIClient.
func NewAPIClient(url string, httpClient *http.Client) *APIClient {
	return &APIClient{
		url:        url,
		httpClient: httpClient,
	}
}

// GetTask - получает задачу от оркестратора.
//
// GetTask отправляет GET запрос на URL оркестратора, добавляя заголовок Authorization.
// В случае успеха, десериализует JSON-ответ в структуру models.TaskResponse.
//
// Returns:
//
//	*models.TaskResponse - Указатель на структуру TaskResponse, содержащую информацию о задаче,
//	 или nil, если задач нет.
func (c *APIClient) GetTask() (*models.TaskResponse, error) {
	req, err := http.NewRequest("GET", c.url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return nil, fmt.Errorf("неожиданный статус: %d", resp.StatusCode)
	}

	var task models.TaskResponse
	err = json.NewDecoder(resp.Body).Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// CompleteTask - отправляет результат выполненной задачи оркестратору.
//
//	CompleteTask отправляет POST запрос на URL оркестратора c JSON-представлением
//	структуры models.TaskCompleted, добавляя заголовок
//	Content-Type: application/json и заголовок Authorization.
//
// Args:
//
//	completedTask : models.TaskCompleted - Структура TaskCompleted, содержащая информацию о выполненной задаче и ее результате.
//
// Returns:
//
//	error - Ошибка, возникшая во время выполнения запроса или сериализации тела запроса.
func (c *APIClient) CompleteTask(completedTask models.TaskCompleted) error {

	body, err := json.Marshal(completedTask)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("неожиданный статус: %d", resp.StatusCode)
	}

	return nil
}
