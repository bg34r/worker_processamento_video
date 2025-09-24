FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o worker ./cmd/server/main.go

# Etapa final com Nginx, Go e ffmpeg
FROM nginx:alpine

WORKDIR /app

# Instala ffmpeg
RUN apk add --no-cache ffmpeg

COPY --from=builder /app/worker /app/worker
COPY web/ /app/web/
COPY uploads/ /app/uploads/
COPY outputs/ /app/outputs/
COPY temp/ /app/temp/

COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80
EXPOSE 8080

CMD sh -c "/app/worker & nginx -g 'daemon off;'"