# Makefile for the project
ROOT := $(shell pwd)
GO := go
GOBUILD := $(GO) build
GOFILES := $(shell find . -name "*.go" -type f)
GOFMT := $(GO) fmt
GOAIR := bin/air
MAIN_PACKAGE_PATH := ./cmd/server/main.go
BINARY_NAME := server
CLIENT_PATH := ./cmd/client/main.go
CLIENT_NAME := chat_client


$(GOAIR):
	@echo "Setting up air for hot reloading..."
	@mkdir -p bin
	@curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b ./bin

.PHONY: setup
setup: $(GOAIR) 
	@echo "Installing tools..."

.PHONY: tidy
tidy:
	@echo "Tidying up the go.mod and go.sum files..."
	@$(GOFMT) ./...
	@$(GO) mod tidy

.PHONY: build
build: 
	@echo "Building the application..."
	@$(GOBUILD) -o ./app/${BINARY_NAME} ${MAIN_PACKAGE_PATH}

.PHONY: run
run: setup
	@echo "Running the server..."
	@$(GOAIR)


.PHONY: build-client
build-client: 
	@echo "Building the client application..."
	@$(GOBUILD) -o ./app/${CLIENT_NAME} ${CLIENT_PATH}


.PHONY: run-client
run-client: build-client 
	@echo "Running the client..."
	./app/${CLIENT_NAME} bob 

