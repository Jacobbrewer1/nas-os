FROM golang:alpine as build
WORKDIR /build

COPY go.sum go.mod /build/

RUN go mod download
RUN go mod tidy

COPY . /build/
RUN go build -o nas-os

FROM ubuntu:latest

COPY --from=build /build/nas-os /usr/local/bin/nas-os
ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["nas-os"]
