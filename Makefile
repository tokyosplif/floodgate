DOCKER_COMPOSE = docker-compose.yaml
GATEWAY_BINARY = bin/gateway
PROCESSOR_BINARY = bin/processor

.PHONY: all up down build run-gateway run-processor migrate help clean

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Infrastructure Targets:"
	@echo "  up             - Spin up Docker containers (Kafka, Redis, ClickHouse)"
	@echo "  down           - Stop and remove all containers"
	@echo "  migrate        - Initialize ClickHouse database schema"
	@echo ""
	@echo "Development Targets:"
	@echo "  build          - Compile both Gateway and Processor binaries"
	@echo "  run-gateway    - Run the Gateway service locally"
	@echo "  run-processor  - Run the Processor service locally"
	@echo "  clean          - Remove compiled binaries and temporary files"

up:
	docker-compose -f $(DOCKER_COMPOSE) up -d
	@echo "Waiting for services to be healthy..."
	@sleep 5

down:
	docker-compose -f $(DOCKER_COMPOSE) down

migrate:
	docker exec -i $$(docker ps -qf "name=clickhouse") clickhouse-client -q "$$(cat scripts/clickhouse/init.sql)"
	@echo "ClickHouse migrations applied successfully!"

build: clean
	@echo "Building binaries..."
	go build -o $(GATEWAY_BINARY) ./cmd/gateway/main.go
	go build -o $(PROCESSOR_BINARY) ./cmd/processor/main.go
	@echo "Build complete! Check the /bin directory."

run-gateway:
	go run cmd/gateway/main.go

run-processor:
	go run cmd/processor/main.go

clean:
	rm -rf bin/