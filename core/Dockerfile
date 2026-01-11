# ========================================
# ЭТАП 1: Сборка (builder)
# ========================================
FROM golang:1.23-alpine AS builder

# Устанавливаем protoc
RUN apk add --no-cache protobuf protobuf-dev make

# Go плагины для protoc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

# Зависимости (для Core нужен ещё Gin)
COPY go.mod go.sum* ./
RUN go mod download

# Proto файл и генерация
COPY ../proto/quotes.proto ./proto/
RUN mkdir -p gen/quotes
RUN protoc --go_out=./gen/quotes --go_opt=paths=source_relative \
           --go-grpc_out=./gen/quotes --go-grpc_opt=paths=source_relative \
           proto/quotes.proto

# Исходники и сборка
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /core main.go

# ========================================
# ЭТАП 2: Запуск (runtime)
# ========================================
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /core .

# REST API порт
EXPOSE 8080

# Запуск
CMD ["./core"]
