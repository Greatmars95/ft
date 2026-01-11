# ========================================
# ЭТАП 1: Сборка (builder)
# ========================================
# Используем образ с Go 1.23 на базе Alpine Linux (лёгкий дистрибутив)
FROM golang:1.23-alpine AS builder

# Устанавливаем protoc — компилятор proto файлов
RUN apk add --no-cache protobuf protobuf-dev make

# Устанавливаем Go плагины для protoc
# Они генерируют Go код из .proto файлов
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod (и go.sum если есть)
# COPY <источник> <назначение>
COPY go.mod go.sum* ./

# Скачиваем зависимости
# Docker кеширует этот слой, поэтому пересборка будет быстрее
RUN go mod download

# Копируем proto файл из корня проекта
# context в docker-compose.yml установлен в корень, поэтому путь без ../
COPY proto/quotes.proto ./proto/

# Генерируем Go код из proto файла
# --go_out = генерируем структуры данных
# --go-grpc_out = генерируем gRPC код (сервер/клиент)
RUN mkdir -p gen/quotes
RUN protoc --go_out=./gen/quotes --go_opt=paths=source_relative \
           --go-grpc_out=./gen/quotes --go-grpc_opt=paths=source_relative \
           proto/quotes.proto

# Копируем исходный код сервиса
COPY . .

# Собираем бинарник
# CGO_ENABLED=0 — статическая линковка (не зависит от системных библиотек)
# -ldflags="-s -w" — уменьшаем размер бинарника (убираем debug info)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /ft main.go

# ========================================
# ЭТАП 2: Запуск (runtime)
# ========================================
# Используем минимальный Alpine образ (~5 MB)
FROM alpine:3.19

# Устанавливаем CA сертификаты (нужны для HTTPS)
RUN apk --no-cache add ca-certificates

# Рабочая директория
WORKDIR /app

# Копируем ТОЛЬКО бинарник из builder этапа
# Всё остальное (исходники, зависимости) не попадает в финальный образ
COPY --from=builder /ft .

# Порт gRPC сервера
EXPOSE 50051

# Команда запуска
CMD ["./ft"]
