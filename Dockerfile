FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# -ldflags="-s -w" strips debug symbols -> smaller binary
RUN go build -ldflags="-s -w" -o /app/server .
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY --from=builder /app/public ./public
COPY --from=builder /app/db/migrations ./db/migrations
EXPOSE 8080
ENTRYPOINT ["/app/server"]