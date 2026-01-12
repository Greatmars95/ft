# 🗄️ База данных Quotopia

## Обзор

Quotopia использует **PostgreSQL 16** как основную базу данных для хранения:
- Пользователей и их ролей
- Торговых инструментов
- JWT токенов
- Истории изменений (audit log)

## 🚀 Быстрый старт

### Запуск БД

```bash
# Запустить все сервисы (включая PostgreSQL)
make up

# Или только PostgreSQL
docker-compose up -d postgres
```

## 🔌 Как подключиться к БД

### 1️⃣ Через Adminer (веб-интерфейс) - САМЫЙ ПРОСТОЙ

```bash
# Откройте в браузере
http://localhost:8081

# Логин:
Server:   postgres
Username: admin
Password: secret123
Database: quotopia
```

**Что можно делать:**
- ✅ Просматривать таблицы
- ✅ Добавлять/редактировать записи
- ✅ Выполнять SQL запросы
- ✅ Экспортировать данные
- ✅ Смотреть структуру таблиц

### 2️⃣ Через Docker CLI

```bash
# Подключиться к psql внутри контейнера
docker exec -it quotopia-postgres psql -U admin -d quotopia

# Или через make команду
make db-shell
```

### 3️⃣ Через psql (если установлен локально)

```bash
psql -h localhost -p 5432 -U admin -d quotopia
# Пароль: secret123
```

### 4️⃣ Через GUI клиенты

**DBeaver / DataGrip / pgAdmin:**
- Host: `localhost`
- Port: `5432`
- Database: `quotopia`
- Username: `admin`
- Password: `secret123`

### 5️⃣ Из Go кода

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

connStr := "host=localhost port=5432 user=admin password=secret123 dbname=quotopia sslmode=disable"
db, err := sql.Open("postgres", connStr)
```

## 📊 Структура БД

### Таблицы

| Таблица | Назначение |
|---------|-----------|
| `users` | Пользователи системы |
| `instruments` | Торговые инструменты |
| `refresh_tokens` | JWT refresh токены |
| `instruments_audit` | История изменений |

### Схема

```sql
-- Пользователи
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(50) DEFAULT 'user',
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW(),
  last_login TIMESTAMP
);

-- Инструменты
CREATE TABLE instruments (
  id SERIAL PRIMARY KEY,
  symbol VARCHAR(10) UNIQUE NOT NULL,
  name VARCHAR(100) NOT NULL,
  initial_price DECIMAL(18, 8) NOT NULL,
  volatility DECIMAL(5, 2) DEFAULT 0.1,
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  created_by INTEGER REFERENCES users(id)
);
```

### Представления (Views)

```sql
-- Активные инструменты
SELECT * FROM active_instruments;

-- Статистика
SELECT * FROM instruments_stats;

-- История изменений
SELECT * FROM instruments_audit_view;
```

## 🔧 Полезные команды

### Просмотр данных

```sql
-- Все пользователи
SELECT * FROM users;

-- Все инструменты
SELECT * FROM instruments;

-- Только активные инструменты
SELECT * FROM active_instruments;

-- Статистика
SELECT * FROM instruments_stats;
```

### Добавление данных

```sql
-- Добавить нового пользователя
INSERT INTO users (email, password_hash, role) 
VALUES ('newuser@example.com', '$2a$10$...', 'trader');

-- Добавить новый инструмент
INSERT INTO instruments (symbol, name, initial_price, volatility, created_by) 
VALUES ('TSLA', 'Tesla Inc.', 245.50, 0.4, 1);
```

### Обновление данных

```sql
-- Изменить цену инструмента
UPDATE instruments 
SET initial_price = 250.00 
WHERE symbol = 'TSLA';

-- Деактивировать инструмент
UPDATE instruments 
SET is_active = false 
WHERE symbol = 'SBER';
```

### Удаление данных

```sql
-- Удалить инструмент
DELETE FROM instruments WHERE symbol = 'AAPL';

-- Очистить таблицу (осторожно!)
TRUNCATE TABLE instruments_audit;
```

## 🛠️ Make команды для БД

```bash
# Открыть psql shell
make db-shell

# Бэкап БД
make db-backup

# Восстановить из бэкапа
make db-restore FILE=backup.sql

# Полная очистка и пересоздание
make db-reset

# Показать статус контейнеров
make ps

# Открыть Adminer (показать credentials)
make adminer
```

## 📝 Начальные данные

### Пользователи

| Email | Пароль | Роль |
|-------|--------|------|
| admin@quotopia.com | admin123 | admin |
| trader@quotopia.com | admin123 | trader |
| user@quotopia.com | admin123 | user |

### Инструменты

| Symbol | Name | Price | Volatility |
|--------|------|-------|-----------|
| BTC | Bitcoin | 95,400.00 | 0.5% |
| ETH | Ethereum | 2,650.20 | 0.3% |
| SBER | Сбербанк | 275.50 | 0.1% |
| AAPL | Apple Inc. | 185.50 | 0.2% |
| GOOGL | Google | 142.30 | 0.2% |

## 🔒 Роли и права

### admin
- ✅ Полный доступ ко всему
- ✅ CRUD инструментов
- ✅ Управление пользователями
- ✅ Просмотр audit logs

### trader
- ✅ Просмотр инструментов
- ✅ Создание ордеров (в будущем)
- ❌ Изменение инструментов

### user
- ✅ Просмотр котировок
- ✅ Основные операции
- ❌ Торговля

### viewer
- ✅ Только чтение
- ❌ Любые изменения

## 🔍 Audit Log

Все изменения в таблице `instruments` автоматически логируются:

```sql
-- Посмотреть историю изменений
SELECT * FROM instruments_audit_view 
ORDER BY created_at DESC 
LIMIT 10;

-- История конкретного инструмента
SELECT * FROM instruments_audit_view 
WHERE symbol = 'BTC' 
ORDER BY created_at DESC;
```

## 🐛 Troubleshooting

### БД не запускается

```bash
# Проверить логи
docker-compose logs postgres

# Пересоздать контейнер
docker-compose down -v
docker-compose up -d postgres
```

### Забыли пароль

Пароль находится в:
- `docker-compose.yml` (переменная `POSTGRES_PASSWORD`)
- `.env.example`
- По умолчанию: `secret123`

### Нужно сбросить БД

```bash
# ОСТОРОЖНО: Удалит все данные!
make db-reset
```

### Подключение отклонено

```bash
# Проверить, что контейнер запущен
docker ps | grep postgres

# Проверить порт
netstat -an | grep 5432
```

## 📚 Полезные ресурсы

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Adminer Documentation](https://www.adminer.org/)
- [Go PostgreSQL Driver](https://github.com/lib/pq)

## 🔐 Security Tips

### Production

1. **Смените пароли!**
   ```yaml
   POSTGRES_PASSWORD: your-strong-password
   ```

2. **Не открывайте порт 5432 наружу**
   ```yaml
   # Для production убрать:
   ports:
     - "5432:5432"
   ```

3. **Используйте SSL**
   ```go
   connStr := "...sslmode=require"
   ```

4. **Регулярные бэкапы**
   ```bash
   # Настроить cron job
   0 2 * * * make db-backup
   ```

## 📊 Мониторинг

### Размер БД

```sql
SELECT 
    pg_size_pretty(pg_database_size('quotopia')) as db_size;
```

### Количество подключений

```sql
SELECT count(*) FROM pg_stat_activity 
WHERE datname = 'quotopia';
```

### Активные запросы

```sql
SELECT pid, usename, state, query 
FROM pg_stat_activity 
WHERE state = 'active';
```

---

**Вопросы?** Проверьте Adminer на http://localhost:8081 👍
