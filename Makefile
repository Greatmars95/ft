.PHONY: proto run build clean docker-build docker-run test

# Генерация proto файлов
proto:
	@./scripts/gen-proto.sh

# Запуск локально
run: proto
	go run main.go

# Сборка бинарника
build: proto
	go build -o ft-mt main.go

# Очистка сгенерированных файлов
clean:
	rm -f proto/*.pb.go
	rm -f ft-mt

# Docker сборка
docker-build:
	docker build -t ft-mt:latest .

# Docker запуск
docker-run:
	docker run -p 50051:50051 ft-mt:latest

# Обновление зависимостей
deps:
	go mod download
	go mod tidy

# Проверка кода
lint:
	go vet ./...
	go fmt ./...

# Docker Compose команды
up:
	docker-compose up --build

down:
	docker-compose down

logs:
	docker-compose logs -f

restart:
	docker-compose restart

# Справка
help:
	@echo "Доступные команды:"
	@echo "  make proto        - Генерация Go кода из proto файлов"
	@echo "  make run          - Запуск сервера локально"
	@echo "  make build        - Сборка бинарника"
	@echo "  make clean        - Очистка сгенерированных файлов"
	@echo "  make docker-build - Сборка Docker образа"
	@echo "  make docker-run   - Запуск в Docker"
	@echo "  make up           - Запуск всех сервисов через docker-compose"
	@echo "  make down         - Остановка всех сервисов"
	@echo "  make logs         - Просмотр логов всех сервисов"
	@echo "  make restart      - Перезапуск всех сервисов"
	@echo "  make deps         - Обновление зависимостей"
	@echo "  make lint         - Проверка кода"
