package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    pb "ft-mt/proto/quotes"
)

func main() {
    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("не удалось подключиться к gRPC серверу: %v", err)
    }
    defer conn.Close()

    client := pb.NewQuoteServiceClient(conn)
    r := gin.Default()

    r.GET("/quotes", func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()

        stream, err := client.StreamQuotes(ctx, &pb.QuoteRequest{
            Symbols: []string{"BTC", "ETH", "SBER"},
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        quotes := []gin.H{}
        for {
            quote, err := stream.Recv()
            if err != nil {
                break
            }
            quotes = append(quotes, gin.H{
                "symbol":    quote.Symbol,
                "price":     quote.Price,
                "timestamp": quote.Timestamp,
            })
        }

        c.JSON(http.StatusOK, quotes)
    })

    r.Run(":9090")
}
