# 🌐 Настройка Nginx + SSL для Quotopia

## Цель

Настроить:
- ✅ Домены вместо IP:порт
- ✅ HTTPS (SSL/TLS через Let's Encrypt)
- ✅ Reverse proxy для всех сервисов
- ✅ Basic Auth для Adminer

---

## 📋 Требования

- Домен (например, `quotopia.com`)
- Доступ к DNS настройкам
- Сервер с публичным IP

---

## 🚀 Шаг 1: DNS настройки

Добавьте A-записи в вашем DNS провайдере:

```
quotopia.com              A    your-server-ip
api.quotopia.com          A    your-server-ip
admin.quotopia.com        A    your-server-ip
```

**Проверка:**
```bash
nslookup quotopia.com
# Должен вернуть ваш IP
```

---

## 🐳 Шаг 2: Добавить Nginx в Docker Compose

Создайте файл `nginx/nginx.conf`:

```nginx
# Основной конфиг будет генерироваться автоматически
# через certbot и docker-compose
```

Обновите `docker-compose.yml`:

```yaml
services:
  # ... существующие сервисы ...

  nginx:
    image: nginx:alpine
    container_name: quotopia-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/conf.d:/etc/nginx/conf.d
      - ./nginx/ssl:/etc/nginx/ssl
      - certbot_www:/var/www/certbot
      - certbot_conf:/etc/letsencrypt
    depends_on:
      - ui
      - ht
      - adminer
    networks:
      - quotopia-net
    restart: unless-stopped

  certbot:
    image: certbot/certbot
    container_name: quotopia-certbot
    volumes:
      - certbot_www:/var/www/certbot
      - certbot_conf:/etc/letsencrypt
    entrypoint: "/bin/sh -c 'trap exit TERM; while :; do certbot renew; sleep 12h & wait $${!}; done;'"
    networks:
      - quotopia-net

volumes:
  certbot_www:
  certbot_conf:
  # ... остальные volumes ...
```

---

## 🔐 Шаг 3: Создать Nginx конфигурацию

### `nginx/conf.d/quotopia.conf`

```nginx
# UI - Основной сайт
server {
    listen 80;
    server_name quotopia.com www.quotopia.com;
    
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
    
    location / {
        return 301 https://$server_name$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name quotopia.com www.quotopia.com;
    
    ssl_certificate /etc/letsencrypt/live/quotopia.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/quotopia.com/privkey.pem;
    
    location / {
        proxy_pass http://ui:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# API
server {
    listen 80;
    server_name api.quotopia.com;
    
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
    
    location / {
        return 301 https://$server_name$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name api.quotopia.com;
    
    ssl_certificate /etc/letsencrypt/live/api.quotopia.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.quotopia.com/privkey.pem;
    
    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
    limit_req zone=api_limit burst=20 nodelay;
    
    location / {
        proxy_pass http://ht:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Adminer - с Basic Auth
server {
    listen 80;
    server_name admin.quotopia.com;
    
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
    
    location / {
        return 301 https://$server_name$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name admin.quotopia.com;
    
    ssl_certificate /etc/letsencrypt/live/admin.quotopia.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/admin.quotopia.com/privkey.pem;
    
    # Basic Auth
    auth_basic "Admin Area";
    auth_basic_user_file /etc/nginx/.htpasswd;
    
    location / {
        proxy_pass http://adminer:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## 🔑 Шаг 4: Создать Basic Auth для Adminer

```bash
# На сервере
docker run --rm httpd:alpine htpasswd -Bbn admin your_password > nginx/.htpasswd
```

---

## 📜 Шаг 5: Получить SSL сертификаты

### Первый раз (инициализация)

```bash
# 1. Запустить nginx без SSL
docker-compose up -d nginx

# 2. Получить сертификаты для каждого домена
docker-compose run --rm certbot certonly --webroot \
  --webroot-path=/var/www/certbot \
  --email your@email.com \
  --agree-tos \
  --no-eff-email \
  -d quotopia.com \
  -d www.quotopia.com

docker-compose run --rm certbot certonly --webroot \
  --webroot-path=/var/www/certbot \
  --email your@email.com \
  --agree-tos \
  --no-eff-email \
  -d api.quotopia.com

docker-compose run --rm certbot certonly --webroot \
  --webroot-path=/var/www/certbot \
  --email your@email.com \
  --agree-tos \
  --no-eff-email \
  -d admin.quotopia.com

# 3. Перезапустить nginx с SSL
docker-compose restart nginx
```

### Автоматическое обновление

Certbot контейнер автоматически обновляет сертификаты каждые 12 часов.

---

## ✅ Шаг 6: Проверка

```bash
# Проверить что всё запустилось
docker-compose ps

# Проверить SSL
curl https://quotopia.com
curl https://api.quotopia.com/quotes
curl https://admin.quotopia.com
# (попросит логин/пароль)
```

---

## 🎯 Результат

```
✅ https://quotopia.com           → UI (React)
✅ https://api.quotopia.com       → API (котировки)
✅ https://admin.quotopia.com     → Adminer (с паролем)
✅ SSL A+ рейтинг
✅ Автообновление сертификатов
✅ Rate limiting на API
```

---

## 🐛 Troubleshooting

### Ошибка: "Connection refused"

```bash
# Проверить что контейнеры запущены
docker-compose ps

# Проверить логи nginx
docker-compose logs nginx
```

### Ошибка: "SSL certificate not found"

Нужно сначала получить сертификаты (Шаг 5).

### Ошибка: "Rate limit exceeded"

Увеличьте лимиты в `nginx.conf`:
```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/s;
```

---

## 📚 Дополнительно

### Оптимизация SSL

Добавьте в `nginx.conf`:

```nginx
ssl_protocols TLSv1.2 TLSv1.3;
ssl_prefer_server_ciphers on;
ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256';
ssl_session_cache shared:SSL:10m;
ssl_session_timeout 10m;
```

### Мониторинг SSL

```bash
# Проверить рейтинг SSL
https://www.ssllabs.com/ssltest/analyze.html?d=quotopia.com
```
