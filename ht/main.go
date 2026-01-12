package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    pb "ft-mt/proto"
)

func main() {
    // Получаем адрес gRPC сервера из переменной окружения
    grpcAddr := os.Getenv("GRPC_SERVER")
    if grpcAddr == "" {
        grpcAddr = "localhost:50051" // Fallback для локальной разработки
    }

    log.Printf("🔌 Подключаюсь к gRPC серверу: %s", grpcAddr)

    conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("❌ Не удалось подключиться к gRPC серверу: %v", err)
    }
    defer conn.Close()

    client := pb.NewQuoteServiceClient(conn)
    r := gin.Default()

    // Добавляем CORS для локальной разработки
    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    })

    r.GET("/quotes", func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        stream, err := client.StreamQuotes(ctx, &pb.QuoteRequest{
            Symbols: []string{"BTC", "ETH", "SBER"},
        })
        if err != nil {
            log.Printf("❌ Ошибка создания стрима: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // Собираем только последние котировки для каждого символа
        latestQuotes := make(map[string]gin.H)

        for {
            quote, err := stream.Recv()
            if err != nil {
                break
            }
            // Перезаписываем, чтобы сохранить только последнее значение
            latestQuotes[quote.Symbol] = gin.H{
                "symbol":    quote.Symbol,
                "price":     quote.Price,
                "timestamp": quote.Timestamp,
            }
        }

        // Преобразуем map в slice
        quotes := make([]gin.H, 0, len(latestQuotes))
        for _, quote := range latestQuotes {
            quotes = append(quotes, quote)
        }

        log.Printf("📊 Отправлено %d котировок", len(quotes))
        c.JSON(http.StatusOK, quotes)
    })

    log.Println("🚀 HT (HTTP Gateway) запущен на порту 8080")
    r.Run(":8080")
}
