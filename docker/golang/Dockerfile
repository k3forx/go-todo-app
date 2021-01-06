FROM golang:1.15.6

RUN apt-get update && \
    apt-get install -y vim default-mysql-client

WORKDIR /usr/src/go-todo-app

RUN go mod init main

COPY . .

RUN go build
