#!/bin/bash

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Генерация proto файлов ===${NC}"

# Проверка наличия protoc
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}Ошибка: protoc не установлен${NC}"
    echo "Установи protobuf compiler:"
    echo "  Ubuntu/Debian: sudo apt install -y protobuf-compiler"
    echo "  macOS: brew install protobuf"
    exit 1
fi

echo -e "${YELLOW}Версия protoc: $(protoc --version)${NC}"

# Установка/обновление Go плагинов для protoc
echo -e "${YELLOW}Установка Go плагинов...${NC}"
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Проверка наличия плагинов
if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${RED}Ошибка: protoc-gen-go не найден в PATH${NC}"
    echo "Добавь \$GOPATH/bin в PATH:"
    echo "  export PATH=\$PATH:\$(go env GOPATH)/bin"
    exit 1
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${RED}Ошибка: protoc-gen-go-grpc не найден в PATH${NC}"
    echo "Добавь \$GOPATH/bin в PATH:"
    echo "  export PATH=\$PATH:\$(go env GOPATH)/bin"
    exit 1
fi

# Создание директории для сгенерированных файлов
mkdir -p proto/quotes

# Генерация Go кода
echo -e "${YELLOW}Генерация Go кода из proto файлов...${NC}"
protoc \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    proto/quotes.proto

echo -e "${GREEN}✓ Генерация завершена успешно!${NC}"
echo -e "${GREEN}Сгенерированные файлы:${NC}"
ls -lh proto/quotes.pb.go proto/quotes_grpc.pb.go 2>/dev/null || echo "  proto/quotes.pb.go"
