# 🚀 Начало работы с Quotopia

## Первый запуск

### 1. Запустить все сервисы

```bash
make up
```

Это запустит:
- ✅ PostgreSQL (база данных)
- ✅ Adminer (веб-интерфейс БД)
- ✅ FT (генератор котировок)
- ✅ HT (HTTP gateway)
- ✅ UI (React интерфейс)

### 2. Проверить что всё запустилось

```bash
make ps
```

Должны быть запущены 5 контейнеров:
- `quotopia-postgres`
- `quotopia-adminer`
- `quotopia-ft-1`
- `quotopia-ht-1`
- `quotopia-ui-1`

### 3. Проверить БД

```bash
./scripts/test-db.sh
```

Или открыть Adminer в браузере:
```
http://localhost:8081
```

Логин:
- **Server:** postgres
- **Username:** admin
- **Password:** secret123
- **Database:** quotopia

### 4. Открыть UI

```
http://localhost:3001
```

Вы должны увидеть котировки в реальном времени!

## 🗄️ Работа с базой данных

### Просмотр через Adminer (веб)

1. Откройте http://localhost:8081
2. Войдите (credentials выше)
3. Выберите таблицу слева
4. Нажмите "Select data"

### Просмотр через командную строку

```bash
# Подключиться к psql
make db-shell

# Затем выполняйте SQL:
SELECT * FROM users;
SELECT * FROM instruments;
SELECT * FROM active_instruments;
```

### Добавить новый инструмент

**Через Adminer:**
1. Откройте таблицу `instruments`
2. Нажмите "New item"
3. Заполните:
   - symbol: `MSFT`
   - name: `Microsoft`
   - initial_price: `420.50`
   - volatility: `0.2`
   - is_active: `true`
4. Сохраните

**Через SQL:**
```bash
make db-shell
```

```sql
INSERT INTO instruments (symbol, name, initial_price, volatility, created_by) 
VALUES ('MSFT', 'Microsoft', 420.50, 0.2, 1);
```

**Через Go (в будущем):**
```bash
curl -X POST http://localhost:9090/api/instruments \
  -H "Authorization: Bearer YOUR_JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "MSFT",
    "name": "Microsoft", 
    "initial_price": 420.50,
    "volatility": 0.2
  }'
```

### Просмотреть все инструменты

```sql
SELECT * FROM instruments;
```

Или через view:
```sql
SELECT * FROM active_instruments;
```

### Изменить цену инструмента

```sql
UPDATE instruments 
SET initial_price = 96000.00 
WHERE symbol = 'BTC';
```

### Деактивировать инструмент

```sql
UPDATE instruments 
SET is_active = false 
WHERE symbol = 'SBER';
```

### История изменений

```sql
SELECT * FROM instruments_audit_view 
ORDER BY created_at DESC 
LIMIT 10;
```

## 🔐 Пользователи и роли

### Начальные пользователи

| Email | Пароль | Роль |
|-------|--------|------|
| admin@quotopia.com | admin123 | admin |
| trader@quotopia.com | admin123 | trader |
| user@quotopia.com | admin123 | user |

### Создать нового пользователя

```sql
INSERT INTO users (email, password_hash, role) 
VALUES (
  'newuser@example.com', 
  '$2a$10$X6xYQqZ9p5F6M3qKvN5Vw.ZYz3YqZ9p5F6M3qKvN5Vw.ZYz3YqZ9p',  -- admin123
  'trader'
);
```

⚠️ **Примечание:** Это placeholder хеш для admin123. В production используйте реальный bcrypt хеш!

### Изменить роль пользователя

```sql
UPDATE users 
SET role = 'admin' 
WHERE email = 'user@example.com';
```

## 📊 Полезные запросы

### Статистика

```sql
SELECT * FROM instruments_stats;
```

### Активные инструменты с создателями

```sql
SELECT 
    i.symbol,
    i.name,
    i.initial_price,
    u.email as created_by
FROM instruments i
LEFT JOIN users u ON i.created_by = u.id
WHERE i.is_active = true;
```

### Топ волатильных инструментов

```sql
SELECT symbol, name, volatility 
FROM instruments 
WHERE is_active = true 
ORDER BY volatility DESC 
LIMIT 5;
```

### Недавно добавленные инструменты

```sql
SELECT symbol, name, created_at 
FROM instruments 
ORDER BY created_at DESC 
LIMIT 10;
```

## 🛠️ Полезные команды

```bash
# Запустить всё
make up

# Остановить всё
make down

# Логи всех сервисов
make logs

# Логи только PostgreSQL
docker-compose logs -f postgres

# Перезапустить сервисы
make restart

# Подключиться к БД
make db-shell

# Открыть Adminer (показать credentials)
make adminer

# Сделать бэкап БД
make db-backup

# Полностью сбросить БД
make db-reset

# Проверить статус
make ps
```

## 🐛 Troubleshooting

### БД не запускается

```bash
# Смотрим логи
docker-compose logs postgres

# Если нужно пересоздать
docker-compose down -v
docker-compose up -d postgres
```

### Adminer не открывается

Проверьте что контейнер запущен:
```bash
docker ps | grep adminer
```

Если не запущен:
```bash
docker-compose up -d adminer
```

### FT не видит инструменты из БД

Сейчас FT ещё читает hardcoded данные. Нужно обновить код (сделаем далее).

### Забыли пароль от БД

Смотрите в `docker-compose.yml`:
```yaml
POSTGRES_PASSWORD: secret123
```

## 📚 Следующие шаги

1. ✅ БД настроена
2. ⏳ Создать Auth Service (JWT авторизация)
3. ⏳ Создать Admin Service (CRUD инструментов)
4. ⏳ Обновить FT (читать из БД)
5. ⏳ Создать Admin UI (React панель)

---

**Готовы продолжить?** Следующий шаг - Auth Service! 🔐
