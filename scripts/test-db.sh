#!/bin/bash

# ============================================
# Скрипт тестирования подключения к БД
# ============================================

set -e

echo "🧪 Тестирование подключения к PostgreSQL..."
echo ""

# Проверяем, что контейнер запущен
if ! docker ps | grep -q quotopia-postgres; then
    echo "❌ PostgreSQL контейнер не запущен!"
    echo "   Запустите: make up"
    exit 1
fi

echo "✅ PostgreSQL контейнер запущен"
echo ""

# Проверяем подключение
echo "📡 Проверяем подключение..."
if docker exec quotopia-postgres pg_isready -U admin -d quotopia > /dev/null 2>&1; then
    echo "✅ Подключение успешно"
else
    echo "❌ Не удалось подключиться"
    exit 1
fi

echo ""
echo "📊 Статистика БД:"
docker exec quotopia-postgres psql -U admin -d quotopia -c "
SELECT 
    (SELECT COUNT(*) FROM users) as users_count,
    (SELECT COUNT(*) FROM instruments) as instruments_count,
    (SELECT COUNT(*) FROM instruments WHERE is_active = true) as active_instruments;
"

echo ""
echo "👥 Пользователи:"
docker exec quotopia-postgres psql -U admin -d quotopia -c "
SELECT id, email, role, is_active FROM users ORDER BY id;
"

echo ""
echo "📈 Инструменты:"
docker exec quotopia-postgres psql -U admin -d quotopia -c "
SELECT id, symbol, name, initial_price, volatility, is_active FROM instruments ORDER BY symbol;
"

echo ""
echo "✅ Все проверки пройдены!"
echo ""
echo "🌐 Adminer доступен на: http://localhost:8081"
echo "   Server: postgres"
echo "   Username: admin"
echo "   Password: secret123"
echo "   Database: quotopia"
