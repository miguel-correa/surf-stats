FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest

RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

COPY --from=builder /app/server .

COPY --from=builder /app/migrations ./migrations

RUN mkdir -p /app/data

ENV DB_PATH=/app/data/surfstats.db

EXPOSE 8080

CMD ["./server"]