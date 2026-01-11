package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	// Gin — популярный HTTP фреймворк для Go
	"github.com/gin-gonic/gin"

	// Импортируем сгенерированный gRPC код
	pb "core/gen/quotes"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ============================================
// QuoteStore — хранилище котировок в памяти
// ============================================
type QuoteStore struct {
	// map[символ]цена — храним последние цены
	quotes map[string]float64
	mu     sync.RWMutex // Мьютекс для безопасного чтения/записи
}

// NewQuoteStore создаёт новое хранилище
func NewQuoteStore() *QuoteStore {
	return &QuoteStore{
		quotes: make(map[string]float64),
	}
}

// Set сохраняет или обновляет котировку
func (s *QuoteStore) Set(symbol string, price float64) {
	s.mu.Lock()         // Блокируем для записи
	defer s.mu.Unlock() // Разблокируем в конце функции

	s.quotes[symbol] = price
	log.Printf("💾 Обновлена котировка: %s = %.2f", symbol, price)
}

// Get возвращает цену для символа
func (s *QuoteStore) Get(symbol string) (float64, bool) {
	s.mu.RLock()         // Блокируем для чтения
	defer s.mu.RUnlock() // Разблокируем в конце

	price, exists := s.quotes[symbol]
	return price, exists
}

// GetAll возвращает все котировки
func (s *QuoteStore) GetAll() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Копируем map чтобы избежать race condition
	result := make(map[string]float64, len(s.quotes))
	for k, v := range s.quotes {
		result[k] = v
	}
	return result
}

// ============================================
// startGRPCClient — подключается к FT и слушает котировки
// ============================================
func startGRPCClient(store *QuoteStore, ftAddress string) {
	log.Printf("🔌 Подключаюсь к FT сервису: %s", ftAddress)

	// Создаём gRPC соединение
	// insecure.NewCredentials() = без TLS (для разработки)
	conn, err := grpc.Dial(ftAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Не удалось подключиться к FT: %v", err)
	}
	defer conn.Close()

	// Создаём клиента QuoteService
	client := pb.NewQuoteServiceClient(conn)

	// Бесконечный цикл с переподключением
	for {
		log.Println("📡 Открываю стрим котировок...")

		// Создаём контекст (можно отменить через cancel())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Запрашиваем стрим котировок
		stream, err := client.StreamQuotes(ctx, &pb.StreamRequest{})
		if err != nil {
			log.Printf("❌ Ошибка открытия стрима: %v. Переподключаюсь через 5 сек...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("✅ Стрим открыт, получаю котировки...")

		// Читаем стрим
		for {
			// Recv() блокируется пока не придёт новая котировка
			quote, err := stream.Recv()
			if err != nil {
				log.Printf("❌ Стрим закрыт: %v. Переподключаюсь...", err)
				break // Выходим из внутреннего цикла
			}

			// Сохраняем котировку в хранилище
			store.Set(quote.Symbol, quote.Price)
		}

		// Ждём перед переподключением
		time.Sleep(5 * time.Second)
	}
}

// ============================================
// REST API handlers
// ============================================

// healthHandler — проверка здоровья сервиса
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// getAllQuotesHandler — GET /quotes
// Возвращает все котировки
func getAllQuotesHandler(store *QuoteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		quotes := store.GetAll()

		// Если котировок нет (FT ещё не подключился)
		if len(quotes) == 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "no quotes available yet",
			})
			return
		}

		c.JSON(http.StatusOK, quotes)
	}
}

// getQuoteBySymbolHandler — GET /quotes/:symbol
// Возвращает одну котировку
func getQuoteBySymbolHandler(store *QuoteStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Param("symbol") // Достаём параметр из URL

		price, exists := store.Get(symbol)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"error":  "not found",
				"symbol": symbol,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"symbol": symbol,
			"price":  price,
		})
	}
}

func main() {
	// Создаём хранилище котировок
	store := NewQuoteStore()

	// Адрес FT сервиса
	// В Docker это будет "ft:50051"
	// Локально можно использовать "localhost:50051"
	ftAddress := "ft:50051"

	// Запускаем gRPC клиент в фоне
	// go = запуск в отдельной горутине (асинхронно)
	go startGRPCClient(store, ftAddress)

	// Создаём HTTP сервер (Gin)
	gin.SetMode(gin.ReleaseMode) // Отключаем debug логи
	router := gin.Default()

	// Регистрируем маршруты (endpoints)
	router.GET("/health", healthHandler)
	router.GET("/quotes", getAllQuotesHandler(store))
	router.GET("/quotes/:symbol", getQuoteBySymbolHandler(store))

	log.Println("🚀 Core сервис запущен на :8080")
	log.Println("📍 Доступные endpoints:")
	log.Println("   GET /health              - проверка здоровья")
	log.Println("   GET /quotes              - все котировки")
	log.Println("   GET /quotes/:symbol      - одна котировка")

	// Запускаем HTTP сервер (блокирующий вызов)
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("❌ Ошибка HTTP сервера: %v", err)
	}
}
