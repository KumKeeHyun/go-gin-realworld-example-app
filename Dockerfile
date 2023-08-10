FROM golang:1.19.3-alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=arm64

WORKDIR /build
COPY . .
RUN go mod download
RUN go build -ldflags="-s -w" -trimpath ./cmd/...

WORKDIR /dist
RUN cp /build/restapp .

FROM alpine:latest

RUN apk add --update \
    sqlite-dev \
    && rm -rf /var/cache/apk/*

COPY --from=builder /dist/restapp /usr/local/bin/

ENTRYPOINT ["restapp"]