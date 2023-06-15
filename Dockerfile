ARG GO_VERSION=1.17

# FIRST STAGE: build
FROM golang:${GO_VERSION}-alpine AS builder

# force the go compiler to use modules 
ENV GO111MODULE=on

# install dependencies rewuire to build
RUN apk add --update make git gcc libc-dev

WORKDIR /app

# download dependencies
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./

# compile binary
RUN make build

# FINAL STAGE: run application
FROM alpine:3.14.9

# dev env always default
ENV ENV=development

COPY --from=builder /app /app

ENTRYPOINT ["/app/bin/containerd-healthcheck"]
