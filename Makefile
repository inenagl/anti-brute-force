DB_IMAGE_NAME := "anti_brute_force_db"
DBSTRING := "user=abfuser password=abfpassword dbname=abf host=localhost port=5432"
MIGRATIONS_DIR := "migrations"

BIN := "./bin/abf"
GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

DOCKER_IMG="anti-brute-force:develop"

run-db:
	docker run -d \
		--name $(DB_IMAGE_NAME) \
		-e POSTGRES_PASSWORD=postgres \
		-e PGDATA=/var/lib/postgresql/data/pgdata \
		-v pg_data:/var/lib/postgresql/data \
		-p 5432:5432 \
		postgres
	docker exec -it $(DB_IMAGE_NAME) psql -Upostgres -dpostgres \
    	-c "CREATE DATABASE abf;" \
    	-c "CREATE USER abfuser WITH ENCRYPTED PASSWORD 'abfpassword';" \
    	-c "GRANT ALL ON DATABASE abf TO abfuser;" \
    	-c "ALTER DATABASE abf OWNER TO abfuser;" \
    	-c "GRANT USAGE, CREATE ON SCHEMA PUBLIC TO abfuser;"

stop-db:
	docker stop $(DB_IMAGE_NAME)
	docker rm $(DB_IMAGE_NAME)

install-goose:
	go install github.com/pressly/goose/v3/cmd/goose@latest

migrate:
	goose -dir $(MIGRATIONS_DIR) postgres $(DBSTRING) up

generate:
	protoc --proto_path=./api/ --go_out=./internal/api --go-grpc_out=./internal/api anti-brute-force.proto

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2

lint: install-lint-deps
	golangci-lint run --timeout 5m0s ./...

test:
	go test -v -race --count=2 ./internal/...

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/abf

run: build
	$(BIN) -config ./configs/config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

up:
	docker compose -f deployments/docker-compose.yaml -p anti-brute-force up -d

up-build:
	docker compose -f deployments/docker-compose.yaml -p anti-brute-force up -d --build

down:
	docker compose -f deployments/docker-compose.yaml -p anti-brute-force down

integration-tests-build:
	docker compose -f deployments/docker-compose.test.yaml -p anti-brute-force-test build

integration-tests-build_tests:
	docker compose -f deployments/docker-compose.test.yaml -p anti-brute-force-test build tests

integration-tests:
	docker compose -f deployments/docker-compose.test.yaml -p anti-brute-force-test up --exit-code-from tests --attach tests && \
	EXIT_CODE=$$? &&\
	docker compose -f deployments/docker-compose.test.yaml -p anti-brute-force-test down && \
    echo "command exited with $$EXIT_CODE" && \
    exit $$EXIT_CODE

.PHONY: run_db stop_db install-goose migrate generate install-lint-deps lint test build run build-img run-img version up up-build down integration-tests-build integration-tests-build_tests integration-tests
