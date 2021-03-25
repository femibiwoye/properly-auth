# Start from golang base image
FROM golang:alpine as builder

# ENV GO111MODULE=on

# Add Maintainer info
LABEL maintainer="Femi Ibiwoye <femibiwoye@gmail.com>"

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
RUN go version

# Set the current working directory inside the container 

# creates working directory for program
WORKDIR /go/src/properlyauth

# copies all program files specified directory in the container
ADD . /go/src/properlyauth

#changes to the specified directory
WORKDIR /go/src/properlyauth

#downloads the required depencies using the go.mod file
RUN go mod download 

# RUN go mod download 


# Build the Go app
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

RUN go get github.com/githubnemo/CompileDaemon
# RUN go get github.com/gin-gonic/gin

# runs --build and the --command if any changes are detected in *.go file
ENTRYPOINT CompileDaemon --build="go build properlyauth" --command=./properlyauth
# ENTRYPOINT CompileDaemon --build="go mod download && go run ." --command=./main

EXPOSE 8080
EXPOSE 3306
