FROM golang:1.23-alpine AS builder

WORKDIR /app

# Установка зависимостей
COPY go.mod ./
RUN go mod download

# Копирование кода
COPY . .

# Сборка
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service .

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Установка ca-certificates для HTTPS
RUN apk --no-cache add ca-certificates

# Копирование бинарника
COPY --from=builder /app/auth-service .

EXPOSE 8090

CMD ["./auth-service"]
