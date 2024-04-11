FROM golang:1.22 AS builder

WORKDIR /app

COPY . .

RUN go mod download

EXPOSE 8080

RUN go build -ldflags '-linkmode external -w -extldflags "-static"' -o /server ./cmd/main.go

FROM alpine:latest

WORKDIR /

COPY --from=builder /server /server

ENV REDIS_HOST=redis
ENV RABBITMQ_HOST=rabbitmq
ENV RABBITMQ_USER=user
ENV RABBITMQ_PASSWORD=password
ENV SQLITE_DB_PATH=/data/app.db/

EXPOSE 8080

ENTRYPOINT ["/server"]