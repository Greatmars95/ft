# Этап 1: Сборка
FROM golang:1.23-alpine AS builder

# Установка необходимых инструментов
RUN apk add --no-cache git protobuf protobuf-dev

# Установка Go плагинов для protoc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Рабочая директория
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum* ./

# Загружаем зависимости
RUN go mod download

# Копируем proto файлы и скрипт генерации
COPY proto/ ./proto/
COPY scripts/ ./scripts/

# Генерируем Go код из proto
RUN chmod +x ./scripts/gen-proto.sh && \
    export PATH=$PATH:$(go env GOPATH)/bin && \
    protoc \
        --go_out=. \
        --go_opt=paths=source_relative \
        --go-grpc_out=. \
        --go-grpc_opt=paths=source_relative \
        proto/quotes.proto

# Копируем исходный код
COPY main.go ./

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ft-mt .

# Этап 2: Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем скомпилированный бинарник
COPY --from=builder /app/ft-mt .

# Открываем порт gRPC
EXPOSE 50051

# Запускаем приложение
CMD ["./ft-mt"]
