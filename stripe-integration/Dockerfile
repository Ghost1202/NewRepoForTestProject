# syntax=docker/dockerfile:1.7

FROM golang:1.24-alpine AS builder

ARG GITHUB_TOKEN
ARG GITHUB_PAT_TOKEN
ARG GOPRIVATE=github.com/turtlepavlo/*
ARG APP_PATH=./cmd/server
ARG APP_BINARY=stripe-integration

WORKDIR /src

RUN apk add --no-cache git openssh-client ca-certificates tzdata && update-ca-certificates
RUN mkdir -p -m 0700 /root/.ssh && ssh-keyscan github.com >> /root/.ssh/known_hosts

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GOPRIVATE=${GOPRIVATE}

COPY go.mod go.sum ./

RUN --mount=type=ssh,required=false \
    --mount=type=secret,id=github_token,required=false \
    TOKEN="$(cat /run/secrets/github_token 2>/dev/null || true)" && \
    if [ -n "${GITHUB_PAT_TOKEN}" ]; then \
      git config --global url."https://${GITHUB_PAT_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"; \
    elif [ -n "${GITHUB_TOKEN}" ]; then \
      git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"; \
    elif [ -n "${TOKEN}" ]; then \
      git config --global url."https://${TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"; \
    elif [ -S /run/buildkit/ssh_agent.0 ]; then \
      git config --global url."git@github.com:".insteadOf "https://github.com/"; \
    fi && \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/${APP_BINARY} ${APP_PATH}

FROM alpine:3.21 AS runtime

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && \
    adduser -S app -G app

COPY --from=builder /out/stripe-integration /app/stripe-integration

USER app

EXPOSE 9094 3755

CMD ["/app/stripe-integration"]
