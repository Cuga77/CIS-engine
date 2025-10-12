FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /crawler ./cmd/crawler
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /indexer ./cmd/indexer

FROM alpine:latest

COPY --from=build /api /api
COPY --from=build /crawler /crawler
COPY --from=build /indexer /indexer