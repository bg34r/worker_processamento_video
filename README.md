# Video Processer Worker

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
4. **Video Processer Worker**: Aplicação que consome mensagens e processa vídeos
5. **FFmpeg**: Extrai frames dos vídeos
6. **Kafka**: Recebe notificações de sucesso/falha do processamento

### Responsabilidades

> **⚠️ IMPORTANTE**: Este repositório contém apenas os **workers de processamento**. 

**✅ O que os workers fazem:**
- Consumir mensagens da fila SQS
- Baixar vídeos já existentes no S3
- Extrair frames dos vídeos
- Criar arquivos ZIP com os frames
- Fazer upload dos ZIPs processados para S3
- Enviar notificações Kafka sobre sucesso/falha

**❌ O que os workers NÃO fazem:**
- Upload de vídeos originais para S3 (responsabilidade do serviço de upload)
- Autenticação de usuários
- API REST para recebimento de vídeos
- Interface web para upload

### Fluxo de Processamento

```
[Serviço Upload] → Vídeo S3 → Mensagem SQS → Worker → Frames → ZIP S3 → Kafka Notification
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

### Modo Desenvolvimento (Worker único)

#### 1. Inicie o LocalStack

```bash
# Se não executou o setup ainda
bash ./setup-localstack.sh

# Ou apenas para iniciar
docker-compose up -d
```

#### 2. Verifique se o LocalStack está funcionando

```bash
curl http://localhost:4566/_localstack/health
```

#### 3. Execute o Worker

```bash
go run cmd/server/main.go
```

Você deve ver a mensagem:
```
🚀 video-processer-worker-1 (ID: 1) iniciando...
✅ video-processer-worker-1 iniciado - monitorando fila SQS...
```

### Modo Produção (Múltiplos Workers) 🚀

#### Usando Scripts de Gerenciamento

**Windows (PowerShell):**
```powershell
# Iniciar 3 workers (padrão)
.\manage-workers.ps1 start

# Escalar para 5 workers
.\manage-workers.ps1 scale -Workers 5

# Ver status dos workers
.\manage-workers.ps1 status

# Ver logs do worker específico
.\manage-workers.ps1 logs -WorkerId 2

# Parar todos os workers
.\manage-workers.ps1 stop
```

**Linux/macOS (Bash):**
```bash
# Dar permissão de execução
chmod +x manage-workers.sh

# Iniciar 3 workers (padrão)
./manage-workers.sh start

# Escalar para 5 workers
./manage-workers.sh scale 5

# Ver status dos workers
./manage-workers.sh status

# Ver logs do worker específico
./manage-workers.sh logs 2

# Parar todos os workers
./manage-workers.sh stop
```

#### Usando Docker Compose Diretamente

```bash
# Iniciar todos os 3 workers configurados
docker-compose up -d

# Iniciar workers específicos
docker-compose up -d video-processer-worker-1 video-processer-worker-2

# Ver status
docker-compose ps

# Ver logs de todos os workers
docker-compose logs -f

# Ver logs de worker específico
docker-compose logs -f video-processer-worker-1

# Parar todos
docker-compose down
```

### 📊 Monitoramento em Tempo Real

```bash
# Status detalhado dos workers
docker stats video-worker-1 video-worker-2 video-worker-3

# Logs combinados com timestamps
docker-compose logs --timestamps --follow

# Monitorar uso de recursos
watch -n 2 'docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}"'
```

## 🧪 Testando o Sistema com Múltiplos Workers

### Teste de Escalabilidade - Processamento Paralelo

#### 1. Inicie múltiplos workers

```powershell
# Windows - Iniciar 3 workers
.\manage-workers.ps1 start

# Linux/macOS
./manage-workers.sh start
```

#### 2. Simulação de múltiplos vídeos para teste (apenas para desenvolvimento)

> **⚠️ IMPORTANTE**: Os workers de processamento **NÃO** são responsáveis por fazer upload de vídeos para o S3. 
> Esta responsabilidade é do serviço de upload em outro repositório. 
> Os comandos abaixo são **apenas para testes** durante o desenvolvimento.

```bash
# APENAS PARA TESTE - simular vídeos já existentes no S3
# Em produção, os vídeos chegam via serviço de upload
docker cp "C:\Users\Caminho\Para\video\video1.mp4" localstack-main:/tmp/video1.mp4
docker cp "C:\Users\Caminho\Para\video\video2.mp4" localstack-main:/tmp/video2.mp4
docker cp "C:\Users\Caminho\Para\video\video3.mp4" localstack-main:/tmp/video3.mp4

# APENAS PARA TESTE - simular vídeos no S3 (normalmente feito pelo serviço de upload)
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 cp /tmp/video1.mp4 s3://video-service-bucket/
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 cp /tmp/video2.mp4 s3://video-service-bucket/
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 cp /tmp/video3.mp4 s3://video-service-bucket/
```

#### 3. Enviar mensagens SQS para processamento paralelo

```bash
# Processar todos os vídeos simultaneamente
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 sqs send-message --queue-url http://localhost:4566/000000000000/video-processing-queue --message-body '{"video_key":"video1.mp4","video_id":"video1_001"}'

docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 sqs send-message --queue-url http://localhost:4566/000000000000/video-processing-queue --message-body '{"video_key":"video2.mp4","video_id":"video2_001"}'

docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 sqs send-message --queue-url http://localhost:4566/000000000000/video-processing-queue --message-body '{"video_key":"video3.mp4","video_id":"video3_001"}'
```

#### 4. Monitorar processamento paralelo

Os workers devem mostrar mensagens identificadas:
```
🎬 video-processer-worker-1 - Processando vídeo: video1.mp4
🎬 video-processer-worker-2 - Processando vídeo: video2.mp4
🎬 video-processer-worker-3 - Processando vídeo: video3.mp4
⚙️ video-processer-worker-1 - Extraindo frames do vídeo: video1.mp4
⚙️ video-processer-worker-2 - Extraindo frames do vídeo: video2.mp4
📦 video-processer-worker-3 - Criando arquivo ZIP: video3_frames.zip
🎉 video-processer-worker-1 - Processamento concluído: video1.mp4 -> processed/video1_frames.zip
```

### Teste com Vídeos Reais (apenas desenvolvimento)

> **⚠️ IMPORTANTE**: Em produção, os vídeos são enviados pelo **serviço de upload** (repositório separado) via API REST.
> Os comandos abaixo são **apenas para testes locais** dos workers de processamento.

#### 1. Simular vídeo já no S3 (para teste)

```bash
# APENAS PARA TESTE - Em produção isso é feito pelo serviço de upload
docker cp "C:\Users\Caminho\Para\video\nomeDoVideo0.mp4" localstack-main:/tmp/nomeDoVideo0.mp4

# APENAS PARA TESTE - Simular upload feito pelo outro serviço
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 cp /tmp/nomeDoVideo0.mp4 s3://video-service-bucket/
```

#### 2. Testar com múltiplos vídeos (desenvolvimento)

```bash
# APENAS PARA TESTE - Simular múltiplos vídeos já enviados pelo serviço de upload
docker cp "C:\Users\Caminho\Para\video\nomeDoVideo1.mp4" localstack-main:/tmp/nomeDoVideo1.mp4
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 cp /tmp/nomeDoVideo1.mp4 s3://video-service-bucket/

docker cp "C:\Users\Caminho\Para\video\nomeDoVideo2.mp4" localstack-main:/tmp/nomeDoVideo2.mp4
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 cp /tmp/nomeDoVideo2.mp4 s3://video-service-bucket/
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

# Listar arquivos no bucket (vídeos originais e ZIPs processados)
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 ls s3://video-service-bucket/ --recursive

# APENAS PARA TESTE - Upload de arquivo (normalmente feito pelo serviço de upload)
docker exec -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 localstack-main aws --endpoint-url=http://localhost:4566 s3 cp /caminho/arquivo s3://video-service-bucket/
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

## � Qualidade de Código e Testes

### Cobertura de Testes

Este projeto mantém padrões rigorosos de qualidade:

- **Cobertura mínima do repositório**: 60%
- **Cobertura mínima para código novo**: 40%
- **Análise estática**: SonarQube

### Executando Testes Localmente

```bash
# PowerShell (Windows)
.\test-coverage.ps1

# Bash (Linux/macOS)  
./test-coverage.sh

# Apenas testes (sem SonarQube)
.\test-coverage.ps1 -OnlyTests

# Com SonarQube
.\test-coverage.ps1 -SonarToken "seu_token_aqui"
```

### Relatórios Gerados

- `coverage.out` - Cobertura para SonarQube
- `coverage.html` - Relatório visual de cobertura  
- `coverage_report.txt` - Relatório texto

### CI/CD Pipeline

O projeto utiliza GitHub Actions com:

1. **Testes automatizados** com LocalStack
2. **Verificação de cobertura** (60% repositório, 40% código novo)
3. **Análise SonarQube** com Quality Gates
4. **Scan de segurança** com gosec
5. **Build Docker** para ambientes dev e k8s
6. **Deploy automático** para staging/produção

### Configuração SonarQube

Para usar SonarQube localmente:

1. Configure as variáveis de ambiente:
```bash
export SONAR_TOKEN=seu_token_sonarqube
export SONAR_HOST_URL=http://localhost:9000
```

2. Execute a análise:
```bash
./test-coverage.sh
```

## �👥 Autores

- **Bruno** - Desenvolvimento inicial e infraestrutura Kubernetes
- **Iana**  - Desenvolvimento inicial
- **Juliano** - Desenvolvimento inicial
- **Rafaelle** - Desenvolvimento inicial

---

**Nota**: Este projeto foi desenvolvido como parte do Hackathon FIAP e utiliza LocalStack para simular serviços AWS em ambiente local.