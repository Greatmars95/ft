#!/bin/bash

# ============================================
# Скрипт для проброса портов с production
# в Cloud Agent
# ============================================

echo "🔌 Создаём SSH туннели к production серверу..."
echo ""

# Проверка переменных окружения
if [ -z "$SERVER_HOST" ] || [ -z "$SERVER_USER" ]; then
    echo "⚠️  Не установлены переменные окружения!"
    echo ""
    echo "Установите:"
    echo "  export SERVER_HOST=your-server-ip"
    echo "  export SERVER_USER=your-username"
    echo ""
    echo "Или используйте напрямую:"
    echo "  SERVER_HOST=1.2.3.4 SERVER_USER=ubuntu ./scripts/tunnel-to-production.sh"
    exit 1
fi

echo "Подключаемся к: $SERVER_USER@$SERVER_HOST"
echo ""

# Создаём туннели для всех портов
ssh -o StrictHostKeyChecking=no \
    -L 8081:localhost:8081 \
    -L 3001:localhost:3001 \
    -L 8080:localhost:8080 \
    -L 5432:localhost:5432 \
    $SERVER_USER@$SERVER_HOST -N

# Если туннель упал, скрипт завершится
echo ""
echo "❌ Туннель отключен"
