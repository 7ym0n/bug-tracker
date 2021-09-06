GOCMD=go
GO111MODULE=on
GOPROXY="https://goproxy.cn,direct"
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get -u

OS_NAME=$(shell uname -s)
LC_OS_NAME=$(shell echo $(OS_NAME) | tr '[A-Z]' '[a-z]' | awk -F '_' '{print $$1}')
ifeq ($(LC_OS_NAME), cygwin)
BINARY_NAME=bugtracker.exe
else
BINARY_NAME=bugtracker
endif

all: test build
build:
	@echo $(LC_OS_NAME)
	$(GOBUILD) -o $(BINARY_NAME) -v
test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

deps:
	$(GOCMD) mod tidy

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME) -v
docker-build:
	docker run --rm -it -v $(shell pwd):/working -w /working \
	-e CGO_ENABLED=0 \
	-e GOOS=linux \
	-e GOPROXY=$(GOPROXY) \
	-e GO111MODULE=$(GO111MODULE) golang:alpine /bin/sh -c \
	"go mod tidy && \
	go build -o $(BINARY_NAME) -v"
