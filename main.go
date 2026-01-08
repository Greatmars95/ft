package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// Описываем структуру котировки
type Quote struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	// Начальные цены для симуляции
	quotes := map[string]float64{
		"SBER": 275.50,
		"BTC":  95400.0,
		"ETH":  2650.20,
	}

	fmt.Println("Генератор запущен... Нажми Ctrl+C для остановки.")

	for {
		for symbol, oldPrice := range quotes {
			// Изменяем цену на случайный процент от -0.1% до +0.1%
			change := (rand.Float64() * 0.002) - 0.001
			newPrice := oldPrice * (1 + change)
			quotes[symbol] = newPrice

			q := Quote{
				Symbol:    symbol,
				Price:     newPrice,
				Timestamp: time.Now(),
			}

			payload, _ := json.Marshal(q)
			fmt.Println(string(payload))
		}
		
		time.Sleep(1 * time.Second)
	}
}
