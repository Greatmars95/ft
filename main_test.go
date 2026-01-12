package main

import (
	"testing"
)

// TestNewQuoteServer проверяет создание нового сервера
func TestNewQuoteServer(t *testing.T) {
	server := NewQuoteServer()

	// Проверяем, что сервер создался
	if server == nil {
		t.Fatal("NewQuoteServer() вернул nil")
	}

	// Проверяем, что quotes карта инициализирована
	if server.quotes == nil {
		t.Fatal("quotes карта не инициализирована")
	}

	// Проверяем, что есть ровно 3 тикера
	if len(server.quotes) != 3 {
		t.Errorf("Ожидалось 3 тикера, получено %d", len(server.quotes))
	}
}

// TestInitialPrices проверяет начальные цены
func TestInitialPrices(t *testing.T) {
	server := NewQuoteServer()

	// Проверяем начальные цены
	tests := []struct {
		symbol        string
		expectedPrice float64
	}{
		{"SBER", 275.50},
		{"BTC", 95400.0},
		{"ETH", 2650.20},
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			price, exists := server.quotes[tt.symbol]

			if !exists {
				t.Errorf("Тикер %s не найден", tt.symbol)
			}

			if price != tt.expectedPrice {
				t.Errorf("Для %s ожидалась цена %.2f, получено %.2f",
					tt.symbol, tt.expectedPrice, price)
			}
		})
	}
}

// TestPriceIsPositive проверяет, что все цены положительные
func TestPriceIsPositive(t *testing.T) {
	server := NewQuoteServer()

	for symbol, price := range server.quotes {
		if price <= 0 {
			t.Errorf("Цена для %s должна быть положительной, получено %.2f",
				symbol, price)
		}
	}
}

// TestUnknownSymbol проверяет поведение с неизвестным символом
func TestUnknownSymbol(t *testing.T) {
	server := NewQuoteServer()

	// Проверяем, что неизвестный символ не существует
	_, exists := server.quotes["UNKNOWN"]
	if exists {
		t.Error("Неизвестный символ UNKNOWN не должен существовать")
	}
}

// TestAllSymbolsExist проверяет наличие всех ожидаемых символов
func TestAllSymbolsExist(t *testing.T) {
	server := NewQuoteServer()
	requiredSymbols := []string{"SBER", "BTC", "ETH"}

	for _, symbol := range requiredSymbols {
		if _, exists := server.quotes[symbol]; !exists {
			t.Errorf("Обязательный символ %s не найден", symbol)
		}
	}
}
