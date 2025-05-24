# syntax=docker/dockerfile:1

FROM golang:1.24 AS build

WORKDIR /app

COPY . .

RUN make

FROM gcr.io/distroless/static:nonroot

WORKDIR /app

LABEL org.opencontainers.image.source="https://github.com/YukariExpress/pgb"
LABEL org.opencontainers.image.description="Pythia Gata Bot (pgb): a Telegram bot in Go."
LABEL org.opencontainers.image.licenses="GPL-3.0-only"

COPY --from=build --chown=nonroot:nonroot /app/pgb ./pgb

ENV HOST=0.0.0.0
ENV PORT=8080

EXPOSE 8080
CMD ["./pgb"]
