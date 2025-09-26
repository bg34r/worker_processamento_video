#!/bin/bash

set -e

echo "🚀 Iniciando Worker de Processamento de Vídeos (Standalone)..."

# Função para limpeza no shutdown
cleanup() {
    echo "📤 Recebido sinal de shutdown, finalizando serviços..."
    
    # Parar worker graciosamente
    if pgrep -f "/app/worker" > /dev/null; then
        echo "⏹️  Parando worker..."
        pkill -TERM -f "/app/worker"
    fi
    
    # Parar nginx graciosamente
    if pgrep -f "nginx" > /dev/null; then
        echo "⏹️  Parando nginx..."
        nginx -s quit
    fi
    
    # Aguardar shutdown gracioso
    sleep 3
    
    # Force kill se ainda estiverem rodando
    pkill -KILL -f "/app/worker" 2>/dev/null || true
    pkill -KILL -f "nginx" 2>/dev/null || true
    
    echo "✅ Shutdown completo"
    exit 0
}

# Função de verificação de saúde dos serviços
check_services() {
    if ! pgrep -f "/app/worker" > /dev/null; then
        echo "❌ Worker morreu, reiniciando container..."
        exit 1
    fi
    
    if ! pgrep -f "nginx" > /dev/null; then
        echo "❌ Nginx morreu, reiniciando container..."
        exit 1
    fi
}

# Configurar manipuladores de sinal
trap cleanup SIGTERM SIGINT SIGQUIT

# Criar diretórios necessários
mkdir -p /app/temp /app/outputs /app/uploads

# Determinar endpoint do LocalStack
LOCALSTACK_ENDPOINT=${LOCALSTACK_ENDPOINT:-"http://localhost:4566"}
echo "🔗 Conectando ao LocalStack em: $LOCALSTACK_ENDPOINT"

# Aguardar LocalStack estar disponível
echo "🔄 Aguardando LocalStack estar disponível..."
for i in {1..60}; do
    if curl -s "$LOCALSTACK_ENDPOINT/_localstack/health" > /dev/null 2>&1; then
        echo "✅ LocalStack disponível!"
        break
    fi
    echo "⏳ Tentativa $i/60 - aguardando LocalStack..."
    sleep 3
done

# Verificar se conseguiu conectar
if ! curl -s "$LOCALSTACK_ENDPOINT/_localstack/health" > /dev/null 2>&1; then
    echo "❌ Não foi possível conectar ao LocalStack em $LOCALSTACK_ENDPOINT"
    echo "💡 Certifique-se de que o projeto video-upload-service-hackaton está rodando"
    exit 1
fi

# Iniciar worker de processamento em background
echo "🎬 Iniciando Worker de Processamento..."
/app/worker &
WORKER_PID=$!

# Aguardar um pouco para o worker inicializar
sleep 2

# Verificar se worker iniciou corretamente
if ! kill -0 $WORKER_PID 2>/dev/null; then
    echo "❌ Falha ao iniciar worker de processamento"
    exit 1
fi

echo "✅ Worker de processamento iniciado (PID: $WORKER_PID)"

# Iniciar nginx em background
echo "🌐 Iniciando Nginx..."
nginx -g 'daemon off;' &
NGINX_PID=$!

echo "✅ Nginx iniciado (PID: $NGINX_PID)"
echo "🎉 Todos os serviços iniciados com sucesso!"
echo "📊 Worker monitora fila SQS: video-processing-queue"
echo "📁 Bucket S3: video-service-bucket"

# Loop de monitoramento
while true; do
    sleep 15
    check_services
done