include .env

MIGRATIONS_PATH ?= ./migrations
POSTGRES_URL ?= $(POSTGRES_URL)
MAIN = cmd/main.go
BUILD_DIR = bin
APP_NAME = order-service
SWAGGER_DIR = docs

.PHONY: migrate-create migrate-up migrate-down run build test lint clean gen-docs

# swagger docs
gen-docs:
	swag init -g $(MAIN) -o $(SWAGGER_DIR)

# migrations
migrate-create:
ifndef name
	$(error "Usage: make migrate-create name=MigrationName")
endif
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(POSTGRES_URL)" up

migrate-down:
ifndef name
	migrate -path $(MIGRATIONS_PATH) -database "$(POSTGRES_URL)" down 1
else
	migrate -path $(MIGRATIONS_PATH) -database "$(POSTGRES_URL)" down $(name)
endif

# app
run:
	go run $(MAIN)

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)

test:
	go test ./... -v

lint:
	golangci-lint run

clean:
	rm -rf $(BUILD_DIR)