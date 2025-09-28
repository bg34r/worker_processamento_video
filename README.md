# Video Processing Worker 🎬

Sistema de processamento de vídeos que consome mensagens SQS, extrai frames com FFmpeg e gera arquivos ZIP. Executa com Docker Compose (local) ou Kubernetes (produção).

## 📋 Índice

- [Arquitetura](#-arquitetura)
- [Execução Rápida](#-execução-rápida)
- [Execução Local](#-execução-local---docker-compose)
- [Execução Kubernetes](#-execução-kubernetes)
- [Estrutura do Projeto](#-estrutura-do-projeto)
- [Troubleshooting](#-troubleshooting)

## 🏗 Arquitetura

**Fluxo**: `SQS Queue → Worker → S3 Download → FFmpeg → ZIP → S3 Upload → Kafka`

**Responsabilidades do Worker:**
- ✅ Consumir mensagens SQS
- ✅ Processar vídeos (extrair frames)  
- ✅ Gerar ZIPs e upload S3
- ✅ Notificações Kafka
- ❌ Upload inicial de vídeos (outro serviço)

## ⚡ Execução Rápida

```bash
# 1. Iniciar infraestrutura + verificação
bash ./setup-notification-queue.sh

# 2. Executar workers
docker-compose up -d

# 3. Verificar logs
docker-compose logs -f
```



## 🐳 Execução Local - Docker Compose

```bash
# 1. Configurar infraestrutura AWS
bash ./setup-notification-queue.sh

# 2. Iniciar workers
docker-compose up -d

# 3. Monitorar
docker-compose logs -f

# 4. Escalar (opcional)
docker-compose up -d --scale video-worker=3
```

**Comandos úteis:**
```bash
# Status
docker-compose ps

# Parar
docker-compose down

# Worker standalone
go run cmd/server/main.go
```

## ☸️ Execução em Produção - Kubernetes

### Setup Rápido
```bash
# 1. Instalar KEDA (auto-scaler)
helm repo add kedacore https://kedacore.github.io/charts
helm install keda kedacore/keda --namespace keda-system --create-namespace

# 2. Deploy automático
.\k8s-deploy.ps1  # Windows
bash k8s-deploy.sh  # Linux/macOS

# 3. Monitorar
kubectl get pods -l app=video-processer-worker
kubectl logs -f deployment/video-processer-worker
```

### Comandos Úteis
```bash
# Escalar manualmente
kubectl scale deployment video-processer-worker --replicas=5

# Ver auto-scaling
kubectl get scaledobjects

# Métricas
kubectl top pods

# Limpeza
kubectl delete -f k8s/





## 📁 Estrutura do Projeto

cmd/server/            # Aplicação principal
internal/
  ├── notification/    # Kafka notifications  
  ├── queue/           # SQS service
  ├── storage/         # S3 + ZIP services
  └── video/           # FFmpeg service
k8s/                   # Kubernetes configs
.github/workflows/     # CI/CD
docker-compose.yml     # Ambiente local

## 🛠 Troubleshooting

**Problemas comuns:**

- **LocalStack não conecta**: `docker-compose up -d` e verificar `docker ps`
- **Recursos AWS ausentes**: Execute `bash ./setup-notification-queue.sh`
- **FFmpeg não encontrado**: Instalar FFmpeg no sistema
- **Verificar saúde**: `curl http://localhost:4566/_localstack/health`

## 🤝 Contribuição

1. Faça fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.

## � Qualidade de Código e Testes

**Cobertura atual**: 66% (mínimo: 60%)

```bash
# Executar testes
go test ./... -v

# Com cobertura
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# CI/CD: GitHub Actions + SonarCloud
```

## �👥 Autores

- **Bruno** - Desenvolvimento inicial e infraestrutura Kubernetes
- **Iana**  - Desenvolvimento inicial
- **Juliano** - Desenvolvimento inicial
- **Rafaelle** - Desenvolvimento inicial

---

**Nota**: Este projeto foi desenvolvido como parte do Hackathon FIAP e utiliza LocalStack para simular serviços AWS em ambiente local.
