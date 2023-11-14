DB_IMAGE_NAME := "anti_brute_force_db"
DBSTRING := "user=abfuser password=abfpassword dbname=abf host=localhost port=5432"
MIGRATIONS_DIR := "migrations"

BIN := "./bin/abf"

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
	go test -race ./internal/...

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/abf

run: build
	$(BIN) -config ./configs/config.yaml


.PHONY: run_db stop_db install-goose migrate generate install-lint-deps lint test build run
