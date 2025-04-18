VERSION 0.8

FROM golang:1.23.4-alpine
WORKDIR /go-workdir

deps:
    COPY go.mod go.sum ./
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

build:
    FROM +deps
    COPY . .
    RUN go mod download
    RUN go build -o build/ .
    SAVE ARTIFACT build/ AS LOCAL ./
