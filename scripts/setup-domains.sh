#!/bin/bash

# ============================================
# Скрипт настройки доменов и SSL
# ============================================

set -e

echo "╔══════════════════════════════════════════════════╗"
echo "║   🌐 Настройка доменов и SSL для Quotopia       ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""

# Проверка наличия .env файла
if [ ! -f .env ]; then
    echo "⚠️  Файл .env не найден!"
    echo ""
    echo "Создайте .env файл:"
    echo "  cp .env.production.example .env"
    echo "  nano .env  # Отредактируйте"
    echo ""
    exit 1
fi

# Загрузка переменных
source .env

# Проверка обязательных переменных
if [ -z "$DOMAIN" ]; then
    echo "❌ DOMAIN не установлен в .env"
    exit 1
fi

if [ -z "$CERTBOT_EMAIL" ]; then
    echo "❌ CERTBOT_EMAIL не установлен в .env"
    exit 1
fi

echo "📋 Конфигурация:"
echo "   Основной домен: $DOMAIN"
echo "   API домен:      api.$DOMAIN"
echo "   Auth домен:     auth.$DOMAIN"
echo "   Admin домен:    admin.$DOMAIN"
echo "   Email:          $CERTBOT_EMAIL"
echo ""

# Шаг 1: Создание nginx конфигурации
echo "1️⃣  Создание nginx конфигурации..."
cp nginx/conf.d/quotopia.conf.template nginx/conf.d/quotopia.conf
sed -i "s/DOMAIN/$DOMAIN/g" nginx/conf.d/quotopia.conf
echo "   ✅ nginx/conf.d/quotopia.conf создан"
echo ""

# Шаг 2: Создание .htpasswd для Adminer
echo "2️⃣  Создание Basic Auth для Adminer..."
if [ -z "$ADMIN_PASSWORD" ]; then
    echo "   ⚠️  ADMIN_PASSWORD не установлен, используется 'admin123'"
    ADMIN_PASSWORD="admin123"
fi

docker run --rm httpd:alpine htpasswd -Bbn admin "$ADMIN_PASSWORD" > nginx/.htpasswd
echo "   ✅ nginx/.htpasswd создан"
echo ""

# Шаг 3: Запуск nginx (без SSL пока)
echo "3️⃣  Запуск Nginx (для получения сертификатов)..."
docker-compose up -d nginx
sleep 5
echo "   ✅ Nginx запущен"
echo ""

# Шаг 4: Получение SSL сертификатов
echo "4️⃣  Получение SSL сертификатов..."
echo ""

echo "   📜 Получение сертификата для $DOMAIN..."
docker-compose run --rm certbot certonly --webroot \
    --webroot-path=/var/www/certbot \
    --email "$CERTBOT_EMAIL" \
    --agree-tos \
    --no-eff-email \
    -d "$DOMAIN" \
    -d "www.$DOMAIN"

echo ""
echo "   📜 Получение сертификата для api.$DOMAIN..."
docker-compose run --rm certbot certonly --webroot \
    --webroot-path=/var/www/certbot \
    --email "$CERTBOT_EMAIL" \
    --agree-tos \
    --no-eff-email \
    -d "api.$DOMAIN"

echo ""
echo "   📜 Получение сертификата для auth.$DOMAIN..."
docker-compose run --rm certbot certonly --webroot \
    --webroot-path=/var/www/certbot \
    --email "$CERTBOT_EMAIL" \
    --agree-tos \
    --no-eff-email \
    -d "auth.$DOMAIN"

echo ""
echo "   📜 Получение сертификата для admin.$DOMAIN..."
docker-compose run --rm certbot certonly --webroot \
    --webroot-path=/var/www/certbot \
    --email "$CERTBOT_EMAIL" \
    --agree-tos \
    --no-eff-email \
    -d "admin.$DOMAIN"

echo ""
echo "   ✅ Все сертификаты получены!"
echo ""

# Шаг 5: Перезапуск nginx с SSL
echo "5️⃣  Перезапуск Nginx с SSL..."
docker-compose restart nginx
echo "   ✅ Nginx перезапущен с SSL"
echo ""

# Проверка
echo "╔══════════════════════════════════════════════════╗"
echo "║   ✅ Настройка завершена успешно!                ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""
echo "🌐 Ваши сервисы доступны:"
echo "   https://$DOMAIN              → UI"
echo "   https://api.$DOMAIN          → API"
echo "   https://auth.$DOMAIN         → Auth Service"
echo "   https://admin.$DOMAIN        → Adminer (Basic Auth)"
echo ""
echo "🔐 Adminer логин:"
echo "   Username: admin"
echo "   Password: (ваш ADMIN_PASSWORD из .env)"
echo ""
echo "📊 Проверить SSL:"
echo "   https://www.ssllabs.com/ssltest/analyze.html?d=$DOMAIN"
echo ""
