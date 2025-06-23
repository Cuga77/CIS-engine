FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /crawler ./cmd/crawler
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /indexer ./cmd/indexer

FROM alpine:latest

RUN apk --no-cache add ca-certificates
COPY YandexInternalRootCA.pem /usr/local/share/ca-certificates/YandexInternalRootCA.crt
RUN update-ca-certificates

COPY --from=builder /api /api
COPY --from=builder /crawler /crawler
COPY --from=builder /indexer /indexer