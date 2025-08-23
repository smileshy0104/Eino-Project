#!/bin/bash

# AI文档助手一键部署脚本
# 作者: AI Assistant
# 版本: 1.0

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 日志函数
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

# 检查系统要求
check_requirements() {
    log_step "检查系统依赖..."
    
    # 检查操作系统
    if [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
        log_info "检测到 macOS 系统"
        # 检查是否为Apple Silicon
        if [[ $(uname -m) == "arm64" ]]; then
            ARCH="arm64"
            log_info "检测到 Apple Silicon (ARM64) 架构"
        else
            ARCH="amd64"
            log_info "检测到 Intel (AMD64) 架构"
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        log_info "检测到 Linux 系统"
        case $(uname -m) in
            x86_64) ARCH="amd64" ;;
            aarch64) ARCH="arm64" ;;
            armv7l) ARCH="arm" ;;
            *) ARCH="amd64" ;;
        esac
        log_info "检测到 $ARCH 架构"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        OS="windows"
        ARCH="amd64"
        log_info "检测到 Windows 系统 (Git Bash/Cygwin)"
        log_warn "建议使用 scripts/setup.bat 脚本"
    else
        log_error "不支持的操作系统: $OSTYPE"
        log_info "支持的系统: macOS, Linux, Windows"
        exit 1
    fi
    
    # 检查Docker
    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version | cut -d' ' -f3 | cut -d',' -f1)
        log_info "✅ Docker 已安装 (版本: $DOCKER_VERSION)"
    else
        log_error "❌ Docker 未安装，请先安装 Docker"
        echo "安装地址: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    # 检查Docker Compose
    if command -v docker-compose &> /dev/null; then
        COMPOSE_VERSION=$(docker-compose --version | cut -d' ' -f3 | cut -d',' -f1)
        log_info "✅ Docker Compose 已安装 (版本: $COMPOSE_VERSION)"
    else
        log_error "❌ Docker Compose 未安装"
        echo "请安装 Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
    
    # 检查Go环境
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | cut -d' ' -f3)
        log_info "✅ Go 已安装 (版本: $GO_VERSION)"
        
        # 检查Go版本是否符合要求 (>= 1.19)
        GO_MAJOR=$(echo $GO_VERSION | sed 's/go//' | cut -d'.' -f1)
        GO_MINOR=$(echo $GO_VERSION | sed 's/go//' | cut -d'.' -f2)
        
        if [[ $GO_MAJOR -gt 1 ]] || [[ $GO_MAJOR -eq 1 && $GO_MINOR -ge 19 ]]; then
            log_info "✅ Go 版本符合要求"
        else
            log_error "❌ Go 版本过低，需要 >= 1.19"
            exit 1
        fi
    else
        log_error "❌ Go 未安装，请先安装 Go 1.19+"
        echo "安装地址: https://golang.org/dl/"
        exit 1
    fi
    
    # 检查端口占用
    check_port() {
        local port=$1
        local service=$2
        
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null; then
            log_warn "⚠️  端口 $port 被占用 ($service)"
            echo "请停止占用端口 $port 的服务，或修改配置文件中的端口号"
            return 1
        else
            log_info "✅ 端口 $port 可用 ($service)"
            return 0
        fi
    }
    
    # 检查必要端口
    PORTS_CHECK=true
    check_port 3306 "MySQL" || PORTS_CHECK=false
    check_port 6379 "Redis" || PORTS_CHECK=false  
    check_port 19530 "Milvus" || PORTS_CHECK=false
    check_port 8080 "应用服务" || PORTS_CHECK=false
    
    if [[ "$PORTS_CHECK" == false ]]; then
        log_error "端口检查失败，请解决端口冲突后重新运行"
        exit 1
    fi
}

# 创建项目目录结构
create_directories() {
    log_step "创建项目目录结构..."
    
    mkdir -p {uploads,logs,data/{mysql,redis,milvus,etcd,minio},nginx/{ssl,conf.d},scripts}
    
    log_info "✅ 目录结构创建完成"
}

# 初始化配置文件
init_configs() {
    log_step "初始化配置文件..."
    
    # 检查配置文件是否存在
    if [[ -f "config/app.yaml" ]]; then
        log_info "配置文件已存在，跳过初始化"
        return
    fi
    
    # 创建示例配置文件
    cp config/app.yaml.example config/app.yaml 2>/dev/null || true
    
    log_info "✅ 配置文件初始化完成"
    log_warn "⚠️  请编辑 config/app.yaml 设置您的API密钥"
}

# 启动基础服务
start_services() {
    log_step "启动基础服务..."
    
    # 拉取镜像
    log_info "拉取Docker镜像..."
    docker-compose pull
    
    # 启动服务
    log_info "启动数据库和向量数据库..."
    docker-compose up -d mysql redis etcd minio milvus-standalone
    
    # 等待服务启动
    log_info "等待服务启动完成..."
    sleep 30
    
    # 检查服务状态
    check_services() {
        local service=$1
        local port=$2
        local max_retries=30
        local retry=0
        
        while [[ $retry -lt $max_retries ]]; do
            if docker-compose ps $service | grep -q "Up"; then
                if nc -z localhost $port 2>/dev/null; then
                    log_info "✅ $service 服务启动成功"
                    return 0
                fi
            fi
            
            retry=$((retry + 1))
            echo -n "."
            sleep 2
        done
        
        log_error "❌ $service 服务启动失败"
        return 1
    }
    
    # 检查各服务状态
    check_services "mysql" 3306
    check_services "redis" 6379  
    check_services "milvus-standalone" 19530
    
    log_info "✅ 基础服务启动完成"
}

# 初始化数据库
init_database() {
    log_step "初始化数据库..."
    
    # 等待MySQL完全启动
    log_info "等待MySQL服务就绪..."
    until docker exec ai-assistant-mysql mysqladmin ping -h"localhost" --silent; do
        echo -n "."
        sleep 2
    done
    echo ""
    
    # 创建数据库和表
    log_info "创建数据库表结构..."
    docker exec -i ai-assistant-mysql mysql -uai_user -pai_password ai_assistant < scripts/init.sql
    
    log_info "✅ 数据库初始化完成"
}

# 构建应用
build_app() {
    log_step "构建应用程序..."
    
    # 下载依赖
    log_info "下载Go依赖..."
    go mod download
    go mod tidy
    
    # 构建应用
    log_info "编译应用程序..."
    go build -o bin/ai-doc-assistant cmd/server/main.go
    
    log_info "✅ 应用构建完成"
}

# 启动应用
start_app() {
    log_step "启动应用服务..."
    
    # 检查配置文件中的API密钥
    if grep -q "your-volcengine-api-key-here" config/app.yaml; then
        log_error "❌ 请先在 config/app.yaml 中设置您的火山方舟API密钥"
        log_info "编辑配置文件: vim config/app.yaml"
        log_info "设置 ai.api_key 字段"
        exit 1
    fi
    
    # 启动应用
    log_info "启动AI文档助手服务..."
    nohup ./bin/ai-doc-assistant > logs/app.log 2>&1 &
    APP_PID=$!
    
    # 保存PID到文件
    echo $APP_PID > .app.pid
    
    # 等待应用启动
    log_info "等待应用服务启动..."
    sleep 5
    
    # 检查应用状态
    if curl -f http://localhost:8080/health >/dev/null 2>&1; then
        log_info "✅ AI文档助手启动成功 (PID: $APP_PID)"
    else
        log_error "❌ AI文档助手启动失败"
        log_error "请检查日志: tail -f logs/app.log"
        exit 1
    fi
}

# 显示部署结果
show_result() {
    log_step "部署完成！"
    
    echo ""
    echo -e "${GREEN}🎉 AI文档助手部署成功！${NC}"
    echo ""
    echo -e "${CYAN}访问地址:${NC}"
    echo "📱 Web界面:     http://localhost:8080"
    echo "📚 API文档:     http://localhost:8080/swagger/index.html"
    echo "❤️  健康检查:   http://localhost:8080/health"
    echo "💾 MySQL管理:   http://localhost:3306 (用户: ai_user, 密码: ai_password)"
    echo "🗃️  Milvus管理: http://localhost:9091"
    echo "📦 MinIO控制台: http://localhost:9001 (用户: minioadmin, 密码: minioadmin)"
    echo ""
    echo -e "${CYAN}管理命令:${NC}"
    echo "🔧 查看日志:     tail -f logs/app.log"
    echo "📊 服务状态:     docker-compose ps"
    echo "🛑 停止服务:     ./scripts/stop.sh"
    echo "🔄 重启服务:     ./scripts/restart.sh"
    echo ""
    echo -e "${YELLOW}下一步:${NC}"
    echo "1. 访问 http://localhost:8080 开始使用"
    echo "2. 上传一些测试文档"
    echo "3. 尝试问答功能"
    echo ""
}

# 清理函数
cleanup() {
    if [[ -n $APP_PID ]]; then
        log_info "清理进程..."
        kill $APP_PID 2>/dev/null || true
    fi
}

# 信号处理
trap cleanup EXIT

# 主函数
main() {
    echo -e "${PURPLE}"
    echo "╔══════════════════════════════════════════════════════════════════════╗"
    echo "║                     AI文档助手一键部署脚本                            ║"
    echo "║                   基于Eino框架 + 火山方舟AI                          ║"
    echo "╚══════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    log_info "开始部署AI文档助手..."
    
    # 执行部署步骤
    check_requirements
    create_directories
    init_configs
    start_services
    init_database
    build_app
    start_app
    show_result
    
    log_info "🎯 部署脚本执行完成！"
}

# 参数处理
case "${1:-}" in
    "check")
        check_requirements
        ;;
    "services")
        start_services
        ;;
    "build")
        build_app
        ;;
    "start")
        start_app
        ;;
    "")
        main
        ;;
    *)
        echo "用法: $0 [check|services|build|start]"
        echo "  check    - 仅检查系统要求"
        echo "  services - 仅启动基础服务"
        echo "  build    - 仅构建应用"
        echo "  start    - 仅启动应用"
        echo "  (无参数)  - 完整部署流程"
        exit 1
        ;;
esac