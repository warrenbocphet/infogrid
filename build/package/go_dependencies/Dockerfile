FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

ADD . ./src/infogrid
WORKDIR ./src/infogrid

# Fetch dependencies.
# Using go get.
RUN go get ./...