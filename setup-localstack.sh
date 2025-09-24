#!/bin/bash

echo "Configurando LocalStack..."

# Parar containers existentes e limpar
echo "Limpando containers anteriores..."
docker-compose down --volumes --remove-orphans

# Limpar dados antigos
docker volume prune -f

# Iniciar LocalStack
echo "Iniciando LocalStack..."
docker-compose up -d

# Verificar se o container está rodando
CONTAINER_NAME="localstack-main"

# Aguardar LocalStack inicializar
echo "Aguardando LocalStack inicializar..."
sleep 20

# Verificar conectividade
echo "Testando conectividade..."
for i in {1..30}; do
    if curl -s http://localhost:4566/_localstack/health > /dev/null 2>&1; then
        echo "LocalStack está pronto!"
        break
    fi
    echo "Tentativa $i/30 - aguardando LocalStack..."
    sleep 2
done

# Criar bucket S3
echo "Criando bucket S3..."
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 $CONTAINER_NAME aws --endpoint-url=http://localhost:4566 s3 mb s3://video-bucket

# Criar fila SQS
echo "Criando fila SQS..."
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 $CONTAINER_NAME aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name video-processing-queue

# Listar recursos criados
echo "Recursos criados:"
echo "Bucket S3:"
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 $CONTAINER_NAME aws --endpoint-url=http://localhost:4566 s3 ls

echo "Filas SQS:"
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 $CONTAINER_NAME aws --endpoint-url=http://localhost:4566 sqs list-queues

echo "Setup concluído com sucesso!"