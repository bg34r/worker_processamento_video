FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o worker ./cmd/server/main.go

# Etapa final com Nginx, Go e ffmpeg
FROM nginx:alpine

WORKDIR /app

# Instala ffmpeg e curl para health checks e dos2unix para corrigir quebras de linha
RUN apk add --no-cache bash curl dos2unix ffmpeg

COPY --from=builder /app/worker /app/worker
COPY web/ /app/web/
COPY scripts/start.sh /app/start.sh

COPY nginx.conf /etc/nginx/nginx.conf

# Criar diretórios necessários, corrigir quebras de linha e dar permissões
RUN mkdir -p /app/outputs /app/temp && \
    dos2unix /app/start.sh && \
    chmod +x /app/start.sh

EXPOSE 80
EXPOSE 8080

# Usar exec form com script
CMD ["bash", "/app/start.sh"]