#!/bin/bash

echo "🔔 Configurando fila de notificação no LocalStack..."

# Verificar se LocalStack está rodando
if ! curl -s http://localhost:4566/_localstack/health > /dev/null 2>&1; then
    echo "❌ LocalStack não encontrado. Certifique-se de que está rodando na porta 4566"
    exit 1
fi

echo "✅ LocalStack encontrado!"

# Criar fila de notificação se não existir
echo "📬 Criando fila de notificação..."
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 \
    video-upload-service-hackaton-localstack-1 \
    aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name notification-queue || echo "Fila já existe ou erro na criação"

# Listar filas para confirmar
echo "📋 Filas disponíveis:"
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 \
    video-upload-service-hackaton-localstack-1 \
    aws --endpoint-url=http://localhost:4566 sqs list-queues

echo "🎉 Configuração concluída!"