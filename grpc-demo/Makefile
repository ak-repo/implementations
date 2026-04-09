.PHONY: all build run run-all install-proto clean test lint help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOMOD=$(GOCMD) mod
PROTOC=protoc

# Binaries
ORDER_BIN=bin/order-service
INVENTORY_BIN=bin/inventory-service
PAYMENT_BIN=bin/payment-service
NOTIFICATION_BIN=bin/notification-service
CLIENT_BIN=bin/client

# Proto paths
PROTO_PATHS=proto/order/v1 proto/inventory/v1 proto/payment/v1 proto/notification/v1

# Ports
ORDER_PORT=:50050
INVENTORY_PORT=:50051
PAYMENT_PORT=:50052
NOTIFICATION_PORT=:50053

# Colors
GREEN:=\033[0;32m
YELLOW:=\033[0;33m
BLUE:=\033[0;34m
NC:=\033[0m

# Default target
all: build

## build: Build all binaries
build: build-order build-inventory build-payment build-notification build-client

## build-order: Build order service
build-order:
	$(GOBUILD) -o $(ORDER_BIN) ./cmd/order-service

## build-inventory: Build inventory service
build-inventory:
	$(GOBUILD) -o $(INVENTORY_BIN) ./cmd/inventory-service

## build-payment: Build payment service
build-payment:
	$(GOBUILD) -o $(PAYMENT_BIN) ./cmd/payment-service

## build-notification: Build notification service
build-notification:
	$(GOBUILD) -o $(NOTIFICATION_BIN) ./cmd/notification-service

## build-client: Build client
build-client:
	$(GOBUILD) -o $(CLIENT_BIN) ./cmd/client

## run-order: Run order service
run-order:
	@echo "$(BLUE)Starting Order Service on $(ORDER_PORT)...$(NC)"
	$(GORUN) ./cmd/order-service

## run-inventory: Run inventory service
run-inventory:
	@echo "$(BLUE)Starting Inventory Service on $(INVENTORY_PORT)...$(NC)"
	$(GORUN) ./cmd/inventory-service

## run-payment: Run payment service
run-payment:
	@echo "$(BLUE)Starting Payment Service on $(PAYMENT_PORT)...$(NC)"
	$(GORUN) ./cmd/payment-service

## run-notification: Run notification service
run-notification:
	@echo "$(BLUE)Starting Notification Service on $(NOTIFICATION_PORT)...$(NC)"
	$(GORUN) ./cmd/notification-service

## run-client: Run client
run-client:
	@echo "$(BLUE)Running Client...$(NC)"
	$(GORUN) ./cmd/client

## run-all: Run all services (requires tmux or background jobs)
run-all: run-all-tmux

## run-all-tmux: Run all services in tmux windows (requires tmux)
run-all-tmux:
	@which tmux > /dev/null || (echo "tmux not found, use 'make run-all-bg' instead" && exit 1)
	@tmux kill-session -t grpc-demo 2>/dev/null || true
	@tmux new-session -d -s grpc-demo -n "inventory" "$(MAKE) run-inventory"
	@tmux new-window -t grpc-demo -n "payment" "$(MAKE) run-payment"
	@tmux new-window -t grpc-demo -n "notification" "$(MAKE) run-notification"
	@tmux new-window -t grpc-demo -n "order" "$(MAKE) run-order"
	@tmux new-window -t grpc-demo -n "client" "$(MAKE) run-client"
	@echo "$(GREEN)All services started in tmux session 'grpc-demo'$(NC)"
	@echo "$(YELLOW)Attach with: tmux attach -t grpc-demo$(NC)"

## run-all-bg: Run all services in background
run-all-bg:
	@echo "$(GREEN)Starting all services in background...$(NC)"
	@mkdir -p bin
	@$(MAKE) build
	@$(INVENTORY_BIN) > /dev/null 2>&1 &
	@echo "$(GREEN)Inventory Service started (PID: $$!)$(NC)"
	@$(PAYMENT_BIN) > /dev/null 2>&1 &
	@echo "$(GREEN)Payment Service started (PID: $$!)$(NC)"
	@$(NOTIFICATION_BIN) > /dev/null 2>&1 &
	@echo "$(GREEN)Notification Service started (PID: $$!)$(NC)"
	@$(ORDER_BIN) > /dev/null 2>&1 &
	@echo "$(GREEN)Order Service started (PID: $$!)$(NC)"
	@echo ""
	@echo "$(GREEN)All services running!$(NC)"
	@echo "$(YELLOW)Services:${NC}"
	@echo "  - Order Service:       localhost:50050"
	@echo "  - Inventory Service:  localhost:50051"
	@echo "  - Payment Service:    localhost:50052"
	@echo "  - Notification Service: localhost:50053"

## stop-all: Stop all running services
stop-all:
	@echo "$(YELLOW)Stopping all services...$(NC)"
	@-pkill -f "order-service" 2>/dev/null || true
	@-pkill -f "inventory-service" 2>/dev/null || true
	@-pkill -f "payment-service" 2>/dev/null || true
	@-pkill -f "notification-service" 2>/dev/null || true
	@tmux kill-session -t grpc-demo 2>/dev/null || true
	@echo "$(GREEN)All services stopped$(NC)"

## install-proto: Install protobuf tools
install-proto:
	$(GOCMD) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOCMD) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

## proto: Generate Go code from proto files
proto: install-proto
	@echo "$(BLUE)Generating Go code from proto files...$(NC)"
	@export PATH="$(PATH):$(shell go env GOPATH)/bin" && \
	for path in $(PROTO_PATHS); do \
		$(PROTOC) --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$$path/*.proto; \
	done
	@echo "$(GREEN)Proto files generated successfully$(NC)"

## tidy: Download and tidy dependencies
tidy:
	$(GOMOD) tidy

## clean: Clean build artifacts
clean:
	@rm -rf bin/
	@$(GOCMD) clean
	@echo "$(GREEN)Cleaned$(NC)"

## test: Run tests
test:
	$(GOCMD) test -v ./...

## lint: Run linter
lint:
	@which golangci-lint > /dev/null || echo "Install golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
	@golangci-lint run ./... || true

## check: Run build and vet
check: build vet

## vet: Run go vet
vet:
	$(GOCMD) vet ./...

## deps: Install dependencies
deps:
	$(GOMOD) download

## tree: Show project structure
tree:
	@find . -type f \( -name "*.go" -o -name "*.proto" -o -name "Makefile" -o -name "*.md" \) | grep -v ".git" | sort

## help: Show this help
help:
	@echo "gRPC Microservices Demo - Makefile"
	@echo ""
	@echo "$(GREEN)Usage:$(NC)"
	@echo "  make $(YELLOW)<target>$(NC)"
	@echo ""
	@echo "$(GREEN)Build targets:$(NC)"
	@echo "  $(YELLOW)build$(NC)              Build all binaries"
	@echo "  $(YELLOW)build-order$(NC)        Build order service"
	@echo "  $(YELLOW)build-inventory$(NC)    Build inventory service"
	@echo "  $(YELLOW)build-payment$(NC)       Build payment service"
	@echo "  $(YELLOW)build-notification$(NC)  Build notification service"
	@echo "  $(YELLOW)build-client$(NC)        Build client"
	@echo ""
	@echo "$(GREEN)Run targets:$(NC)"
	@echo "  $(YELLOW)run-order$(NC)           Run order service"
	@echo "  $(YELLOW)run-inventory$(NC)        Run inventory service"
	@echo "  $(YELLOW)run-payment$(NC)         Run payment service"
	@echo "  $(YELLOW)run-notification$(NC)     Run notification service"
	@echo "  $(YELLOW)run-client$(NC)           Run client"
	@echo "  $(YELLOW)run-all$(NC)             Run all services (tmux)"
	@echo "  $(YELLOW)run-all-bg$(NC)          Run all services (background)"
	@echo "  $(YELLOW)stop-all$(NC)            Stop all services"
	@echo ""
	@echo "$(GREEN)Proto targets:$(NC)"
	@echo "  $(YELLOW)proto$(NC)               Generate Go code from proto files"
	@echo "  $(YELLOW)install-proto$(NC)       Install protoc tools"
	@echo ""
	@echo "$(GREEN)Other targets:$(NC)"
	@echo "  $(YELLOW)deps$(NC)                Download dependencies"
	@echo "  $(YELLOW)tidy$(NC)                Tidy dependencies"
	@echo "  $(YELLOW)test$(NC)                Run tests"
	@echo "  $(YELLOW)vet$(NC)                Run go vet"
	@echo "  $(YELLOW)lint$(NC)               Run linter"
	@echo "  $(YELLOW)check$(NC)              Build and vet"
	@echo "  $(YELLOW)clean$(NC)              Clean build artifacts"
	@echo "  $(YELLOW)tree$(NC)               Show project structure"
	@echo "  $(YELLOW)help$(NC)               Show this help"
