FROM golang:1.24-alpine AS builder

# Устанавливаем protoc и необходимые инструменты
RUN apk add --no-cache protobuf protobuf-dev

WORKDIR /workspace

# Копируем корневые файлы проекта
COPY go.mod go.sum ./

# Копируем proto файлы и генерируем код
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

# Копируем файлы HT сервиса
COPY ht ./ht

# Переходим в директорию HT и собираем
WORKDIR /workspace/ht
RUN go mod download
RUN go build -o ht main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /workspace/ht/ht .
EXPOSE 8080
CMD ["./ht"]
