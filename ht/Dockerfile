FROM golang:1.24-alpine AS builder

# Устанавливаем protoc и необходимые инструменты
RUN apk add --no-cache protobuf protobuf-dev

WORKDIR /app

# Копируем proto файлы из корневой директории
COPY proto ./proto

# Устанавливаем protoc плагины
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Генерируем proto файлы
RUN protoc \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    proto/quotes.proto

# Копируем go.mod и go.sum для HT сервиса
COPY ht/go.mod ht/go.sum ./
RUN go mod download

# Копируем исходный код HT
COPY ht/main.go ./

# Собираем бинарник
RUN go build -o ht main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/ht .
EXPOSE 8080
CMD ["./ht"]
