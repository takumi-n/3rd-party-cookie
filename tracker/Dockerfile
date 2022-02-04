FROM golang:1.16.13-bullseye as builder

WORKDIR /go/src/app
COPY . /go/src/app

RUN go build -o /go/bin/app

ENTRYPOINT [ "/go/bin/app" ]