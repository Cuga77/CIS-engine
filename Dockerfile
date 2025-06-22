# --- Сборочный этап ---
# ИСПРАВЛЕНО: Версия Go обновлена до 1.22, чтобы соответствовать go.mod
FROM golang:1.24-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum ./
# Эта команда теперь выполнится успешно
RUN go mod download

# Копируем весь исходный код проекта
COPY . .

# Собираем два бинарных файла: для API и для краулера
# Флаги -ldflags="-w -s" уменьшают размер бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /crawler ./cmd/crawler

# --- Финальный этап ---
# Используем минималистичный образ Alpine для финального контейнера
FROM alpine:latest

# Копируем скомпилированные бинарники из сборочного этапа
COPY --from=builder /api /api
COPY --from=builder /crawler /crawler

# Команды по умолчанию будут определены в docker-compose.yml
