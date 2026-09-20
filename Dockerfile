FROM golang:1.26-alpine AS builder
LABEL authors="nhassl3"

RUN apk add --no-cache git gcc musl-dev

ENV GOTOOLCHAIN=auto
WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/ipbuild ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 10001 app
USER app

WORKDIR /app

COPY --from=builder /app/bin/ipbuild ./ipbuild
COPY config ./config
COPY migrations ./migrations

EXPOSE 8080

ENTRYPOINT ["./ipbuild"]
HEALTHCHECK --interval=30s CMD wget -q0- http://localhost:8080/health || exit 1