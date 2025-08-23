# AI文档助手 Makefile
# 提供便捷的开发和部署命令

.PHONY: help dev build test clean install start stop restart logs status deps

# 默认目标
.DEFAULT_GOAL := help

# 项目变量
APP_NAME = ai-doc-assistant
VERSION = 1.0.0
BUILD_TIME = $(shell date +%Y%m%d_%H%M%S)
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_FLAGS = -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Go相关变量
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
BINARY_NAME = $(APP_NAME)
BINARY_PATH = bin/$(BINARY_NAME)
MAIN_FILE = cmd/server/main.go

# Docker相关变量
DOCKER_IMAGE = $(APP_NAME):$(VERSION)
DOCKER_COMPOSE = docker-compose

# 颜色定义
GREEN = \033[0;32m
YELLOW = \033[1;33m
RED = \033[0;31m
NC = \033[0m # No Color

## help: 显示帮助信息
help:
	@echo "AI文档助手 - 可用命令:"
	@echo
	@echo "  $(GREEN)开发命令:$(NC)"
	@echo "    dev        - 开发模式启动 (热重载)"
	@echo "    build      - 构建应用程序"
	@echo "    test       - 运行测试"
	@echo "    clean      - 清理构建文件"
	@echo
	@echo "  $(GREEN)部署命令:$(NC)"
	@echo "    install    - 一键部署 (完整流程)"
	@echo "    start      - 启动所有服务"
	@echo "    stop       - 停止所有服务"
	@echo "    restart    - 重启所有服务"
	@echo
	@echo "  $(GREEN)运维命令:$(NC)"
	@echo "    logs       - 查看应用日志"
	@echo "    status     - 检查服务状态"
	@echo "    health     - 健康检查"
	@echo
	@echo "  $(GREEN)工具命令:$(NC)"
	@echo "    deps       - 安装/更新依赖"
	@echo "    format     - 格式化代码"
	@echo "    lint       - 代码检查"
	@echo

## dev: 开发模式启动
dev:
	@echo "$(GREEN)启动开发模式...$(NC)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(YELLOW)air未安装，使用go run启动...$(NC)"; \
		echo "$(YELLOW)推荐安装air实现热重载: go install github.com/cosmtrek/air@latest$(NC)"; \
		$(GOCMD) run $(MAIN_FILE); \
	fi

## build: 构建应用程序
build:
	@echo "$(GREEN)构建应用程序...$(NC)"
	@mkdir -p bin
	@$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_PATH) $(MAIN_FILE)
	@echo "$(GREEN)✅ 构建完成: $(BINARY_PATH)$(NC)"

## test: 运行测试
test:
	@echo "$(GREEN)运行测试...$(NC)"
	@$(GOTEST) -v ./...

## test-cover: 运行测试并生成覆盖率报告
test-cover:
	@echo "$(GREEN)运行测试并生成覆盖率报告...$(NC)"
	@$(GOTEST) -v -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ 覆盖率报告生成: coverage.html$(NC)"

## clean: 清理构建文件
clean:
	@echo "$(GREEN)清理构建文件...$(NC)"
	@$(GOCLEAN)
	@rm -rf bin/
	@rm -rf logs/*.log
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✅ 清理完成$(NC)"

## install: 一键部署
install:
	@echo "$(GREEN)开始一键部署...$(NC)"
	@chmod +x scripts/setup.sh
	@./scripts/setup.sh

## start: 启动所有服务
start:
	@echo "$(GREEN)启动所有服务...$(NC)"
	@$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)等待服务启动...$(NC)"
	@sleep 10
	@make status

## stop: 停止所有服务  
stop:
	@echo "$(GREEN)停止所有服务...$(NC)"
	@$(DOCKER_COMPOSE) down
	@if [ -f .app.pid ]; then \
		PID=$$(cat .app.pid); \
		if ps -p $$PID > /dev/null; then \
			kill $$PID; \
			echo "$(GREEN)✅ 应用进程已停止 (PID: $$PID)$(NC)"; \
		fi; \
		rm -f .app.pid; \
	fi

## restart: 重启所有服务
restart: stop start

## logs: 查看应用日志
logs:
	@echo "$(GREEN)显示应用日志...$(NC)"
	@if [ -f logs/app.log ]; then \
		tail -f logs/app.log; \
	else \
		echo "$(YELLOW)日志文件不存在: logs/app.log$(NC)"; \
	fi

## logs-docker: 查看Docker服务日志
logs-docker:
	@echo "$(GREEN)显示Docker服务日志...$(NC)"
	@$(DOCKER_COMPOSE) logs -f

## status: 检查服务状态
status:
	@echo "$(GREEN)检查服务状态...$(NC)"
	@echo
	@echo "$(YELLOW)Docker服务状态:$(NC)"
	@$(DOCKER_COMPOSE) ps
	@echo
	@echo "$(YELLOW)应用服务状态:$(NC)"
	@if [ -f .app.pid ]; then \
		PID=$$(cat .app.pid); \
		if ps -p $$PID > /dev/null; then \
			echo "✅ 应用服务运行中 (PID: $$PID)"; \
		else \
			echo "❌ 应用服务未运行"; \
		fi; \
	else \
		echo "❌ 应用服务未启动"; \
	fi

## health: 健康检查
health:
	@echo "$(GREEN)执行健康检查...$(NC)"
	@echo
	@echo "$(YELLOW)检查MySQL连接...$(NC)"
	@if docker exec ai-assistant-mysql mysqladmin ping -h"localhost" --silent 2>/dev/null; then \
		echo "✅ MySQL 连接正常"; \
	else \
		echo "❌ MySQL 连接失败"; \
	fi
	@echo
	@echo "$(YELLOW)检查Redis连接...$(NC)"
	@if docker exec ai-assistant-redis redis-cli ping | grep -q PONG 2>/dev/null; then \
		echo "✅ Redis 连接正常"; \
	else \
		echo "❌ Redis 连接失败"; \
	fi
	@echo
	@echo "$(YELLOW)检查Milvus连接...$(NC)"
	@if curl -f http://localhost:9091/healthz >/dev/null 2>&1; then \
		echo "✅ Milvus 连接正常"; \
	else \
		echo "❌ Milvus 连接失败"; \
	fi
	@echo
	@echo "$(YELLOW)检查应用服务...$(NC)"
	@if curl -f http://localhost:8080/health >/dev/null 2>&1; then \
		echo "✅ 应用服务正常"; \
	else \
		echo "❌ 应用服务异常"; \
	fi

## deps: 安装/更新依赖
deps:
	@echo "$(GREEN)安装/更新Go依赖...$(NC)"
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "$(GREEN)✅ 依赖更新完成$(NC)"

## format: 格式化代码
format:
	@echo "$(GREEN)格式化Go代码...$(NC)"
	@$(GOCMD) fmt ./...
	@echo "$(GREEN)✅ 代码格式化完成$(NC)"

## lint: 代码检查
lint:
	@echo "$(GREEN)执行代码检查...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint未安装，使用go vet...$(NC)"; \
		$(GOCMD) vet ./...; \
	fi

## mock-data: 生成模拟数据
mock-data:
	@echo "$(GREEN)生成模拟数据...$(NC)"
	@$(GOCMD) run scripts/mock_data.go
	@echo "$(GREEN)✅ 模拟数据生成完成$(NC)"

## db-migrate: 数据库迁移
db-migrate:
	@echo "$(GREEN)执行数据库迁移...$(NC)"
	@docker exec -i ai-assistant-mysql mysql -uai_user -pai_password ai_assistant < scripts/init.sql
	@echo "$(GREEN)✅ 数据库迁移完成$(NC)"

## db-backup: 数据库备份
db-backup:
	@echo "$(GREEN)备份数据库...$(NC)"
	@mkdir -p backups
	@docker exec ai-assistant-mysql mysqldump -uai_user -pai_password ai_assistant > backups/backup_$(BUILD_TIME).sql
	@echo "$(GREEN)✅ 数据库备份完成: backups/backup_$(BUILD_TIME).sql$(NC)"

## docker-build: 构建Docker镜像
docker-build:
	@echo "$(GREEN)构建Docker镜像...$(NC)"
	@docker build -t $(DOCKER_IMAGE) -f docker/Dockerfile .
	@echo "$(GREEN)✅ Docker镜像构建完成: $(DOCKER_IMAGE)$(NC)"

## docker-push: 推送Docker镜像
docker-push: docker-build
	@echo "$(GREEN)推送Docker镜像...$(NC)"
	@docker push $(DOCKER_IMAGE)
	@echo "$(GREEN)✅ Docker镜像推送完成$(NC)"

## benchmark: 性能测试
benchmark:
	@echo "$(GREEN)运行性能测试...$(NC)"
	@$(GOTEST) -bench=. -benchmem ./...

## security-scan: 安全扫描
security-scan:
	@echo "$(GREEN)执行安全扫描...$(NC)"
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)gosec未安装，跳过安全扫描$(NC)"; \
		echo "$(YELLOW)安装命令: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest$(NC)"; \
	fi

## update-deps: 更新所有依赖到最新版本
update-deps:
	@echo "$(GREEN)更新所有依赖到最新版本...$(NC)"
	@$(GOCMD) get -u ./...
	@$(GOMOD) tidy
	@echo "$(GREEN)✅ 依赖更新完成$(NC)"

## info: 显示项目信息
info:
	@echo "$(GREEN)项目信息:$(NC)"
	@echo "  应用名称: $(APP_NAME)"
	@echo "  版本号: $(VERSION)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  Git提交: $(GIT_COMMIT)"
	@echo "  Go版本: $(shell $(GOCMD) version)"
	@echo "  Docker版本: $(shell docker --version 2>/dev/null || echo '未安装')"