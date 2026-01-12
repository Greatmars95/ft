# 🚀 Production Setup - Пошаговая инструкция

## 📋 Что нужно перед началом

- ✅ Сервер с Ubuntu 20.04+ (ваш: 45.132.255.115)
- ✅ Домен (например: quotopia.com)
- ✅ Доступ к DNS настройкам
- ✅ SSH доступ к серверу

---

## Шаг 1: Настройка DNS

Добавьте A-записи в DNS:

```
quotopia.com            A    45.132.255.115
api.quotopia.com        A    45.132.255.115
auth.quotopia.com       A    45.132.255.115
admin.quotopia.com      A    45.132.255.115
```

**Проверка:**
```bash
nslookup quotopia.com
# Должен вернуть: 45.132.255.115
```

---

## Шаг 2: Подготовка сервера

### 2.1. Подключитесь к серверу

```bash
ssh root@45.132.255.115
```

### 2.2. Установите Docker

```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Проверка
docker --version
docker-compose --version
```

### 2.3. Клонируйте проект

```bash
cd ~
git clone https://github.com/Greatmars95/ft.git
cd ft
git checkout main
```

---

## Шаг 3: Конфигурация

### 3.1. Создайте .env файл

```bash
cp .env.production.example .env
nano .env
```

**Заполните:**

```env
# Database
DB_PASSWORD=<сгенерируйте сильный пароль>

# JWT
JWT_SECRET=<openssl rand -hex 64>

# Domains
DOMAIN=quotopia.com  # ВАШ ДОМЕН!

# SSL
CERTBOT_EMAIL=your-email@example.com

# Adminer
ADMIN_PASSWORD=<сильный пароль>
```

**Генерация паролей:**
```bash
# Сильный пароль для БД
openssl rand -base64 32

# JWT Secret
openssl rand -hex 64
```

### 3.2. Проверьте конфигурацию

```bash
cat .env | grep -v '^#' | grep -v '^$'
```

---

## Шаг 4: Запуск (БЕЗ SSL)

Сначала запустим без SSL для получения сертификатов:

```bash
# Запустить БД и сервисы
docker-compose up -d postgres auth ft ht ui

# Подождать 10 секунд
sleep 10

# Проверить
docker-compose ps
```

Должны быть запущены:
- quotopia-postgres
- quotopia-auth
- quotopia-ft-1
- quotopia-ht-1
- quotopia-ui-1

---

## Шаг 5: Настройка доменов и SSL

**Автоматический скрипт:**

```bash
./scripts/setup-domains.sh
```

Скрипт сделает:
1. Создаст nginx конфигурацию
2. Создаст Basic Auth для Adminer
3. Запустит Nginx
4. Получит SSL сертификаты для всех доменов
5. Перезапустит Nginx с SSL

**Время:** ~5-10 минут

---

## Шаг 6: Проверка

### 6.1. Проверить сервисы

```bash
docker-compose ps
```

Должны быть запущены:
- postgres
- auth
- ft
- ht
- ui
- nginx
- certbot

### 6.2. Проверить домены

```bash
# UI
curl -I https://quotopia.com

# API
curl https://api.quotopia.com/quotes

# Auth
curl https://auth.quotopia.com/health

# Adminer (попросит Basic Auth)
curl -I https://admin.quotopia.com
```

### 6.3. Открыть в браузере

```
https://quotopia.com          → UI (котировки)
https://api.quotopia.com      → API
https://auth.quotopia.com     → Auth Service
https://admin.quotopia.com    → Adminer (Basic Auth)
```

---

## Шаг 7: Первый пользователь

Создайте admin аккаунт:

```bash
curl -X POST https://auth.quotopia.com/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@quotopia.com",
    "password": "your-secure-password",
    "role": "admin"
  }'
```

Сохраните полученный токен!

---

## Шаг 8: Настройка CI/CD

### 8.1. Добавьте GitHub Secrets

```
Settings → Secrets → Actions → New repository secret

Secrets:
  SERVER_HOST=45.132.255.115
  SERVER_USER=root
  SSH_PRIVATE_KEY=<ваш приватный SSH ключ>
```

### 8.2. Проверьте CI/CD

Сделайте любой коммит в main → GitHub Actions запустится → автодеплой!

---

## 🔒 Безопасность (После запуска)

### Firewall

```bash
# Разрешить только нужные порты
ufw allow 22/tcp   # SSH
ufw allow 80/tcp   # HTTP
ufw allow 443/tcp  # HTTPS
ufw enable

# Проверить
ufw status
```

### Смена паролей

```bash
# Зайти в БД
docker exec -it quotopia-postgres psql -U admin -d quotopia

# Обновить пароль админа
UPDATE users 
SET password_hash = '$2a$10$...' 
WHERE email = 'admin@quotopia.com';
```

### Регулярные бэкапы

```bash
# Настроить cron
crontab -e

# Добавить:
0 2 * * * cd ~/ft && make db-backup
```

---

## 📊 Мониторинг

### Логи

```bash
# Все сервисы
docker-compose logs -f

# Конкретный сервис
docker-compose logs -f auth
docker-compose logs -f nginx
```

### Статус

```bash
# Docker
docker-compose ps

# Ресурсы
docker stats

# Диск
df -h
```

---

## 🔄 Обновление

После push в main, CI/CD автоматически:
1. Запустит тесты
2. Соберёт новые образы
3. Задеплоит на сервер

**Вручную:**
```bash
cd ~/ft
git pull origin main
docker-compose down
docker-compose up --build -d
```

---

## 🐛 Troubleshooting

### SSL сертификаты не получаются

```bash
# Проверить что домены указывают на сервер
nslookup quotopia.com

# Проверить что nginx запущен
docker-compose ps nginx

# Проверить логи certbot
docker-compose logs certbot
```

### Сервисы не запускаются

```bash
# Логи
docker-compose logs

# Перезапуск
docker-compose restart

# Полная пересборка
docker-compose down -v
docker-compose up --build -d
```

### 502 Bad Gateway

Сервис не запущен или не отвечает:
```bash
docker-compose ps  # Проверить статус
docker-compose logs ft  # Посмотреть логи
```

---

## ✅ Checklist

- [ ] DNS настроены
- [ ] Docker установлен
- [ ] Проект склонирован
- [ ] .env создан и заполнен
- [ ] Сильные пароли установлены
- [ ] Сервисы запущены
- [ ] SSL сертификаты получены
- [ ] Домены работают
- [ ] Admin создан
- [ ] Firewall настроен
- [ ] Бэкапы настроены
- [ ] CI/CD работает

---

**Production готов! 🎉**
