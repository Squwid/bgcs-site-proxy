FROM golang:1.22.6-alpine3.19 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bgcs-site-proxy
FROM alpine:3.20.2

COPY --from=builder /app/bgcs-site-proxy /bgcs-site-proxy

ENTRYPOINT ["/bgcs-site-proxy"] 