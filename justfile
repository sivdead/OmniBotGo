# 导入环境变量
set dotenv-load := true

# 设置 Windows 下的 shell
set shell := ["pwsh.exe", "-c"]

# 变量定义
local_bin := justfile_directory() + "/bin"
base_stack := "docker compose -f docker-compose.yml"
integration_test_stack := base_stack + " -f docker-compose-integration-test.yml"
all_stack := integration_test_stack

# 默认命令 - 显示帮助
default:
    @just --list

# 运行 docker compose (不包含后端和反向代理)
compose-up:
    {{base_stack}} up --build -d db rabbitmq && docker compose logs -f

# 运行 docker compose (包含后端和反向代理)
compose-up-all:
    {{base_stack}} up --build -d

# 运行 docker compose 并进行集成测试
compose-up-integration-test:
    {{integration_test_stack}} up --build --abort-on-container-exit --exit-code-from integration-test

# 停止 docker compose
compose-down:
    {{all_stack}} down --remove-orphans

# 生成 swagger 文档
swag-v1:
    swag init -g internal/controller/http/router.go

# 从 proto 文件生成源代码
proto-v1:
    protoc --go_out=. \
        --go_opt=paths=source_relative \
        --go-grpc_out=. \
        --go-grpc_opt=paths=source_relative \
        docs/proto/v1/*.proto

# 整理和验证依赖
deps:
    go mod tidy && go mod verify

# 检查依赖漏洞
deps-audit:
    govulncheck ./...

# 运行代码格式化
format:
    gofumpt -l -w .
    gci write . --skip-generated -s standard -s default

# 运行应用 (依赖于 deps, swag-v1, proto-v1, wire)
run: deps swag-v1 proto-v1 wire
    go mod download && \
    CGO_ENABLED=0 go run -tags migrate ./cmd/app

# 删除 docker volume
docker-rm-volume:
    docker volume rm go-clean-template_pg-data

# golangci lint 检查
linter-golangci:
    golangci-lint run

# hadolint lint 检查
linter-hadolint:
    git ls-files --exclude='Dockerfile*' --ignored | xargs hadolint

# dotenv lint 检查
linter-dotenv:
    dotenv-linter

# 运行测试
test:
    go test -v -race -covermode atomic -coverprofile=coverage.txt ./internal/...

# 运行集成测试
integration-test:
    go clean -testcache && go test -v ./integration-test/...

# 生成 Wire 依赖注入代码
wire:
    wire ./internal/app

# 生成 mock 文件
mock:
    mockgen -source ./internal/repo/contracts.go -package usecase_test > ./internal/usecase/mocks_repo_test.go
    mockgen -source ./internal/usecase/contracts.go -package usecase_test > ./internal/usecase/mocks_usecase_test.go

# 创建新的数据库迁移文件
migrate-create name:
    migrate create -ext sql -dir migrations '{{name}}'

# 执行数据库迁移
migrate-up:
    migrate -path migrations -database '$PG_URL?sslmode=disable' up

# 安装工具依赖
bin-deps:
    GOBIN={{local_bin}} go install tool

# 预提交检查 (依赖于多个任务)
pre-commit: swag-v1 proto-v1 wire mock format linter-golangci test 