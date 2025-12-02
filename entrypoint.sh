#!/bin/sh

# Define porta padrão se não estiver definida
export API_PORT=${PORT:-8080}

# Substitui variáveis de ambiente no nginx.conf
envsubst '${API_PORT}' < /etc/nginx/nginx.conf > /etc/nginx/nginx.conf.tmp
mv /etc/nginx/nginx.conf.tmp /etc/nginx/nginx.conf

# Inicia o Nginx em background
nginx

# Inicia o Go App
/app/server
