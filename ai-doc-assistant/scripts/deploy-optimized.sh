#!/bin/bash

# AI文档助手 - 优化部署脚本
# 复用现有Docker环境，最小化资源占用

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

echo "=========================================="
echo "🚀 AI文档助手 - 优化部署"
echo "复用现有Docker环境，最小化资源占用"
echo "=========================================="

log_step "1. 检查现有环境..."

# 检查现有Milvus服务
if docker ps | grep -q milvus-standalone; then
    log_info "✅ 检测到现有Milvus服务正在运行"
    MILVUS_STATUS="running"
else
    log_warn "⚠️  Milvus服务未运行，需要先启动"
    MILVUS_STATUS="stopped"
fi

# 检查现有网络
if docker network ls | grep -q eino_default; then
    log_info "✅ 检测到现有eino_default网络"
    NETWORK_EXISTS="yes"
else
    log_warn "⚠️  eino_default网络不存在，将创建"
    NETWORK_EXISTS="no"
fi

log_step "2. 环境准备..."

# 创建必要目录
mkdir -p logs uploads backups
log_info "✅ 目录创建完成"

# 创建网络（如果不存在）
if [ "$NETWORK_EXISTS" = "no" ]; then
    docker network create eino_default
    log_info "✅ 创建eino_default网络"
fi

log_step "3. 构建应用镜像..."
docker-compose -f docker-compose.optimized.yml build ai-doc-assistant
log_info "✅ 应用镜像构建完成"

log_step "4. 启动服务..."

# 如果Milvus未运行，给出提示
if [ "$MILVUS_STATUS" = "stopped" ]; then
    log_warn "请先启动Milvus相关服务："
    log_warn "  docker-compose -f /path/to/milvus/docker-compose.yml up -d"
    echo ""
    read -p "Milvus服务已启动？继续部署 (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_error "部署中止"
        exit 1
    fi
fi

# 启动服务
docker-compose -f docker-compose.optimized.yml up -d
log_info "✅ 服务启动完成"

log_step "5. 等待服务就绪..."
sleep 15

log_step "6. 健康检查..."

# 检查服务状态
echo "服务状态："
docker-compose -f docker-compose.optimized.yml ps

echo ""
echo "网络连接测试："

# 测试MySQL连接
if docker exec ai-assistant-mysql mysqladmin ping -h"localhost" --silent 2>/dev/null; then
    log_info "✅ MySQL 连接正常"
else
    log_warn "⚠️  MySQL 连接异常"
fi

# 测试Redis连接
if docker exec ai-assistant-redis redis-cli ping | grep -q PONG 2>/dev/null; then
    log_info "✅ Redis 连接正常"
else
    log_warn "⚠️  Redis 连接异常"
fi

# 测试Milvus连接
if curl -f http://localhost:9091/healthz >/dev/null 2>&1; then
    log_info "✅ Milvus 连接正常"
else
    log_warn "⚠️  Milvus 连接异常"
fi

# 测试应用服务
sleep 5
if curl -f http://localhost:8080/health >/dev/null 2>&1; then
    log_info "✅ AI文档助手 服务正常"
else
    log_warn "⚠️  AI文档助手 服务异常，检查日志："
    docker logs ai-doc-assistant-app --tail 10
fi

echo ""
echo "=========================================="
log_info "🎉 部署完成！"
echo "=========================================="

echo ""
echo "🌟 服务访问地址："
echo "  • 应用主页: http://localhost:8080"
echo "  • Web界面: http://localhost:8081"
echo "  • API文档: http://localhost:8080/swagger/index.html"
echo "  • 健康检查: http://localhost:8080/health"
echo "  • Milvus管理: http://localhost:8001 (Attu)"

echo ""
echo "🔧 管理命令："
echo "  • 查看日志: docker-compose -f docker-compose.optimized.yml logs -f"
echo "  • 停止服务: docker-compose -f docker-compose.optimized.yml down"
echo "  • 重启服务: docker-compose -f docker-compose.optimized.yml restart"
echo "  • 查看状态: docker-compose -f docker-compose.optimized.yml ps"

echo ""
echo "📊 资源使用："
echo "  • 复用现有Milvus服务 ✅"
echo "  • 新增MySQL: ai-assistant-mysql:3307"
echo "  • 新增Redis: ai-assistant-redis:6380" 
echo "  • 新增应用: ai-doc-assistant-app:8080"
echo "  • 新增Web: ai-assistant-web:8081"

echo ""
log_info "部署优化完成！现在可以开始使用AI文档助手了。"