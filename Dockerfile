# Dockerfile for Pythia Gata Bot (pgb)
FROM golang:1.24-alpine AS build

WORKDIR /app
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o pgb ./pgb.go

FROM alpine:3.21
WORKDIR /app
COPY --from=build /app/pgb ./pgb

# Set environment variables for runtime configuration
ENV HOST=0.0.0.0
ENV PORT=8080

EXPOSE 8080
CMD ["./pgb"]
