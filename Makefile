# Bobber the SWE - Makefile
# Go-based job processing pipeline with hot reloading and development tools

# Variables
BINARY_NAME=bobber
MAIN_PATH=./cmd/main/main.go
BUILD_DIR=./bin
VERSION?=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=${VERSION}"

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Docker
DOCKER_COMPOSE_INFRA=infra/docker-compose.yml
DOCKER_COMPOSE_OBSERVABILITY=infra/docker-compose.observability.yml

.PHONY: all build clean test deps dev hot-reload \
        infra-up infra-down infra-logs infra-clean \
        obs-up obs-down obs-logs \
        docker-build docker-run \
        help

# Default target
all: clean deps build test

## Development Commands

# Build the application
build:
	@echo "${GREEN}Building ${BINARY_NAME}...${NC}"
	@mkdir -p ${BUILD_DIR}
	${GOBUILD} ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ${MAIN_PATH}
	@echo "${GREEN}Build complete: ${BUILD_DIR}/${BINARY_NAME}${NC}"

# Clean build artifacts
clean:
	@echo "${YELLOW}Cleaning...${NC}"
	${GOCLEAN}
	@rm -rf ${BUILD_DIR}
	@echo "${GREEN}Clean complete${NC}"

# Run tests
test:
	@echo "${BLUE}Running tests...${NC}"
	${GOTEST} -v ./...

# Download dependencies
deps:
	@echo "${BLUE}Downloading dependencies...${NC}"
	${GOMOD} download
	${GOMOD} tidy

# Run with debug logging
run-debug:
	@echo "${GREEN}Starting Bobber the SWE (debug mode)...${NC}"
	DEBUG=true LOG_LEVEL=debug ${GOCMD} run ${MAIN_PATH}

# Development mode with automatic rebuilding and restarting (requires air)
hot-reload:
	@echo "${GREEN}Starting hot reload development server...${NC}"
	@echo "${YELLOW}Make sure infrastructure is running: make infra-up${NC}"
	air

## Infrastructure Management

# Start all infrastructure services
infra-up:
	@echo "${GREEN}Starting infrastructure services...${NC}"
	docker compose -f ${DOCKER_COMPOSE_INFRA} up -d
	@echo "${GREEN}Infrastructure services started${NC}"
	@echo "${BLUE}PostgreSQL: localhost:5432 (postgres/postgres)${NC}"
	@echo "${BLUE}Redis: localhost:6379${NC}"

# Stop infrastructure services
infra-down:
	@echo "${YELLOW}Stopping infrastructure services...${NC}"
	docker compose -f ${DOCKER_COMPOSE_INFRA} down

# Show infrastructure logs
infra-logs:
	docker compose -f ${DOCKER_COMPOSE_INFRA} logs -f

# Clean infrastructure (remove volumes)
infra-clean:
	@echo "${RED}Cleaning infrastructure (removing volumes)...${NC}"
	docker compose -f ${DOCKER_COMPOSE_INFRA} down -v
	docker system prune -f

# Start observability services (Prometheus, Grafana)
obs-up:
	@echo "${GREEN}Starting observability services...${NC}"
	docker compose -f ${DOCKER_COMPOSE_OBSERVABILITY} up -d
	@echo "${GREEN}Observability services started${NC}"
	@echo "${BLUE}Prometheus: http://localhost:9090${NC}"

# Stop observability services
obs-down:
	@echo "${YELLOW}Stopping observability services...${NC}"
	docker compose -f ${DOCKER_COMPOSE_OBSERVABILITY} down

# Show observability logs
obs-logs:
	docker compose -f ${DOCKER_COMPOSE_OBSERVABILITY} logs -f

## Production

# Build for production
build-prod:
	@echo "${GREEN}Building for production...${NC}"
	@mkdir -p ${BUILD_DIR}
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ${GOBUILD} ${LDFLAGS} -a -installsuffix cgo -o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64 ${MAIN_PATH}
	@echo "${GREEN}Production build complete: ${BUILD_DIR}/${BINARY_NAME}-linux-amd64${NC}"

# Build Docker image
docker-build:
	@echo "${GREEN}Building Docker image...${NC}"
	docker build -t ${BINARY_NAME}:${VERSION} -t ${BINARY_NAME}:latest .

# Run Docker container
docker-run:
	@echo "${GREEN}Running Docker container...${NC}"
	docker run --rm -p 8080:8080 --network host ${BINARY_NAME}:latest

## Development Workflow

# Stop all services
dev-down:
	@make infra-down
	@make obs-down


# Help
help:
	@echo "${GREEN}Bobber the SWE - Available Commands${NC}"
	@echo ""
	@echo "${BLUE}Development:${NC}"
	@echo "  build          - Build the application"
	@echo "  run-debug      - Run with debug logging"
	@echo "  hot-reload     - Development mode with auto-restart"
	@echo "  test           - Run tests"
	@echo ""
	@echo "${BLUE}Infrastructure:${NC}"
	@echo "  infra-up       - Start PostgreSQL and Redis"
	@echo "  infra-down     - Stop infrastructure services"
	@echo "  infra-logs     - Show infrastructure logs"
	@echo "  obs-up         - Start monitoring (Prometheus)"
	@echo "  obs-down       - Stop monitoring services"
	@echo ""
	@echo "${BLUE}Production:${NC}"
	@echo "  build-prod     - Build for production"
	@echo "  docker-build   - Build Docker image"
	@echo ""
	@echo "${BLUE}Utilities:${NC}"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download dependencies"
	@echo "  help           - Show this help" 