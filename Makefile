###########################################################
ANT_NETWORK=cm-ant-local-net
DB_CONTAINER_NAME=ant-local-postgres
DB_NAME=cm-ant-db
DB_USER=cm-ant-user 
DB_PASSWORD=cm-ant-secret

ANT_CONTAINER_NAME=cm-ant
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

ifeq ($(ARCH),x86_64)
    ARCH := amd64
else ifeq ($(ARCH),arm64)
    ARCH := arm64
else ifeq ($(ARCH),aarch64)
    ARCH := arm64
endif
###########################################################

###########################################################
.PHONY: swag
swag:
	@swag init -g cmd/cm-ant/main.go --output api/
###########################################################

###########################################################
.PHONY: build
build:
	@go build -o ant ./cmd/cm-ant/...
###########################################################

###########################################################
# Standalone docker-compose dev stack (cm-ant + minimal dependencies).
# See docs/standalone-dev-environment.md for the full workflow.

# Guard: the docker-compose stack reads .env (SPIDER/TB/TERRARIUM
# credentials, etc.). .env is gitignored (it may hold secrets), so it is
# NOT created automatically. When missing, print how to create it from
# .env.example, then stop — the stack is not started.
.PHONY: env
env: ## Check that .env exists for the docker-compose stack
	@if [ -f .env ]; then \
		echo ".env is present."; \
	else \
		echo "ERROR: .env not found."; \
		echo "The docker-compose stack needs a .env file. Create and configure it:"; \
		echo ""; \
		echo "    cp .env.example .env"; \
		echo ""; \
		echo "Edit the values (SPIDER/TB/TERRARIUM credentials, etc.), then run 'make up' again."; \
		exit 1; \
	fi

.PHONY: up
up: env ## Build the cm-ant image from source and start the stack (docker compose up -d --build)
	@docker compose up -d --build

.PHONY: compose-down
compose-down: ## Stop and remove the docker-compose stack (docker compose down)
	@docker compose down

.PHONY: logs
logs: ## Follow docker-compose logs (docker compose logs -f)
	@docker compose logs -f

.PHONY: ps
ps: ## Show docker-compose service status (docker compose ps)
	@docker compose ps
###########################################################

###########################################################
.PHONY: run
run: run-db
	@go run cmd/cm-ant/main.go
###########################################################

###########################################################
.PHONY: create-network
create-network:
	@if [ -z "$$(docker network ls -q -f name=$(ANT_NETWORK))" ]; then \
		echo "Creating cm-ant network..."; \
		docker network create --driver bridge $(ANT_NETWORK); \
		echo "cm-ant network created!"; \
	else \
		echo "cm-ant network already exist..."; \
	fi
###########################################################

###########################################################
.PHONY: run-db
run-db: create-network
	@if [ -z "$$(docker container ps -q -f name=$(DB_CONTAINER_NAME))" ]; then \
		echo "Run database container...."; \
		docker container run \
			--name $(DB_CONTAINER_NAME) \
			--network $(ANT_NETWORK) \
			-p 5432:5432 \
			-e POSTGRES_USER=$(DB_USER) \
			-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
			-e POSTGRES_DB=$(DB_NAME) \
			-d --rm \
			timescale/timescaledb:latest-pg16; \
		echo "Started Postgres database container!"; \
		echo "Waiting for database to be ready..."; \
		for i in $$(seq 1 10); do \
			docker container exec $(DB_CONTAINER_NAME) pg_isready -U $(DB_USER) -d $(DB_NAME); \
			if [ $$? -eq 0 ]; then \
				echo "Database is ready!"; \
				break; \
			fi; \
			echo "Database is not ready yet. Waiting..."; \
			sleep 5; \
		done; \
		if [ $$i -eq 10 ]; then \
			echo "Failed to start the database"; \
			exit 1; \
		fi; \
		echo "Database $(DB_NAME) successfully started!"; \
	else \
		echo "Database container is already running."; \
	fi
###########################################################

###########################################################
.PHONY: down
down:
	@echo "Checking if the database container is running..."
	@if [ -n "$$(docker container ps -q -f name=$(DB_CONTAINER_NAME))" ]; then \
		echo "Stopping and removing the database container..."; \
		docker container stop $(DB_CONTAINER_NAME); \
		echo "Database container stopped!"; \
	else \
		echo "No running database container found."; \
	fi
###########################################################