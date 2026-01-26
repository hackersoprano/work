FROM golang:1.25-alpine

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY main.go .

# Собираем приложение
RUN go build -o main .

# Открываем порт
EXPOSE 8080

# Команда запуска
CMD ["./main"]
