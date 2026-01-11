package main

import (
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	// Импортируем сгенерированный из proto код
	// Этот код создастся автоматически из quotes.proto
	pb "ft/gen/quotes"

	"google.golang.org/grpc"
)

// ============================================
// QuoteServer — наш gRPC сервер
// Он реализует интерфейс из proto файла
// ============================================
type QuoteServer struct {
	pb.UnimplementedQuoteServiceServer // Встраиваем базовую реализацию

	// Текущие котировки — общие для всех клиентов
	// map[символ]цена
	quotes map[string]float64
	mu     sync.RWMutex // Мьютекс для безопасного доступа из разных горутин
}

// NewQuoteServer создаёт новый сервер с начальными ценами
func NewQuoteServer() *QuoteServer {
	return &QuoteServer{
		quotes: map[string]float64{
			"SBER": 275.50,  // Начальная цена Сбербанка
			"BTC":  95400.0, // Начальная цена Bitcoin
			"ETH":  2650.20, // Начальная цена Ethereum
		},
	}
}

// StreamQuotes — реализация метода из proto
// Этот метод вызывается когда клиент подключается
// stream — это канал для отправки котировок клиенту
func (s *QuoteServer) StreamQuotes(req *pb.StreamRequest, stream pb.QuoteService_StreamQuotesServer) error {
	log.Println("✅ Новый клиент подключился к стриму")

	// Бесконечный цикл отправки котировок
	for {
		// Блокируем для чтения текущих котировок
		s.mu.RLock()

		// Отправляем каждую котировку
		for symbol, price := range s.quotes {
			quote := &pb.Quote{
				Symbol:    symbol,
				Price:     price,
				Timestamp: time.Now().UnixMilli(), // Unix время в миллисекундах
			}

			// Отправляем котировку в стрим
			// Если клиент отключился, вернётся ошибка
			if err := stream.Send(quote); err != nil {
				log.Printf("❌ Клиент отключился: %v", err)
				s.mu.RUnlock()
				return err
			}
		}

		s.mu.RUnlock()

		// Ждём 1 секунду перед следующей итерацией
		time.Sleep(1 * time.Second)
	}
}

// updateQuotes — обновляет цены котировок
// Запускается в отдельной горутине и работает постоянно
func (s *QuoteServer) updateQuotes() {
	ticker := time.NewTicker(1 * time.Second) // Таймер на 1 секунду
	defer ticker.Stop()

	log.Println("📊 Генератор котировок запущен")

	for range ticker.C { // Каждую секунду
		s.mu.Lock() // Блокируем для записи

		// Обновляем каждую котировку
		for symbol, oldPrice := range s.quotes {
			// Изменяем цену на -0.1% до +0.1%
			change := (rand.Float64() * 0.002) - 0.001
			newPrice := oldPrice * (1 + change)
			s.quotes[symbol] = newPrice

			log.Printf("📈 %s: %.2f → %.2f (%.2f%%)",
				symbol, oldPrice, newPrice, change*100)
		}

		s.mu.Unlock() // Разблокируем
	}
}

func main() {
	// Инициализируем генератор случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Создаём наш сервер
	server := NewQuoteServer()

	// Запускаем генератор котировок в фоне
	go server.updateQuotes()

	// Создаём TCP listener на порту 50051
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("❌ Не удалось запустить listener: %v", err)
	}

	// Создаём gRPC сервер
	grpcServer := grpc.NewServer()

	// Регистрируем наш QuoteService
	pb.RegisterQuoteServiceServer(grpcServer, server)

	log.Println("🚀 FT сервис запущен на :50051")
	log.Println("⏳ Ожидание подключений от Core...")

	// Запускаем gRPC сервер (блокирующий вызов)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("❌ Ошибка gRPC сервера: %v", err)
	}
}
