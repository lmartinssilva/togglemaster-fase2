# Etapa 1 — Build
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN go build -o auth-service .

# Etapa 2 — Run
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/auth-service .
COPY db/init.sql db/init.sql
EXPOSE 8001
CMD ["./auth-service"]