#!/bin/bash

# ============================================
# Скрипт для подключения к PostgreSQL
# ============================================

set -e

echo "🔌 Подключение к PostgreSQL..."
echo ""
echo "Способ 1: Через Docker (локально)"
echo "  docker exec -it quotopia-postgres psql -U admin -d quotopia"
echo ""
echo "Способ 2: Через psql (если установлен локально)"
echo "  psql -h localhost -p 5432 -U admin -d quotopia"
echo ""
echo "Способ 3: Через Adminer (веб-интерфейс)"
echo "  Откройте: http://localhost:8081"
echo "  Server: postgres"
echo "  Username: admin"
echo "  Password: secret123"
echo "  Database: quotopia"
echo ""

# Если передан аргумент, выполняем SQL запрос
if [ -n "$1" ]; then
    echo "Выполняем запрос: $1"
    docker exec -it quotopia-postgres psql -U admin -d quotopia -c "$1"
else
    echo "Подключаемся к БД..."
    docker exec -it quotopia-postgres psql -U admin -d quotopia
fi
