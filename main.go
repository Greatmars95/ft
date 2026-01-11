package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	pb "ft-mt/proto"

	"google.golang.org/grpc"
)

// QuoteServer реализует gRPC сервис генерации котировок
type QuoteServer struct {
	pb.UnimplementedQuoteServiceServer
	quotes map[string]float64
}

// NewQuoteServer создаёт новый сервер с начальными ценами
func NewQuoteServer() *QuoteServer {
	return &QuoteServer{
		quotes: map[string]float64{
			"SBER": 275.50,
			"BTC":  95400.0,
			"ETH":  2650.20,
		},
	}
}

// StreamQuotes реализует стриминг котировок
func (s *QuoteServer) StreamQuotes(req *pb.QuoteRequest, stream pb.QuoteService_StreamQuotesServer) error {
	log.Printf("Новое подключение. Запрошенные символы: %v", req.Symbols)

	// Определяем какие символы отправлять
	symbols := req.Symbols
	if len(symbols) == 0 {
		// Если не указано - отправляем все
		for symbol := range s.quotes {
			symbols = append(symbols, symbol)
		}
	}

	// Бесконечный стрим котировок
	for {
		for _, symbol := range symbols {
			oldPrice, exists := s.quotes[symbol]
			if !exists {
				log.Printf("Символ %s не найден, пропускаем", symbol)
				continue
			}

			// Изменяем цену на случайный процент от -0.1% до +0.1%
			change := (rand.Float64() * 0.002) - 0.001
			newPrice := oldPrice * (1 + change)
			s.quotes[symbol] = newPrice

			quote := &pb.Quote{
				Symbol:    symbol,
				Price:     newPrice,
				Timestamp: time.Now().UnixMilli(),
			}

			// Отправляем котировку в стрим
			if err := stream.Send(quote); err != nil {
				log.Printf("Ошибка отправки: %v", err)
				return err
			}

			log.Printf("Отправлено: %s = %.2f", symbol, newPrice)
		}

		// Пауза между обновлениями
		time.Sleep(1 * time.Second)
	}
}

func main() {
	// Создаём TCP listener
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Ошибка создания listener: %v", err)
	}

	// Создаём gRPC сервер
	grpcServer := grpc.NewServer()
	quoteServer := NewQuoteServer()
	pb.RegisterQuoteServiceServer(grpcServer, quoteServer)

	fmt.Println("🚀 FT (Quote Generator) запущен на порту 50051")
	fmt.Println("📊 Доступные тикеры:", []string{"SBER", "BTC", "ETH"})
	fmt.Println("⏳ Ожидание подключений...")

	// Запускаем сервер
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
