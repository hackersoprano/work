FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей
COPY ../go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
#COPY ../cmd/server/main.go .
COPY . .
# Собираем приложение
RUN go build -o app ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app .

# Открываем порт
EXPOSE 8080

# Команда запуска
CMD ["./app"]
