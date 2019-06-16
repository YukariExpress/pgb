FROM golang:1.12-alpine
WORKDIR /pgb/
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app -mod=vendor .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /pgb/app .
CMD ["./app"]
