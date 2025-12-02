#!/bin/sh

# Define porta padrão se não estiver definida
export API_PORT=${PORT:-8080}

echo "=== GoRestream Iniciando ==="
echo "PORT env: ${PORT}"
echo "API_PORT: ${API_PORT}"

# Substitui APENAS ${API_PORT} no nginx.conf (protege $host, $scheme, etc)
envsubst '${API_PORT}' < /etc/nginx/nginx.conf > /etc/nginx/nginx.conf.tmp
mv /etc/nginx/nginx.conf.tmp /etc/nginx/nginx.conf

# Verifica se a substituição funcionou
echo "=== Nginx Config (server block) ==="
grep -A 5 "listen" /etc/nginx/nginx.conf | head -10

# Testa configuração do Nginx
nginx -t

# Inicia o Nginx em background
nginx
echo "Nginx iniciado na porta ${API_PORT}"

# Aguarda um momento para o Nginx iniciar
sleep 1

# Verifica se Nginx está rodando
if pgrep nginx > /dev/null; then
    echo "Nginx rodando OK"
else
    echo "ERRO: Nginx não iniciou!"
    exit 1
fi

# Inicia o Go App (porta interna 3000)
echo "Iniciando Go server na porta 3000..."
exec /app/server
