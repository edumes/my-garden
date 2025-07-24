FROM golang:1.24.4-alpine3.19 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/main.go

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/env.example .env

EXPOSE 8080

CMD ["./main"] 