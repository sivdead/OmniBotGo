set dotenv-load := true

# Linux/macOS first — bash only, no Windows PowerShell
set shell := ["bash", "-cu"]

local_bin := justfile_directory() + "/bin"
base_stack := "docker compose -f docker/docker-compose.yml"
integration_test_stack := base_stack + " -f docker/docker-compose-integration-test.yml"
all_stack := integration_test_stack

default:
    @just --list

compose-up:
    cd docker && docker compose up --build -d db rabbitmq && docker compose logs -f

compose-up-all:
    cd docker && docker compose up --build -d

compose-up-integration-test:
    cd docker && docker compose -f docker-compose.yml -f docker-compose-integration-test.yml up --build --abort-on-container-exit --exit-code-from integration-test

compose-down:
    cd docker && docker compose -f docker-compose.yml -f docker-compose-integration-test.yml down --remove-orphans

swag-v1:
    swag init -g internal/controller/http/router.go

proto-v1:
    @echo "Proto generation disabled - proto files removed"

deps:
    go mod tidy && go mod verify

deps-audit:
    govulncheck ./...

format:
    gofumpt -l -w .
    gci write . --skip-generated -s standard -s default

run: deps swag-v1 proto-v1 wire
    go mod download
    CGO_ENABLED=0 go run -tags migrate ./cmd/app

run-dev: deps swag-v1 proto-v1 wire
    go mod download
    LOG_LEVEL=debug SWAGGER_ENABLED=true CGO_ENABLED=0 go run -tags migrate ./cmd/app

run-prod: deps swag-v1 proto-v1 wire
    go mod download
    LOG_LEVEL=info SWAGGER_ENABLED=false CGO_ENABLED=0 go run ./cmd/app

docker-rm-volume:
    docker volume rm go-clean-template_pg-data

linter-golangci:
    golangci-lint run

linter-hadolint:
    git ls-files --exclude='Dockerfile*' --ignored | xargs hadolint

linter-dotenv:
    dotenv-linter

test:
    go test -v -race -covermode atomic -coverprofile=coverage.txt ./internal/...

integration-test:
    go clean -testcache && go test -v ./integration-test/...

wire:
    wire ./internal/app

mock:
    mockgen -source ./internal/repo/contracts.go -package usecase_test > ./internal/usecase/mocks_repo_test.go
    mockgen -source ./internal/usecase/contracts.go -package usecase_test > ./internal/usecase/mocks_usecase_test.go

# Create a new migration in a dialect folder (default mysql). Example:
#   just migrate-create add_foo
#   just migrate-create add_foo postgres
migrate-create name dialect="mysql":
    migrate create -ext sql -dir migrations/{{dialect}} '{{name}}'

# Apply migrations. Set DB_TYPE + database URL. Examples:
#   just migrate-up                                 # uses PG_URL against migrations/postgres if DB_TYPE=postgres
#   DB_TYPE=sqlite DB_URL="sqlite:///tmp/x.db" just migrate-up
migrate-up:
    #!/usr/bin/env bash
    set -euo pipefail
    db_type="${DB_TYPE:-postgres}"
    case "$db_type" in
      mysql) path="migrations/mysql"; url="${DB_URL:-mysql://${MYSQL_DSN}}" ;;
      postgres|postgresql) path="migrations/postgres"; url="${DB_URL:-${PG_URL}}" ;;
      sqlite|sqlite3) path="migrations/sqlite"; url="${DB_URL:?set DB_URL=sqlite:///path/to.db}" ;;
      *) echo "unsupported DB_TYPE=$db_type"; exit 1 ;;
    esac
    migrate -path "$path" -database "$url" up

bin-deps:
    GOBIN={{local_bin}} go install tool

pre-commit: swag-v1 proto-v1 wire mock format linter-golangci test
