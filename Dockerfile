FROM golang:alpine AS builder


# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git bash && mkdir -p /build/ciscofon-server

WORKDIR /build/ciscofon-server

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download -json

COPY . .

RUN mkdir -p /app && CGO_ENABLED=0 GOOS=${TARGETPLATFORM%%/*} GOARCH=${TARGETPLATFORM##*/} \
    go build -ldflags='-s -w -extldflags="-static"' -o /app/ciscofon-server ./cmd/ciscofon-server

FROM scratch AS bin-unix
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/ciscofon-server /app/ciscofon-server

LABEL org.opencontainers.image.description="A simple TFTP and HTTP server for Cisco IP Phones, with debugging functionality."

ENTRYPOINT ["/app/ciscofon-server"]
