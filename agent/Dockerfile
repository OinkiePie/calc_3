FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем файлы управления зависимостями
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код приложения
COPY .env ./.env
COPY pkg ./pkg
COPY agent ./agent
COPY config ./config

# Собираем приложение
RUN go build -o /app/agent-app ./agent/cmd/main.go

FROM alpine:latest

# ТА САМАЯ СТРОКА ДЛЯ УСТАНОВКИ BASH ↓↓↓
# RUN apk update && apk add bash

WORKDIR /app

# Копируем исполняемый файл
COPY --from=builder /app/agent-app ./

# Копируем конфигурационные файлы
COPY --from=builder /app/.env ./.env
COPY --from=builder /app/config ./config

# Устанавливаем права на выполнение
RUN chmod +x ./agent-app

# Переменные окружения для gRPC
ENV ORCHESTRATOR_ADDR="orchestrator" \
    ORCHESTRATOR_GRPC_PORT="50051"

CMD ["./agent-app"]