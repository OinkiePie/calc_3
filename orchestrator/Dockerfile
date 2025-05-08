FROM golang:1.23-alpine AS builder

WORKDIR /app

# Устанавливаем компилятор C и зависимости SQLite
RUN apk add --no-cache gcc musl-dev git

# Включаем CGO (обязательно для SQLite3)
ENV CGO_ENABLED=1

# Копируем файлы управления зависимостями
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код приложения
COPY .env ./.env
COPY pkg ./pkg
COPY orchestrator ./orchestrator
COPY config ./config

# Собираем приложение
RUN go build -o /app/orchestrator-app ./orchestrator/cmd/main.go

FROM alpine:latest

# ТА САМАЯ СТРОКА ДЛЯ УСТАНОВКИ BASH ↓↓↓
# RUN apk update && apk add bash

WORKDIR /app

# Устанавливаем SQLite
RUN apk add --no-cache sqlite

# Копируем исполняемый файл
COPY --from=builder /app/orchestrator-app ./

# Копируем конфигурационные файлы
COPY --from=builder /app/.env ./.env
COPY --from=builder /app/config ./config

# Устанавливаем права на выполнение
RUN chmod +x ./orchestrator-app

# Объявляем порты
ARG HTTP_PORT=8080
ARG GRPC_PORT=50051

# Экспортируем порты
EXPOSE $HTTP_PORT $GRPC_PORT

# Передаем переменные окружения
ENV ORCHESTRATOR_HTTP_PORT=$HTTP_PORT \
    ORCHESTRATOR_GRPC_PORT=$GRPC_PORT \
    ORCHESTRATOR_ADDR=0.0.0.0

# Запускаем приложение
CMD ["./orchestrator-app"]