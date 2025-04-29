FROM golang:1.23-alpine AS builder

WORKDIR /app

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

# Копируем исполняемый файл из билдера
COPY --from=builder /app/orchestrator-app ./

# Копируем .env файл
COPY --from=builder /app/.env ./.env

# Копируем config
COPY --from=builder /app/config ./config

# Устанавливаем права на выполнение
RUN chmod +x ./orchestrator-app

# Запускаем приложение
ARG DEFAULT_PORT=8080
ENV PORT_ORCHESTRATOR=$DEFAULT_PORT

EXPOSE $PORT_ORCHESTRATOR
CMD ["./orchestrator-app"]