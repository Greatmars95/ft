#!/bin/bash
echo "🔌 Проверка подключения к серверу..."
ssh -o ConnectTimeout=5 -o StrictHostKeyChecking=no root@45.132.255.115 "echo '✅ SSH работает!' && docker ps --format 'table {{.Names}}\t{{.Status}}'"
