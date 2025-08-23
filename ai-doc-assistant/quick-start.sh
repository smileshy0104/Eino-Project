#!/bin/bash

# AI文档助手快速启动脚本
# 简化版本，快速体验

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════╗"
echo "║         AI文档助手 - 快速启动               ║"
echo "║         基于Eino框架 + 火山方舟            ║"
echo "╚════════════════════════════════════════════╝"
echo -e "${NC}"

# 1. 检查基本依赖
echo -e "${GREEN}[1/5]${NC} 检查依赖..."
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}请先安装Docker: https://docs.docker.com/get-docker/${NC}"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}请先安装Go 1.19+: https://golang.org/dl/${NC}"
    exit 1
fi

# 2. 创建目录
echo -e "${GREEN}[2/5]${NC} 创建项目目录..."
mkdir -p {uploads,logs,data}

# 3. 启动基础服务
echo -e "${GREEN}[3/5]${NC} 启动数据库服务..."
docker-compose up -d mysql redis milvus-standalone

echo -e "${GREEN}等待服务启动...${NC}"
sleep 20

# 4. 初始化数据库
echo -e "${GREEN}[4/5]${NC} 初始化数据库..."
until docker exec ai-assistant-mysql mysqladmin ping -h"localhost" --silent 2>/dev/null; do
    echo -n "."
    sleep 2
done

docker exec -i ai-assistant-mysql mysql -uai_user -pai_password ai_assistant < scripts/init.sql

# 5. 提示配置API密钥
echo -e "${GREEN}[5/5]${NC} 配置检查..."
if grep -q "your-volcengine-api-key-here" config/app.yaml 2>/dev/null || [ ! -f config/app.yaml ]; then
    echo -e "${YELLOW}⚠️  请配置您的火山方舟API密钥:${NC}"
    echo "1. 编辑配置文件: vim config/app.yaml"
    echo "2. 设置 ai.api_key 字段"
    echo "3. 然后运行: go run cmd/server/main.go"
else
    echo -e "${GREEN}✅ 配置文件已就绪${NC}"
fi

echo ""
echo -e "${GREEN}🎉 快速启动完成！${NC}"
echo ""
echo -e "${BLUE}下一步:${NC}"
echo "1. 配置API密钥: vim config/app.yaml"
echo "2. 启动应用: make dev 或 go run cmd/server/main.go"
echo "3. 访问: http://localhost:8080"
echo ""
echo -e "${BLUE}完整命令:${NC}"
echo "• make help      - 查看所有命令"
echo "• make install   - 完整部署"
echo "• make start     - 启动服务"
echo "• make stop      - 停止服务"