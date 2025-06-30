# Makefile for 统计服务

.PHONY: help build run test clean docker-build docker-run docker-stop

# 默认目标
help:
	@echo "可用的命令："
	@echo "  build        - 构建应用"
	@echo "  run          - 运行应用"
	@echo "  test         - 运行测试"
	@echo "  clean        - 清理构建文件"
	@echo "  docker-build - 构建 Docker 镜像"
	@echo "  docker-run   - 启动 Docker 服务"
	@echo "  docker-stop  - 停止 Docker 服务"
	@echo "  docker-logs  - 查看 Docker 日志"

# 构建应用
build:
	@echo "构建应用..."
	go build -o bin/stat ./cmd

# 运行应用
run:
	@echo "运行应用..."
	go run ./cmd

# 运行测试
test:
	@echo "运行测试..."
	go test ./...

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf bin/
	go clean

# 构建 Docker 镜像
docker-build:
	@echo "构建 Docker 镜像..."
	docker build -t stat-service .

# 启动 Docker 服务
docker-run:
	@echo "启动 Docker 服务..."
	docker-compose up -d

# 停止 Docker 服务
docker-stop:
	@echo "停止 Docker 服务..."
	docker-compose down

# 查看 Docker 日志
docker-logs:
	@echo "查看 Docker 日志..."
	docker-compose logs -f

# 重启 Docker 服务
docker-restart: docker-stop docker-run

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
lint:
	@echo "代码检查..."
	golangci-lint run

# 安装依赖
deps:
	@echo "安装依赖..."
	go mod download
	go mod tidy

# 生成文档
docs:
	@echo "生成文档..."
	godoc -http=:6060

# 性能测试
bench:
	@echo "性能测试..."
	go test -bench=. ./...

# 覆盖率测试
cover:
	@echo "覆盖率测试..."
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html 