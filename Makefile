swag:
	@swag init -g cmd/cm-ant/main.go --output api/

run: rundb
	@go run cmd/cm-ant/main.go

build:
	@go build -o ant ./cmd/cm-ant/...


OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

ifeq ($(ARCH),x86_64)
    ARCH := amd64
else ifeq ($(ARCH),arm64)
    ARCH := arm64
else ifeq ($(ARCH),aarch64)
    ARCH := arm64
endif

docker-build:
	@docker image build --build-arg TARGETOS=$(OS) --build-arg TARGETARCH=$(ARCH) --tag cm-ant --file Dockerfile .

DB_CONTAINER_NAME=ant-postgres
DB_NAME=cm-ant-db

check-db:
	@docker ps -q -f name=$(DB_CONTAINER_NAME)

run-db:
	@echo "Checking if the database container is already running..."
	@if [ -z "$$(make check-db)" ]; then \
		echo "Run database container...."; \
		docker container run \
			--name $(DB_CONTAINER_NAME) \
			-p 5432:5432 \
			-e POSTGRES_USER=cm-ant-user \
			-e POSTGRES_PASSWORD=cm-ant-secret \
			-d --rm \
			postgres:16.2-alpine3.19; \
		echo "Started Postgres database container!"; \
		docker exec -it $(DB_CONTAINER_NAME) createdb --username=cm-ant-user --owner=cm-ant-user $(DB_NAME); \
		echo "database $(DB_NAME) successfully started!"
	else \
		echo "Database container is already running."; \
	fi

.PHONY: swag run build docker-build check-db run-db