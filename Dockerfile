FROM docker.io/library/golang:1.24-alpine AS builder

WORKDIR /app
COPY . ./

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /wspad ./cmd/wspad

FROM alpine:latest

WORKDIR /app

COPY --from=builder /wspad /wspad
COPY web ./web

EXPOSE 8080

ENTRYPOINT ["/wspad"]
