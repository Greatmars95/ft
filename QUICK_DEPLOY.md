# 🚀 Быстрый деплой на production

## ✅ DNS уже настроены:
```
ft           → 45.132.255.115
api.ft       → 45.132.255.115
auth.ft      → 45.132.255.115
admin.ft     → 45.132.255.115
```

## 📋 Шаги деплоя

### 1. Подключитесь к серверу
```bash
ssh root@45.132.255.115
```

### 2. Установите Docker (если нет)
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
docker --version
```

### 3. Клонируйте/обновите проект
```bash
cd ~
git clone https://github.com/Greatmars95/ft.git || (cd ft && git pull)
cd ft
```

### 4. Создайте .env файл
```bash
cp .env.server .env
nano .env
```

**Заполните:**
```env
DB_PASSWORD=<сгенерируйте>
JWT_SECRET=<openssl rand -hex 64>
CERTBOT_EMAIL=your-email@example.com
ADMIN_PASSWORD=<сильный пароль>
```

**Генерация:**
```bash
# Пароль БД
openssl rand -base64 32

# JWT Secret
openssl rand -hex 64
```

### 5. Запустите всё
```bash
# Сначала без nginx (для SSL)
docker-compose up -d postgres auth ft ht ui adminer

# Подождать 10 секунд
sleep 10

# Проверить
docker-compose ps
```

### 6. Настройте домены и SSL
```bash
./scripts/setup-domains.sh
```

Скрипт автоматически:
- ✅ Создаст nginx конфигурацию
- ✅ Получит SSL сертификаты
- ✅ Настроит Basic Auth для Adminer
- ✅ Запустит всё с HTTPS

### 7. Проверьте
```bash
curl -I https://ft
curl https://api.ft/quotes
curl https://auth.ft/health
```

## 🎉 Готово!

Теперь доступно:
- https://ft → UI (котировки в реальном времени)
- https://api.ft → REST API
- https://auth.ft → Auth Service (JWT)
- https://admin.ft → Adminer (Basic Auth)

## 🔐 Создайте первого админа

```bash
curl -X POST https://auth.ft/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@ft",
    "password": "ваш-сильный-пароль",
    "role": "admin"
  }'
```

Сохраните полученный token!

## 📊 Мониторинг

```bash
# Логи
docker-compose logs -f

# Статус
docker-compose ps

# Перезапуск
docker-compose restart
```

## 🔥 Troubleshooting

### SSL не получается
```bash
# Проверить DNS
nslookup ft
nslookup api.ft

# Логи certbot
docker-compose logs certbot
```

### Порты заняты
```bash
# Остановить старые контейнеры
docker-compose down
docker system prune -f

# Запустить заново
docker-compose up -d
```
