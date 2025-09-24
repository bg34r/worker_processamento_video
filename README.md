# Video Processing Worker

Este projeto é um sistema de processamento de vídeos que utiliza AWS S3 e SQS via LocalStack para extrair frames de vídeos e gerar arquivos ZIP com os frames extraídos.

## 📋 Índice

- [Arquitetura](#-arquitetura)
- [Pré-requisitos](#-pré-requisitos)
- [Instalação](#-instalação)
- [Configuração](#-configuração)
- [Execução](#-execução)
- [Testando o Sistema](#-testando-o-sistema)
- [Estrutura do Projeto](#-estrutura-do-projeto)
- [API Reference](#-api-reference)

## 🏗 Arquitetura

O sistema funciona com os seguintes componentes:

1. **LocalStack**: Simula serviços AWS (S3 e SQS) localmente
2. **S3 Bucket**: Armazena vídeos de entrada e ZIPs de saída
3. **SQS Queue**: Fila de mensagens para processamento de vídeos
4. **Worker Go**: Aplicação que consome mensagens e processa vídeos
5. **FFmpeg**: Extrai frames dos vídeos

### Fluxo de Processamento

```
Vídeo → S3 → Mensagem SQS → Worker → Extração de Frames → ZIP → S3
```

## 📋 Pré-requisitos

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Go 1.21+](https://golang.org/dl/)
- [Git Bash](https://git-scm.com/downloads) (para executar scripts bash no Windows)
- [FFmpeg](https://ffmpeg.org/download.html) (instalado no sistema)

## 🚀 Instalação

### 1. Clone o repositório

```bash
git clone <url-do-repositorio>
cd worker
```

### 2. Instale as dependências Go

```bash
go mod download
```

### 3. Configure o LocalStack

Certifique-se de que o Docker Desktop está rodando, então execute:

```bash
# Dar permissão de execução ao script (se necessário)
chmod +x setup-localstack.sh

# Configurar LocalStack
bash ./setup-localstack.sh
```

Este script irá:
- Parar containers existentes
- Iniciar o LocalStack
- Criar bucket S3 (`video-bucket`)
- Criar fila SQS (`video-processing-queue`)

## ⚙️ Configuração

### Estrutura do docker-compose.yml

```yaml
services:
  localstack:
    image: localstack/localstack:2.3
    container_name: localstack-main
    ports:
      - "4566:4566"
    environment:
      - SERVICES=s3,sqs
      - DEBUG=1
      - LOCALSTACK_HOST=localhost
      - DOCKER_HOST=unix:///var/run/docker.sock
    volumes:
      - "localstack-data:/var/lib/localstack"
      - "/var/run/docker.sock:/var/run/docker.sock"

volumes:
  localstack-data:
```

### Variáveis de Ambiente

O sistema usa as seguintes credenciais para o LocalStack:
- `AWS_ACCESS_KEY_ID=test`
- `AWS_SECRET_ACCESS_KEY=test`
- `AWS_DEFAULT_REGION=us-east-1`

## 🏃‍♂️ Execução

### 1. Inicie o LocalStack

```bash
# Se não executou o setup ainda
bash ./setup-localstack.sh

# Ou apenas para iniciar
docker-compose up -d
```

### 2. Verifique se o LocalStack está funcionando

```bash
curl http://localhost:4566/_localstack/health
```

### 3. Execute o Worker

```bash
go run cmd/server/main.go
```

Você deve ver a mensagem:
```
2025/09/23 20:36:13 Worker iniciado - monitorando fila SQS...
```

## 🧪 Testando o Sistema

### Teste Completo com Vídeos Reais

#### 1. Enviar vídeos para o S3

```bash
# Copiar vídeo para o container
docker cp "C:\Users\Caminho\Para\video\nomeDoVideo0.mp4" localstack-main:/tmp/nomeDoVideo0.mp4

# Fazer upload para S3
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 cp /tmp/nomeDoVideo0.mp4 s3://video-bucket/
```

#### 2. Enviar múltiplos vídeos

```bash
# Exemplos com mais vídeos
docker cp "C:\Users\Caminho\Para\video\nomeDoVideo1.mp4" localstack-main:/tmp/nomeDoVideo1.mp4
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 cp /tmp/nomeDoVideo1.mp4 s3://video-bucket/

docker cp "C:\Users\Caminho\Para\video\nomeDoVideo2.mp4" localstack-main:/tmp/nomeDoVideo2.mp4
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 cp /tmp/nomeDoVideo2.mp4 s3://video-bucket/
```

#### 3. Verificar vídeos no bucket

```bash
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 ls s3://video-bucket/
```

Resultado esperado:
```
2025-09-23 23:40:08    2317191 nomeDoVideo0.mp4
2025-09-23 23:40:18   18542096 nomeDoVideo1.mp4
2025-09-23 23:40:30    4630949 nomeDoVideo2.mp4
```

#### 4. Enviar mensagens para processamento

```bash
# Processar nomeDoVideo0.mp4
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 sqs send-message --queue-url http://localhost:4566/000000000000/video-processing-queue --message-body '{"video_key":"nomeDoVideo0.mp4","video_id":"nomeDoVideo0_001"}'

# Processar nomeDoVideo1.mp4
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 sqs send-message --queue-url http://localhost:4566/000000000000/video-processing-queue --message-body '{"video_key":"nomeDoVideo1.mp4","video_id":"nomeDoVideo1_001"}'

# Processar nomeDoVideo2.mp4
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 sqs send-message --queue-url http://localhost:4566/000000000000/video-processing-queue --message-body '{"video_key":"nomeDoVideo2.mp4","video_id":"nomeDoVideo2_001"}'
```

#### 5. Monitorar o processamento

O worker deve mostrar mensagens como:
```
2025/09/23 20:41:05 Processando vídeo: nomeDoVideo0.mp4
2025/09/23 20:41:05 Processamento concluído: nomeDoVideo0.mp4 -> processed/nomeDoVideo0_frames.zip
```

#### 6. Verificar resultados

```bash
# Ver todos os arquivos processados
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 ls s3://video-bucket/ --recursive
```

Resultado esperado:
```
2025-09-23 23:40:08    2317191 nomeDoVideo0.mp4
2025-09-23 23:40:18   18542096 nomeDoVideo1.mp4
2025-09-23 23:41:05   11899270 processed/nomeDoVideo0_frames.zip
2025-09-23 23:41:20   48835647 processed/nomeDoVideo1_frames.zip
2025-09-23 23:41:28   11288170 processed/nomeDoVideo2_frames.zip
2025-09-23 23:40:30    4630949 nomeDoVideo2.mp4
```

#### 7. Baixar arquivo processado

```bash
# Baixar ZIP de frames
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 cp s3://video-bucket/processed/nomeDoVideo0_frames.zip /tmp/bamboo_frames.zip

# Copiar para o host
docker cp localstack-main:/tmp/nomeDoVideo0_frames.zip ./nomeDoVideo0_frames.zip
```

#### 8. Verificar fila vazia

```bash
# Verificar se não há mais mensagens na fila
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 sqs receive-message --queue-url http://localhost:4566/000000000000/video-processing-queue
```

Se não retornar nenhuma mensagem, significa que todas foram processadas com sucesso.

## 📁 Estrutura do Projeto

```
worker/
├── cmd/
│   └── server/
│       └── main.go              # Aplicação principal
├── internal/
│   ├── domain/
│   │   ├── entities/            # Entidades do domínio
│   │   └── services/            # Serviços de negócio
│   └── infrastructure/
│       ├── queue/
│       │   └── sqs_service.go   # Serviço SQS
│       ├── storage/
│       │   ├── s3_service.go    # Serviço S3
│       │   └── zip_service.go   # Serviço de compactação
│       └── video/
│           └── ffmpeg_service.go # Extração de frames
├── outputs/                     # Arquivos ZIP gerados
├── temp/                        # Arquivos temporários
├── docker-compose.yml           # Configuração LocalStack
├── setup-localstack.sh          # Script de configuração
├── go.mod                       # Dependências Go
└── README.md                    # Este arquivo
```

## 📚 API Reference

### Serviços Principais

#### S3Service
- `NewS3Service(bucket string)`: Cria novo serviço S3
- `DownloadVideo(key, localPath string)`: Baixa vídeo do S3
- `UploadZip(localPath, key string)`: Faz upload de ZIP para S3
- `ListVideos()`: Lista vídeos no bucket

#### SQSService
- `NewSQSService(queueURL string)`: Cria novo serviço SQS
- `ReceiveMessages()`: Recebe mensagens da fila
- `DeleteMessage(receiptHandle string)`: Remove mensagem processada
- `SendMessage(videoKey string)`: Envia nova mensagem

#### Formato da Mensagem SQS

```json
{
  "video_key": "nome_do_video.mp4",
  "video_id": "identificador_unico"
}
```

## 🛠 Comandos Úteis

### LocalStack
```bash
# Iniciar
docker-compose up -d

# Parar
docker-compose down

# Ver logs
docker-compose logs localstack

# Status da saúde
curl http://localhost:4566/_localstack/health
```

### S3 Operations
```bash
# Listar buckets
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 ls

# Listar arquivos no bucket
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 ls s3://video-bucket/

# Upload arquivo
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 cp /caminho/arquivo s3://video-bucket/
```

### SQS Operations
```bash
# Listar filas
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 sqs list-queues

# Enviar mensagem
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 sqs send-message --queue-url http://localhost:4566/000000000000/video-processing-queue --message-body '{"video_key":"video.mp4","video_id":"123"}'

# Receber mensagens
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 sqs receive-message --queue-url http://localhost:4566/000000000000/video-processing-queue
```

## 🐛 Troubleshooting

### Erro: "NoCredentialProviders"
- **Problema**: Credenciais AWS não configuradas nos serviços
- **Solução**: Verificar se os serviços S3 e SQS têm `credentials.NewStaticCredentials("test", "test", "")`

### Erro: "Container not running"
- **Problema**: LocalStack não está executando
- **Solução**: `docker-compose up -d` e verificar `docker ps`

### Erro: "Device or resource busy"
- **Problema**: Conflito de volumes do Docker
- **Solução**: `docker-compose down --volumes` e `docker volume prune -f`

### FFmpeg não encontrado
- **Problema**: FFmpeg não instalado no sistema
- **Solução**: Instalar FFmpeg e adicionar ao PATH do sistema

## 🤝 Contribuição

1. Faça fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.

## 👥 Autores

- **Bruno** - Desenvolvimento inicial
- **Iana**  - Desenvolvimento inicial
- **Juliano** - Desenvolvimento inicial
- **Rafaelle** - Desenvolvimento inicial
---

**Nota**: Este projeto foi desenvolvido como parte do Hackathon FIAP e utiliza LocalStack para simular serviços AWS em ambiente local.