FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/seed ./cmd/seed

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 1000 appuser

WORKDIR /app

COPY --from=builder /bin/server /bin/seed ./

RUN mkdir -p /app/storage && chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./server"]
