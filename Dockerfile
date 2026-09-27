# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o cloudops-cockpit main.go

# Final runtime stage
FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/cloudops-cockpit /app/cloudops-cockpit
COPY --from=builder /app/config.json /app/config.json

EXPOSE 8080
VOLUME ["/data"]

ENV DB_PATH="/data/telemetry.db"

ENTRYPOINT ["/app/cloudops-cockpit"]
