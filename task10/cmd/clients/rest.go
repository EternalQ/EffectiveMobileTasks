package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type RestClient struct {
	baseURL string
	client  *http.Client
}

func NewRestClient(addr string) *RestClient {
	return &RestClient{
		baseURL: "http://" + addr,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *RestClient) Run() {
	for {
		fmt.Println("\n1. Создать")
		fmt.Println("2. Обновить")
		fmt.Println("3. Удалить")
		fmt.Println("4. Список")
		fmt.Println("0. Выход")

		choice := ReadInput("\nВыбор: ")

		switch choice {
		case "1":
			c.createUser()
		case "2":
			c.updateUser()
		case "3":
			c.deleteUser()
		case "4":
			c.listUsers()
		case "0":
			fmt.Println("До свидания!")
			return
		default:
			fmt.Println("Неверный выбор")
		}
	}
}

func (c *RestClient) doRequest(method, path string, body []byte) ([]byte, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	}
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return respBody, resp.StatusCode, nil
}

func (c *RestClient) createUser() {
	name := ReadInput("Имя: ")
	if name == "" {
		fmt.Println("Имя не может быть пустым")
		return
	}

	user := User{Name: name}
	body, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	respBody, statusCode, err := c.doRequest(http.MethodPost, "/users", body)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	if statusCode != http.StatusCreated {
		fmt.Printf("Ошибка: статус %d\n", statusCode)
		return
	}

	var created User
	if err := json.Unmarshal(respBody, &created); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	fmt.Printf("✅ Создан: ID=%d, Name=%s\n", created.ID, created.Name)
}

func (c *RestClient) updateUser() {
	id := ReadInt("ID: ")
	if id == 0 {
		fmt.Println("ID не может быть пустым")
		return
	}

	name := ReadInput("Новое имя (Enter - пропустить): ")

	if name == "" {
		fmt.Println("Укажите хотя бы одно поле")
		return
	}

	user := User{Name: name}
	body, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	respBody, statusCode, err := c.doRequest(http.MethodPut, fmt.Sprintf("/users/%d", id), body)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	if statusCode != http.StatusOK {
		fmt.Printf("Ошибка: статус %d\n", statusCode)
		return
	}

	var updated User
	if err := json.Unmarshal(respBody, &updated); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	fmt.Printf("✅ Обновлен: ID=%d, Name=%s\n", updated.ID, updated.Name)
}

func (c *RestClient) deleteUser() {
	id := ReadInt("ID для удаления: ")
	if id == 0 {
		fmt.Println("ID не может быть пустым")
		return
	}

	confirm := ReadInput(fmt.Sprintf("Удалить ID=%d? (y/N): ", id))
	if strings.ToLower(confirm) != "y" {
		fmt.Println("Отменено")
		return
	}

	_, statusCode, err := c.doRequest(http.MethodDelete, fmt.Sprintf("/users/%d", id), nil)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	if statusCode != http.StatusNoContent {
		fmt.Printf("Ошибка: статус %d\n", statusCode)
		return
	}

	fmt.Printf("✅ Удален пользователь ID=%d\n", id)
}

func (c *RestClient) listUsers() {
	respBody, statusCode, err := c.doRequest(http.MethodGet, "/users", nil)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	if statusCode != http.StatusOK {
		fmt.Printf("Ошибка: статус %d\n", statusCode)
		return
	}

	var users []User
	if err := json.Unmarshal(respBody, &users); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	if len(users) == 0 {
		fmt.Println("Нет пользователей")
		return
	}

	fmt.Printf("\nВсего: %d\n", len(users))
	fmt.Println(strings.Repeat("-", 40))
	for _, u := range users {
		fmt.Printf("ID: %d | Имя: %s\n", u.ID, u.Name)
	}
	fmt.Println(strings.Repeat("-", 40))
}
