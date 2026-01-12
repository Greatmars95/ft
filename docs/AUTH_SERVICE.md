# 🔐 Auth Service - Документация

## Обзор

Auth Service предоставляет JWT-based аутентификацию для Quotopia.

**Порт:** 8090  
**База:** PostgreSQL (таблица `users`)  
**Технологии:** Go + Gin + JWT + bcrypt

---

## 🚀 Endpoints

### Публичные (без токена)

#### POST `/auth/register`
Регистрация нового пользователя.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "role": "user"  // optional: "user", "trader", "admin"
}
```

**Response (201):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-01-13T00:00:00Z",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "role": "user",
    "is_active": true,
    "created_at": "2026-01-12T22:00:00Z"
  }
}
```

#### POST `/auth/login`
Вход пользователя.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-01-13T00:00:00Z",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "role": "user",
    "is_active": true,
    "created_at": "2026-01-12T22:00:00Z",
    "last_login": "2026-01-12T23:00:00Z"
  }
}
```

---

### Защищённые (требуется токен)

**Authorization Header:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### GET `/auth/me`
Получить текущего пользователя.

**Response (200):**
```json
{
  "id": 1,
  "email": "user@example.com",
  "role": "user",
  "is_active": true,
  "created_at": "2026-01-12T22:00:00Z",
  "last_login": "2026-01-12T23:00:00Z"
}
```

#### POST `/auth/logout`
Выйти (инвалидировать токен).

**Response (200):**
```json
{
  "message": "Logged out successfully"
}
```

---

### Admin endpoints (требуется роль admin)

#### GET `/auth/admin/users`
Список всех пользователей.

**Response (200):**
```json
[
  {
    "id": 1,
    "email": "admin@quotopia.com",
    "role": "admin",
    "is_active": true,
    "created_at": "2026-01-12T22:00:00Z"
  },
  {
    "id": 2,
    "email": "user@quotopia.com",
    "role": "user",
    "is_active": true,
    "created_at": "2026-01-12T22:05:00Z"
  }
]
```

---

## 🔑 JWT Token

### Структура

```json
{
  "user_id": 1,
  "email": "user@example.com",
  "role": "user",
  "exp": 1736812800,
  "iat": 1736809200,
  "iss": "quotopia-auth"
}
```

### Время жизни

- **Access Token:** 1 час
- **Refresh Token:** 7 дней

---

## 👥 Роли

| Роль | Описание | Права |
|------|----------|-------|
| **admin** | Администратор | Полный доступ ко всему |
| **trader** | Трейдер | Торговля + просмотр |
| **user** | Пользователь | Просмотр котировок |
| **viewer** | Наблюдатель | Только чтение |

---

## 🧪 Примеры использования

### cURL

**Регистрация:**
```bash
curl -X POST https://auth.quotopia.com/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "securepassword123"
  }'
```

**Логин:**
```bash
curl -X POST https://auth.quotopia.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "securepassword123"
  }'
```

**Получить текущего пользователя:**
```bash
TOKEN="your-jwt-token-here"

curl https://auth.quotopia.com/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

### JavaScript (Fetch API)

```javascript
// Регистрация
const register = async (email, password) => {
  const response = await fetch('https://auth.quotopia.com/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  return response.json();
};

// Логин
const login = async (email, password) => {
  const response = await fetch('https://auth.quotopia.com/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  const data = await response.json();
  
  // Сохранить токен
  localStorage.setItem('token', data.token);
  return data;
};

// Запрос с авторизацией
const getMe = async () => {
  const token = localStorage.getItem('token');
  const response = await fetch('https://auth.quotopia.com/auth/me', {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  return response.json();
};
```

### React Hook

```jsx
import { useState, useEffect } from 'react';

export const useAuth = () => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      setLoading(false);
      return;
    }

    fetch('https://auth.quotopia.com/auth/me', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
      .then(res => res.json())
      .then(data => setUser(data))
      .catch(() => localStorage.removeItem('token'))
      .finally(() => setLoading(false));
  }, []);

  const login = async (email, password) => {
    const res = await fetch('https://auth.quotopia.com/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });
    const data = await res.json();
    localStorage.setItem('token', data.token);
    setUser(data.user);
    return data;
  };

  const logout = () => {
    localStorage.removeItem('token');
    setUser(null);
  };

  return { user, loading, login, logout };
};
```

---

## 🔒 Безопасность

### Хеширование паролей

Используется bcrypt с cost=10:
```go
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

### JWT Secret

**Development:**
```env
JWT_SECRET=your-super-secret-jwt-key-change-in-production
```

**Production (генерация):**
```bash
openssl rand -hex 64
```

### HTTPS Only

В production Auth Service доступен **только через HTTPS**.

### Rate Limiting

Nginx ограничивает запросы к Auth Service:
- 5 запросов/секунду
- Burst: 10 запросов

---

## 🧪 Тестирование

```bash
cd auth-service
go test -v
```

---

## 📊 Мониторинг

### Health Check

```bash
curl https://auth.quotopia.com/health
```

**Response:**
```json
{
  "status": "ok",
  "service": "auth"
}
```

---

## 🐛 Troubleshooting

### "Invalid token"

- Проверьте что токен не истёк
- Проверьте формат: `Bearer <token>`

### "Email already registered"

Пользователь с таким email уже существует.

### "Invalid credentials"

Неправильный email или пароль.

### "Admin access required"

Требуется роль `admin`.

---

## 📚 Дополнительно

- [JWT.io](https://jwt.io/) - дебаггер JWT токенов
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) - документация
