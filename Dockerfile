FROM golang:1.24-alpine AS builde

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/api     ./cmd/api/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/crawler ./cmd/crawler/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/indexer ./cmd/indexer/main.go

FROM gcr.io/distroless/static-debian12 AS final

RUN groupadd --system nonroot && \
    useradd --system --gid nonroot nonroot

COPY --from=builder /bin/api /api
COPY --from=builder /bin/crawler /crawler
COPY --from=builder /bin/indexer /indexer
COPY --from=builder --chown=nonroot:nonroot /app/frontend /frontend

USER nonroot