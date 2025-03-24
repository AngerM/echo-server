FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . /app
RUN go build -o echo-server

FROM alpine:latest
COPY --from=builder /app/echo-server /echo-server
EXPOSE 8080
ENTRYPOINT ["/echo-server"]
