# Makefile for the project
ROOT := $(shell pwd)
GO := go
GOBUILD := $(GO) build
GOFILES := $(shell find . -name "*.go" -type f)
GOFMT := $(GO) fmt
MAIN_PACKAGE_PATH := ./main.go
BINARY_NAME := server


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
run: build
	@echo "Running the server..."
	./app/${BINARY_NAME} 	

