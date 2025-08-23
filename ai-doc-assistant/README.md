# AI文档助手 - 基于Eino框架的智能问答系统

## 🎯 项目介绍

基于Eino框架开发的企业级AI文档助手，支持飞书文档集成、智能问答、语义检索等功能。

## 🏗️ 项目结构

```
ai-doc-assistant/
├── cmd/                    # 应用程序入口
│   └── server/
│       └── main.go
├── internal/              # 内部代码
│   ├── config/           # 配置管理
│   ├── handler/          # HTTP处理器
│   ├── service/          # 业务逻辑层
│   ├── repository/       # 数据访问层
│   └── model/           # 数据模型
├── pkg/                  # 可共享的库代码
│   ├── feishu/          # 飞书API客户端
│   ├── eino/            # Eino框架封装
│   └── utils/           # 工具函数
├── web/                 # 前端资源
├── scripts/             # 脚本文件
├── config/              # 配置文件
├── docker/              # Docker相关文件
└── docs/                # 文档
```

## 🚀 快速开始

### 1. 环境准备
```bash
# 安装Go 1.19+
go version

# 安装Docker
docker --version

# 克隆项目
git clone <repo-url>
cd ai-doc-assistant
```

### 2. 启动基础服务
```bash
# 启动数据库和向量数据库
docker-compose up -d

# 等待服务启动
sleep 30
```

### 3. 配置应用
```bash
# 复制配置文件
cp config/app.yaml.example config/app.yaml

# 编辑配置文件，设置API密钥
vim config/app.yaml
```

### 4. 运行应用
```bash
# 安装依赖
go mod tidy

# 运行应用
go run cmd/server/main.go
```

### 5. 访问应用
- Web界面: http://localhost:8080
- API文档: http://localhost:8080/swagger
- 健康检查: http://localhost:8080/health

## 🔧 开发指南

### 添加新的文档处理器
1. 在 `internal/service/document.go` 中添加处理逻辑
2. 在 `internal/handler/document.go` 中添加HTTP接口
3. 更新路由配置

### 扩展问答功能
1. 修改 `internal/service/qa.go` 中的问答逻辑
2. 添加新的Tool到 `pkg/eino/tools.go`
3. 更新Chain配置

## 📊 监控和日志

- 应用日志: `logs/app.log`
- 访问日志: `logs/access.log`
- 性能监控: Prometheus metrics端点 `/metrics`

## 🧪 测试

```bash
# 运行单元测试
go test ./...

# 运行集成测试
go test -tags=integration ./...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📝 License

MIT License