include .env

MIGRATIONS_PATH ?= ./migrations
POSTGRES_URL ?= $(POSTGRES_URL)
MAIN = cmd/main.go
ORDER_GENERATOR = tests/order-generator/main.go
REQUESTER = tests/requester/main.go
BUILD_DIR = bin
APP_NAME = order-service
SWAGGER_DIR = docs

.DEFAULT_GOAL := help

.PHONY: migrate-create migrate-up migrate-down run build test lint clean gen-docs help run-generator run-requester

help: # Show available make commands
	@grep -E '^[a-zA-Z0-9 -]+:.*#' Makefile | sort | while read -r l; do \
		printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; \
	done

gen-docs: # Generate Swagger documentation
	@swag init -g $(MAIN) -o $(SWAGGER_DIR)

migrate-create: # Create a new database migration (Usage: make migrate-create name=MigrationName)
ifndef name
	$(error "Usage: make migrate-create name=MigrationName")
endif
	@migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

migrate-up: # Apply all up migrations
	@migrate -path $(MIGRATIONS_PATH) -database "$(POSTGRES_URL)" up

migrate-down: # Rollback migrations (Usage: make migrate-down name=N | default=1)
ifndef name
	@migrate -path $(MIGRATIONS_PATH) -database "$(POSTGRES_URL)" down 1
else
	@migrate -path $(MIGRATIONS_PATH) -database "$(POSTGRES_URL)" down $(name)
endif

run: # Run the application
	@go run $(MAIN)

run-generator: # Run the order generator
	@go run $(ORDER_GENERATOR)

run-requester: # Run requester
	@go run $(REQUESTER)

build: # Build the binary
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)

test: # Run all tests
	@go test ./... -v

lint: # Run linter
	@golangci-lint run

clean: # Remove build artifacts
	@rm -rf $(BUILD_DIR)