# syntax=docker/dockerfile:1

## Build
FROM golang:1.20-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /radicale_exporter

## Deploy
FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /radicale_exporter /radicale_exporter

USER nonroot:nonroot

ENTRYPOINT ["/radicale_exporter"]
