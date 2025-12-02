#!/bin/sh

echo "=== GoRestream Iniciando ==="

# Testa configuração do Nginx
nginx -t

# Inicia o Nginx em background
nginx
echo "Nginx iniciado na porta 80"

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
