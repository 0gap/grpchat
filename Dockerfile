FROM golang:alpine as build-env

ENV GO111MODULE=on

RUN apk update && apk add bash ca-certificates git gcc g++ libc-dev

RUN mkdir -p /learningGo/proto

WORKDIR /learningGo

COPY ./proto/service.pb.go /learningGo/proto
COPY ./proto/service_grpc.pb.go /learningGo/proto
COPY ./server.go /learningGo

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go build -o learningGo .

CMD ./learningGo -port 8080
