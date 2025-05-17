FROM golang:1.23-alpine

RUN apk add --no-cache git

# Рабочая директория
WORKDIR /app

# Сначала копируем только go.mod и go.sum
COPY go.mod go.sum ./

# Устанавливаем зависимости
RUN go mod tidy

# Устанавливаем air и migrate
RUN go install github.com/cosmtrek/air@v1.40.4
RUN go install github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.0

# Теперь копируем весь код
COPY . .

CMD ["air"]
