-- ============================================
-- Quotopia Database Schema
-- ============================================

-- Таблица пользователей (для Auth Service)
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(50) DEFAULT 'user' CHECK (role IN ('admin', 'trader', 'user', 'viewer')),
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW(),
  last_login TIMESTAMP,
  CONSTRAINT email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- Индекс для быстрого поиска по email
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

-- Таблица инструментов (для Admin Service + FT)
CREATE TABLE IF NOT EXISTS instruments (
  id SERIAL PRIMARY KEY,
  symbol VARCHAR(10) UNIQUE NOT NULL,
  name VARCHAR(100) NOT NULL,
  initial_price DECIMAL(18, 8) NOT NULL CHECK (initial_price > 0),
  volatility DECIMAL(5, 2) DEFAULT 0.1 CHECK (volatility >= 0 AND volatility <= 100),
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  created_by INTEGER REFERENCES users(id) ON DELETE SET NULL
);

-- Индексы для быстрого доступа
CREATE INDEX idx_instruments_symbol ON instruments(symbol);
CREATE INDEX idx_instruments_is_active ON instruments(is_active);

-- Таблица refresh токенов (для Auth Service)
CREATE TABLE IF NOT EXISTS refresh_tokens (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token VARCHAR(255) UNIQUE NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  CONSTRAINT future_expiration CHECK (expires_at > created_at)
);

-- Индекс для быстрой проверки токенов
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Таблица для blacklist токенов (logout)
CREATE TABLE IF NOT EXISTS token_blacklist (
  id SERIAL PRIMARY KEY,
  token VARCHAR(1000) UNIQUE NOT NULL,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Индекс для быстрой проверки blacklist
CREATE INDEX idx_token_blacklist_token ON token_blacklist(token);
CREATE INDEX idx_token_blacklist_expires_at ON token_blacklist(expires_at);

-- Таблица истории изменений инструментов (audit log)
CREATE TABLE IF NOT EXISTS instruments_audit (
  id SERIAL PRIMARY KEY,
  instrument_id INTEGER REFERENCES instruments(id) ON DELETE CASCADE,
  action VARCHAR(50) NOT NULL CHECK (action IN ('created', 'updated', 'deleted', 'activated', 'deactivated')),
  changed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
  old_data JSONB,
  new_data JSONB,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_instruments_audit_instrument_id ON instruments_audit(instrument_id);
CREATE INDEX idx_instruments_audit_created_at ON instruments_audit(created_at);

-- ============================================
-- Начальные данные
-- ============================================

-- Создаём первого администратора
-- Email: admin@quotopia.com
-- Пароль: admin123 (хеш bcrypt)
INSERT INTO users (email, password_hash, role) VALUES
  ('admin@quotopia.com', '$2a$10$X6xYQqZ9p5F6M3qKvN5Vw.ZYz3YqZ9p5F6M3qKvN5Vw.ZYz3YqZ9p', 'admin'),
  ('trader@quotopia.com', '$2a$10$X6xYQqZ9p5F6M3qKvN5Vw.ZYz3YqZ9p5F6M3qKvN5Vw.ZYz3YqZ9p', 'trader'),
  ('user@quotopia.com', '$2a$10$X6xYQqZ9p5F6M3qKvN5Vw.ZYz3YqZ9p5F6M3qKvN5Vw.ZYz3YqZ9p', 'user')
ON CONFLICT (email) DO NOTHING;

-- Добавляем начальные инструменты
INSERT INTO instruments (symbol, name, initial_price, volatility, created_by) VALUES
  ('BTC', 'Bitcoin', 95400.00, 0.5, 1),
  ('ETH', 'Ethereum', 2650.20, 0.3, 1),
  ('SBER', 'Сбербанк', 275.50, 0.1, 1),
  ('AAPL', 'Apple Inc.', 185.50, 0.2, 1),
  ('GOOGL', 'Google', 142.30, 0.2, 1)
ON CONFLICT (symbol) DO NOTHING;

-- ============================================
-- Функции и триггеры
-- ============================================

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер для автоматического обновления updated_at в instruments
CREATE TRIGGER update_instruments_updated_at
BEFORE UPDATE ON instruments
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Функция для логирования изменений инструментов
CREATE OR REPLACE FUNCTION log_instrument_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO instruments_audit (instrument_id, action, new_data)
        VALUES (NEW.id, 'created', row_to_json(NEW)::jsonb);
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO instruments_audit (instrument_id, action, old_data, new_data)
        VALUES (NEW.id, 'updated', row_to_json(OLD)::jsonb, row_to_json(NEW)::jsonb);
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO instruments_audit (instrument_id, action, old_data)
        VALUES (OLD.id, 'deleted', row_to_json(OLD)::jsonb);
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Триггер для логирования изменений
CREATE TRIGGER log_instruments_changes
AFTER INSERT OR UPDATE OR DELETE ON instruments
FOR EACH ROW
EXECUTE FUNCTION log_instrument_changes();

-- ============================================
-- Полезные представления (views)
-- ============================================

-- Активные инструменты с информацией о создателе
CREATE OR REPLACE VIEW active_instruments AS
SELECT 
    i.id,
    i.symbol,
    i.name,
    i.initial_price,
    i.volatility,
    i.created_at,
    i.updated_at,
    u.email as created_by_email
FROM instruments i
LEFT JOIN users u ON i.created_by = u.id
WHERE i.is_active = true
ORDER BY i.symbol;

-- Статистика по инструментам
CREATE OR REPLACE VIEW instruments_stats AS
SELECT 
    COUNT(*) as total_instruments,
    COUNT(*) FILTER (WHERE is_active = true) as active_instruments,
    COUNT(*) FILTER (WHERE is_active = false) as inactive_instruments,
    AVG(initial_price) as avg_price,
    MAX(initial_price) as max_price,
    MIN(initial_price) as min_price
FROM instruments;

-- История изменений с информацией о пользователе
CREATE OR REPLACE VIEW instruments_audit_view AS
SELECT 
    a.id,
    a.instrument_id,
    i.symbol,
    a.action,
    a.old_data,
    a.new_data,
    u.email as changed_by_email,
    a.created_at
FROM instruments_audit a
LEFT JOIN instruments i ON a.instrument_id = i.id
LEFT JOIN users u ON a.changed_by = u.id
ORDER BY a.created_at DESC;

-- ============================================
-- Комментарии к таблицам
-- ============================================

COMMENT ON TABLE users IS 'Пользователи системы';
COMMENT ON TABLE instruments IS 'Торговые инструменты (акции, криптовалюты)';
COMMENT ON TABLE refresh_tokens IS 'Refresh токены для JWT авторизации';
COMMENT ON TABLE instruments_audit IS 'История изменений инструментов';

COMMENT ON COLUMN users.role IS 'Роль: admin (полный доступ), trader (торговля), user (просмотр), viewer (только чтение)';
COMMENT ON COLUMN instruments.volatility IS 'Волатильность в процентах (например, 0.1 = ±0.1% изменение)';

-- ============================================
-- Готово!
-- ============================================

-- Выводим статистику
DO $$
BEGIN
    RAISE NOTICE '✅ База данных Quotopia успешно инициализирована!';
    RAISE NOTICE '📊 Создано пользователей: %', (SELECT COUNT(*) FROM users);
    RAISE NOTICE '📈 Создано инструментов: %', (SELECT COUNT(*) FROM instruments);
    RAISE NOTICE '🔐 Логины: admin@quotopia.com / trader@quotopia.com / user@quotopia.com';
    RAISE NOTICE '🔑 Пароль для всех: admin123 (измените в production!)';
END $$;
