APP_NAME ?= financebot
BIN_DIR ?= bin
MAIN_PKG := ./cmd

.PHONY: build run test fmt tidy clean

build: ## Compile the bot into $(BIN_DIR)/$(APP_NAME)
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(MAIN_PKG)

run: ## Run the bot with go run
	go run $(MAIN_PKG)

test: ## Execute the full test suite
	go test ./...

fmt: ## Format Go sources
	gofmt -w cmd internal

tidy: ## Sync module dependencies
	go mod tidy

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR)
