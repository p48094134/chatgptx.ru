package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Структура для запроса к API OpenAI
type OpenAIRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	MaxTokens int          `json:"max_tokens,omitempty"` // Опционально: максимальное количество токенов в ответе
	// Temperature float64    `json:"temperature,omitempty"` // Опционально: "креативность" ответа
}

// Структура для сообщения в диалоге
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Структура для ответа от API OpenAI
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

const openAIAPIURL = "https://api.openai.com/v1/chat/completions"

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Переменная окружения OPENAI_API_KEY не установлена.")
	}

	userPrompt := "Напиши короткий рассказ о роботе, который научился мечтать." // Ваш промпт здесь

	// Формируем запрос
	requestPayload := OpenAIRequest{
		Model: "gpt-3.5-turbo", // или другая доступная модель, например "gpt-4"
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a helpful assistant."}, // Системное сообщение (опционально)
			{Role: "user", Content: userPrompt},
		},
		MaxTokens: 150, // Ограничим длину ответа
	}

	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		log.Fatalf("Ошибка при маршалинге JSON: %v", err)
	}

	// Создаем HTTP запрос
	req, err := http.NewRequest("POST", openAIAPIURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Fatalf("Ошибка при создании HTTP запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка при выполнении HTTP запроса: %v", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка при чтении тела ответа: %v", err)
	}

	// Проверяем статус код
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("API вернул ошибку: %s\nТело ответа: %s", resp.Status, string(body))
	}

	// Парсим JSON ответ
	var openAIResp OpenAIResponse
	err = json.Unmarshal(body, &openAIResp)
	if err != nil {
		log.Fatalf("Ошибка при демаршалинге JSON ответа: %v\nТело ответа: %s", err, string(body))
	}

	// Выводим ответ от ассистента
	if len(openAIResp.Choices) > 0 {
		assistantResponse := openAIResp.Choices[0].Message.Content
		fmt.Println("Ответ от ChatGPT:")
		fmt.Println(assistantResponse)
		fmt.Printf("\nИспользовано токенов: %d (промпт: %d, ответ: %d)\n",
			openAIResp.Usage.TotalTokens,
			openAIResp.Usage.PromptTokens,
			openAIResp.Usage.CompletionTokens,
		)
	} else {
		fmt.Println("Не удалось получить ответ от API.")
		fmt.Println("Тело ответа:", string(body))
	}
}
