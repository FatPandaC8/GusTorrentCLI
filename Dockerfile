# 1. Build stage
# Start from a base image that already has Go installed and Linux Alpine
# named it as builder
FROM golang:1.26-alpine AS builder

WORKDIR /app

# copy go.mod first (cache optimization)
COPY go.mod go.sum ./
RUN go mod download

# copy source code (everything)
COPY . .

# build binary
RUN go build -o app ./cmd

# 2. Run stage (small image)
# Start a fresh linux alpine (because I dont need Go to run the app, only to build it)
FROM alpine:latest

WORKDIR /app

# copy binary from builder
COPY --from=builder /app/app .

# run it
CMD ["./app"]