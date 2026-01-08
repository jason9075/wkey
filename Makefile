.PHONY: build run test mock help

# Default message for mock target
MSG ?= Hello World

build: ## Build the binary
	go build -o wkey ./cmd/wkey

run: ## Run the application
	go run ./cmd/wkey

test: ## Run tests
	go test ./...

mock: ## Run with mock response, auto stop after 2s
	@echo "▶ start wkey (mock mode)"
	@go run ./cmd/wkey --mock-response "$(MSG)" &\
	sleep 5; \
	echo "■ stop wkey";
	@go run ./cmd/wkey

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
