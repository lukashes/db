FROM golang:latest

MAINTAINER Yegor Lukash

RUN mkdir -p src/github.com/lukashes/db

COPY . src/github.com/lukashes/db

ENTRYPOINT ["go", "run", "src/github.com/lukashes/db/main.go"]