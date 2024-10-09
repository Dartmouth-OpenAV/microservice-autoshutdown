FROM golang:latest

COPY source /go/src

ENV GOPATH=
WORKDIR /go
RUN go mod init github.com/Dartmouth-OpenAV/microservice-autoshutdown
RUN go mod tidy
WORKDIR /go/src
RUN go get -u
RUN go build -o /go/bin/microservice

ENTRYPOINT /go/bin/microservice
