FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server
RUN CGO_ENABLED=1 GOOS=linux go build -o scrape-maps ./cmd/scrape-maps
RUN CGO_ENABLED=1 GOOS=linux go build -o scrape-player ./cmd/scrape-player
RUN CGO_ENABLED=1 GOOS=linux go build -o weekly-ingest ./cmd/weekly-ingest

FROM alpine:latest

RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/scrape-maps .
COPY --from=builder /app/scrape-player .
COPY --from=builder /app/weekly-ingest .

COPY --from=builder /app/migrations ./migrations

RUN mkdir -p /app/data

ENV DB_PATH=/app/data/surfstats.db

EXPOSE 8080

CMD ["./server"]