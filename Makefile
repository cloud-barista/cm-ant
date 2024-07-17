###########################################################
ANT_NETWORK=cm-ant-network
DB_CONTAINER_NAME=ant-postgres
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
.PHONY: run 
run: run-db
	@go run cmd/cm-ant/main.go
###########################################################

###########################################################
.PHONY: build
build:
	@go build -o ant ./cmd/cm-ant/...
###########################################################

###########################################################
.PHONY: create-network
create-network:
	@if [ -z "$$(docker network ls -f name=$(ANT_NETWORK))" ]; then \
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
	@if [ -z "$$(docker ps -q -f name=$(DB_CONTAINER_NAME))" ]; then \
		echo "Run database container...."; \
		docker container run \
			--name $(DB_CONTAINER_NAME) \
			--network $(ANT_NETWORK) \
			-p 5432:5432 \
			-e POSTGRES_USER=$(DB_USER) \
			-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
			-e POSTGRES_DB=$(DB_NAME) \
			-d --rm \
			postgres:16.2-alpine3.19; \
		echo "Started Postgres database container!"; \
		echo "Waiting for database to be ready..."; \
		for i in $$(seq 1 10); do \
			docker exec $(DB_CONTAINER_NAME) pg_isready -U $(DB_USER) -d $(DB_NAME); \
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
down: down-container image-remove
	@echo "Checking if the database container is running..."
	@if [ -n "$$(docker ps -q -f name=$(DB_CONTAINER_NAME))" ]; then \
		echo "Stopping and removing the database container..."; \
		docker container stop $(DB_CONTAINER_NAME); \
		echo "Database container stopped!"; \
	else \
		echo "No running database container found."; \
	fi
###########################################################

###########################################################
.PHONY: image-remove
image-remove:
	@if [ -n "$$(docker images -q $(ANT_CONTAINER_NAME):latest)" ]; then \
		echo "Image $(ANT_CONTAINER_NAME):latest exists. Removing now..."; \
		docker image rm $(ANT_CONTAINER_NAME):latest; \
	else \
		echo "Image $(ANT_CONTAINER_NAME):latest does not exist. Skipping removal."; \
	fi
###########################################################

###########################################################
.PHONY: image-build
image-build:
	@if [ -z "$$(docker images -q cm-ant:latest)" ]; then \
		echo "Image $(ANT_CONTAINER_NAME):latest does not exist. Building now..."; \
		docker image build --build-arg TARGETOS=$(OS) --build-arg TARGETARCH=$(ARCH) --tag $(ANT_CONTAINER_NAME):latest --file Dockerfile .; \
	else \
		echo "Image $(ANT_CONTAINER_NAME):latest already exists. Skipping build."; \
	fi
###########################################################

###########################################################
.PHONY: up
up: run-db image-build
	@if [ -z "$$(docker ps -q -f name=$(ANT_CONTAINER_NAME))" ]; then \
		echo "Run cm-ant application container...."; \
		docker container run \
			--name $(ANT_CONTAINER_NAME) \
			--network $(ANT_NETWORK) \
			-p 8880:8880 \
			-e ANT_DATABASE_HOST=ant-postgres \
			-d --rm \
			$(ANT_CONTAINER_NAME):latest; \
		echo "Started cm-ant application container!"; \
	else \
		echo "cm-ant application container is already running."; \
	fi
###########################################################

###########################################################
.PHONY: down-container
down-container: 
	@echo "Checking if the cm-ant application container is running..."
	@if [ -n "$$(docker ps -q -f name=$(ANT_CONTAINER_NAME))" ]; then \
		echo "Stopping and removing the cm-ant application container..."; \
		docker container stop $(ANT_CONTAINER_NAME); \
		echo "cm-ant application container stopped!"; \
	else \
		echo "No running cm-ant application container found."; \
	fi
###########################################################