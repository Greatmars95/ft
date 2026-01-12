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

# Database команды
db-shell:
	docker exec -it quotopia-postgres psql -U admin -d quotopia

db-backup:
	docker exec quotopia-postgres pg_dump -U admin quotopia > backup_$(shell date +%Y%m%d_%H%M%S).sql

db-restore:
	@echo "Использование: make db-restore FILE=backup.sql"
	docker exec -i quotopia-postgres psql -U admin quotopia < $(FILE)

db-reset:
	docker-compose down -v
	docker-compose up -d postgres
	@echo "База данных очищена и пересоздана"

# Полезные команды
ps:
	docker-compose ps

adminer:
	@echo "Adminer доступен на: http://localhost:8081"
	@echo "Server: postgres"
	@echo "Username: admin"
	@echo "Password: secret123"
	@echo "Database: quotopia"

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
