FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY src src/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o vifa ./src/

FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/vifa .
COPY static static/
COPY res res/
COPY logs logs/
CMD ["./vifa"]