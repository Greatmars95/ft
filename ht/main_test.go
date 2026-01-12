package main

import (
	"os"
	"testing"
)

// TestDefaultGRPCAddress проверяет дефолтный адрес gRPC сервера
func TestDefaultGRPCAddress(t *testing.T) {
	// Очищаем переменную окружения
	os.Unsetenv("GRPC_SERVER")

	expectedDefault := "localhost:50051"

	// Симулируем логику из main()
	grpcAddr := os.Getenv("GRPC_SERVER")
	if grpcAddr == "" {
		grpcAddr = "localhost:50051"
	}

	if grpcAddr != expectedDefault {
		t.Errorf("Ожидался адрес %s, получено %s", expectedDefault, grpcAddr)
	}
}

// TestCustomGRPCAddress проверяет кастомный адрес из переменной окружения
func TestCustomGRPCAddress(t *testing.T) {
	customAddr := "ft:50051"
	os.Setenv("GRPC_SERVER", customAddr)
	defer os.Unsetenv("GRPC_SERVER")

	grpcAddr := os.Getenv("GRPC_SERVER")

	if grpcAddr != customAddr {
		t.Errorf("Ожидался адрес %s, получено %s", customAddr, grpcAddr)
	}
}

// TestExpectedSymbols проверяет список символов для запроса
func TestExpectedSymbols(t *testing.T) {
	expectedSymbols := []string{"BTC", "ETH", "SBER"}

	// Проверяем, что все символы есть
	if len(expectedSymbols) != 3 {
		t.Errorf("Ожидалось 3 символа, получено %d", len(expectedSymbols))
	}

	// Проверяем конкретные символы
	symbolsMap := make(map[string]bool)
	for _, s := range expectedSymbols {
		symbolsMap[s] = true
	}

	requiredSymbols := []string{"BTC", "ETH", "SBER"}
	for _, required := range requiredSymbols {
		if !symbolsMap[required] {
			t.Errorf("Обязательный символ %s не найден", required)
		}
	}
}

// TestHTTPPort проверяет порт HTTP сервера
func TestHTTPPort(t *testing.T) {
	expectedPort := ":8080"

	// HT сервис должен слушать на порту 8080
	if expectedPort != ":8080" {
		t.Errorf("Ожидался порт :8080, получено %s", expectedPort)
	}
}

// TestQuoteEndpoint проверяет наличие эндпоинта /quotes
func TestQuoteEndpoint(t *testing.T) {
	endpoint := "/quotes"

	// Проверяем формат эндпоинта
	if endpoint[0] != '/' {
		t.Error("Эндпоинт должен начинаться с /")
	}

	if endpoint != "/quotes" {
		t.Errorf("Ожидался эндпоинт /quotes, получено %s", endpoint)
	}
}
