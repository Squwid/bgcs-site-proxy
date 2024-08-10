FROM golang:1.22.6-alpine3.19 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bgcs-site-proxy
FROM scratch

COPY --from=builder /app/bgcs-site-proxy /bgcs-site-proxy

ENTRYPOINT ["/bgcs-site-proxy"] 