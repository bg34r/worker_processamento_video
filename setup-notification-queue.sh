#!/bin/bash

echo "� Configurando infraestrutura AWS no LocalStack..."

# Configurações
AWS_CMD="docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 video-upload-service-hackaton-localstack-1"
ENDPOINT="--endpoint-url=http://localhost:4566"

# Verificar se LocalStack está rodando
echo "🔍 Verificando status do LocalStack..."
if ! curl -s http://localhost:4566/_localstack/health > /dev/null 2>&1; then
    echo "❌ LocalStack não encontrado. Certifique-se de que está rodando na porta 4566"
    exit 1
fi

echo "✅ LocalStack encontrado!"

# ========================================
# 1. VERIFICAR E CRIAR FILAS SQS
# ========================================
echo ""
echo "📬 Configurando filas SQS..."

# Fila de processamento de vídeo
echo "🎬 Verificando fila de processamento de vídeo..."
QUEUE_EXISTS=$(${AWS_CMD} aws ${ENDPOINT} sqs list-queues --query 'QueueUrls[?contains(@, `video-processing-queue`)]' --output text 2>/dev/null)

if [ -z "$QUEUE_EXISTS" ]; then
    echo "📥 Criando fila video-processing-queue..."
    ${AWS_CMD} aws ${ENDPOINT} sqs create-queue --queue-name video-processing-queue
    echo "✅ Fila video-processing-queue criada!"
else
    echo "✅ Fila video-processing-queue já existe!"
fi

# Fila de notificação
echo "🔔 Verificando fila de notificação..."
NOTIFICATION_QUEUE_EXISTS=$(${AWS_CMD} aws ${ENDPOINT} sqs list-queues --query 'QueueUrls[?contains(@, `notification-queue`)]' --output text 2>/dev/null)

if [ -z "$NOTIFICATION_QUEUE_EXISTS" ]; then
    echo "📥 Criando fila notification-queue..."
    ${AWS_CMD} aws ${ENDPOINT} sqs create-queue --queue-name notification-queue
    echo "✅ Fila notification-queue criada!"
else
    echo "✅ Fila notification-queue já existe!"
fi

# ========================================
# 2. VERIFICAR E CRIAR BUCKET S3
# ========================================
echo ""
echo "🪣 Configurando bucket S3..."

echo "📦 Verificando bucket video-service-bucket..."
BUCKET_EXISTS=$(${AWS_CMD} aws ${ENDPOINT} s3api head-bucket --bucket video-service-bucket 2>/dev/null && echo "true" || echo "false")

if [ "$BUCKET_EXISTS" = "false" ]; then
    echo "🪣 Criando bucket video-service-bucket..."
    ${AWS_CMD} aws ${ENDPOINT} s3 mb s3://video-service-bucket --region us-east-1
    echo "✅ Bucket video-service-bucket criado!"
else
    echo "✅ Bucket video-service-bucket já existe!"
fi

# Verificar estrutura de pastas no bucket
echo "� Verificando estrutura de pastas no bucket..."
${AWS_CMD} aws ${ENDPOINT} s3api put-object --bucket video-service-bucket --key processed/ --content-length 0 2>/dev/null || echo "Pasta processed já existe"
echo "✅ Estrutura do bucket configurada!"

# ========================================
# 3. VERIFICAR E CRIAR TABELA DYNAMODB
# ========================================
echo ""
echo "🗄️  Configurando tabela DynamoDB..."

echo "📊 Verificando tabela video-processing..."
TABLE_EXISTS=$(${AWS_CMD} aws ${ENDPOINT} dynamodb describe-table --table-name video-processing 2>/dev/null && echo "true" || echo "false")

if [ "$TABLE_EXISTS" = "false" ]; then
    echo "🏗️  Criando tabela video-processing..."
    ${AWS_CMD} aws ${ENDPOINT} dynamodb create-table \
        --table-name video-processing \
        --attribute-definitions \
            AttributeName=video_id,AttributeType=S \
            AttributeName=created_at,AttributeType=S \
        --key-schema \
            AttributeName=video_id,KeyType=HASH \
            AttributeName=created_at,KeyType=RANGE \
        --provisioned-throughput \
            ReadCapacityUnits=5,WriteCapacityUnits=5 \
        --region us-east-1
    
    # Aguardar a tabela ficar ativa
    echo "⏳ Aguardando tabela ficar ativa..."
    sleep 3
    echo "✅ Tabela video-processing criada!"
else
    echo "✅ Tabela video-processing já existe!"
fi

# ========================================
# 4. VERIFICAR CONFIGURAÇÕES
# ========================================
echo ""
echo "🔍 Verificando configurações finais..."

echo "📋 Filas SQS disponíveis:"
${AWS_CMD} aws ${ENDPOINT} sqs list-queues --output table

echo ""
echo "🪣 Buckets S3 disponíveis:"
${AWS_CMD} aws ${ENDPOINT} s3 ls

echo ""
echo "📁 Conteúdo do bucket video-service-bucket:"
${AWS_CMD} aws ${ENDPOINT} s3 ls s3://video-service-bucket/ --recursive

echo ""
echo "🗄️  Tabelas DynamoDB disponíveis:"
${AWS_CMD} aws ${ENDPOINT} dynamodb list-tables --output table

echo ""
echo "📊 Descrição da tabela video-processing:"
${AWS_CMD} aws ${ENDPOINT} dynamodb describe-table --table-name video-processing --query 'Table.[TableName,TableStatus,ItemCount]' --output table

echo ""
echo "🎉 Configuração completa da infraestrutura AWS concluída!"
echo "════════════════════════════════════════════════════════"
echo "✅ SQS: video-processing-queue, notification-queue"
echo "✅ S3: video-service-bucket (com pasta processed/)"
echo "✅ DynamoDB: video-processing (video_id, created_at)"
echo "════════════════════════════════════════════════════════"