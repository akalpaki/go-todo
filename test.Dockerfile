#syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build

WORKDIR /app


COPY go.mod ./
RUN go mod download

COPY . .

RUN go test -v ./...