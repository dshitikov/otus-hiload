.PHONY: all docker clean
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
BINARY_NAME=bin/build

all: build
docker-build: build docker
up: docker-build compose-up
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME) -v src/main.go
clean:
	$(GOCLEAN) src/
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
docker:
	 docker build --file="./docker/Dockerfile" --tag="otus-hiload:v1" --force-rm .
compose-up:
	 docker-compose --file "./docker/docker-compose.yml" --project-directory . up --abort-on-container-exit
