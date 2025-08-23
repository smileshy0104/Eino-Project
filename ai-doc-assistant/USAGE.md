# AI文档助手使用指南

## 🚀 快速开始

### 方式一：快速体验（推荐）

```bash
# 1. 进入项目目录
cd ai-doc-assistant

# 2. 快速启动基础服务
./quick-start.sh

# 3. 配置API密钥
vim config/app.yaml
# 修改 ai.api_key 为您的火山方舟API密钥

# 4. 启动应用
go run cmd/server/main.go
```

### 方式二：完整部署

```bash
# 使用一键部署脚本
make install

# 或手动执行
./scripts/setup.sh
```

### 方式三：开发模式

```bash
# 安装开发依赖
go mod tidy

# 启动开发模式（推荐安装air热重载工具）
make dev

# 或直接运行
go run cmd/server/main.go
```

## 📋 系统要求

- **Go**: 1.19 或更高版本
- **Docker**: 最新版本
- **Docker Compose**: 最新版本
- **内存**: 至少 4GB 可用内存
- **存储**: 至少 2GB 可用空间

## ⚙️ 配置说明

### 核心配置 (config/app.yaml)

```yaml
# 服务器配置
server:
  port: 8080
  mode: "debug"  # debug, release, test

# AI服务配置（必须配置）
ai:
  provider: "volcengine"
  api_key: "your-api-key-here"  # ⚠️ 必须替换
  base_url: "https://ark.cn-beijing.volces.com/api/v3"
  models:
    embedding: "doubao-embedding"
    chat: "doubao-seed"

# 数据库配置（默认值一般无需修改）
database:
  mysql:
    host: "localhost"
    port: 3306
    username: "ai_user"
    password: "ai_password"
    database: "ai_assistant"
  
  milvus:
    host: "localhost"
    port: 19530
    
  redis:
    host: "localhost"
    port: 6379
```

### 环境变量配置

```bash
# 可选：通过环境变量设置API密钥
export AI_DOC_AI_API_KEY="d0666bb8-8a41-42f4-bd06-94ca6ba08457"
export AI_DOC_SERVER_PORT=8080
```

## 📱 使用说明

### 1. 访问Web界面

启动成功后，访问: http://localhost:8080

### 2. 上传文档

支持的文档格式：
- PDF (.pdf)
- Word文档 (.docx)  
- Markdown (.md)
- 纯文本 (.txt)
- Excel表格 (.xlsx)
- PowerPoint (.pptx)

### 3. 智能问答

#### 问题示例：
```
用户登录的验证码有效期是多少？
支付模块支持哪些支付方式？
API接口的错误码定义是什么？
最近版本有什么重要变更？
```

#### 高级提问技巧：
- **具体询问**: "用户登录模块的密码策略是什么？"
- **版本对比**: "支付模块v1.0和v1.5有什么区别？"
- **功能查询**: "如何实现微信支付集成？"
- **历史追溯**: "验证码有效期之前有调整过吗？"

### 4. API接口调用

#### 上传文档
```bash
curl -X POST http://localhost:8080/api/v1/documents \
  -F "file=@document.pdf" \
  -F "title=测试文档" \
  -F "author=张三" \
  -F "department=产品部"
```

#### 智能问答
```bash
curl -X POST http://localhost:8080/api/v1/qa/ask \
  -H "Content-Type: application/json" \
  -d '{
    "question": "用户登录验证码有效期是多少？",
    "user_id": "user_001"
  }'
```

#### 查询历史
```bash
curl http://localhost:8080/api/v1/qa/history?user_id=user_001
```

## 🛠️ 开发命令

```bash
# 查看所有可用命令
make help

# 开发相关
make dev          # 开发模式启动
make build        # 构建应用
make test         # 运行测试
make format       # 格式化代码

# 服务管理
make start        # 启动所有服务
make stop         # 停止所有服务
make restart      # 重启服务
make status       # 检查服务状态

# 运维工具
make logs         # 查看应用日志
make health       # 健康检查
make db-backup    # 数据库备份
```

## 📊 监控和管理

### 服务状态检查

```bash
# 检查所有服务状态
make status

# 健康检查
make health

# 查看Docker服务
docker-compose ps
```

### 日志查看

```bash
# 应用日志
make logs

# Docker服务日志
docker-compose logs -f

# 特定服务日志
docker-compose logs mysql
docker-compose logs milvus-standalone
```

### 数据管理

```bash
# 数据库备份
make db-backup

# 重新初始化数据库
make db-migrate

# 生成测试数据
make mock-data
```

## 🎯 使用场景

### 1. 企业知识库问答
- 上传公司的产品需求文档、技术文档、API文档
- 员工可以自然语言提问快速获取信息
- 支持版本对比和历史追溯

### 2. 开发团队协作
- 技术规范文档智能检索
- 接口文档快速查询
- 架构设计决策追溯

### 3. 产品经理工具
- 需求文档智能分析
- 功能变更历史追踪
- 竞品分析文档管理

### 4. 客服支持系统
- 产品使用手册智能问答
- 常见问题自动回复
- 客户反馈分析

## 📈 性能优化

### 向量数据库优化
```bash
# Milvus集合信息查看
curl http://localhost:9091/collections

# 索引状态检查
curl http://localhost:9091/indexes
```

### 缓存配置
Redis缓存用于：
- API响应缓存
- 热门问题缓存
- 用户会话管理

### 数据库优化
- MySQL慢查询日志: `logs/mysql-slow.log`
- 索引优化建议: 查看 `scripts/init.sql`

## 🔧 故障排除

### 常见问题

#### 1. API密钥错误
```
错误: AI API密钥未配置或无效
解决: 检查 config/app.yaml 中的 ai.api_key 设置
```

#### 2. 数据库连接失败
```bash
# 检查MySQL服务
docker-compose ps mysql

# 重启数据库
docker-compose restart mysql
```

#### 3. Milvus连接超时
```bash
# 检查Milvus服务状态
curl http://localhost:9091/healthz

# 重启Milvus
docker-compose restart milvus-standalone
```

#### 4. 端口冲突
```bash
# 查看端口占用
lsof -ti:8080

# 杀死占用进程
kill -9 $(lsof -ti:8080)
```

#### 5. 向量化失败
- 检查火山方舟API配额
- 验证网络连接
- 查看应用日志: `make logs`

### 调试模式

```bash
# 详细日志模式
export AI_DOC_LOG_LEVEL=debug
go run cmd/server/main.go

# 查看详细日志
tail -f logs/app.log
```

## 🔐 安全注意事项

1. **API密钥安全**
   - 不要将API密钥提交到版本控制
   - 生产环境使用环境变量
   - 定期轮换API密钥

2. **数据库安全**
   - 修改默认密码
   - 限制网络访问
   - 定期备份数据

3. **文件上传安全**
   - 限制文件类型和大小
   - 扫描恶意文件
   - 权限控制

## 📞 获取帮助

- **使用问题**: 查看本文档或运行 `make help`
- **技术支持**: 查看项目GitHub Issues
- **功能建议**: 提交Feature Request
- **Bug报告**: 提交详细的错误日志和复现步骤

## 🎉 最佳实践

1. **文档管理**
   - 使用有意义的文档标题
   - 添加适当的标签和分类
   - 定期更新文档内容

2. **问答技巧**
   - 问题尽量具体明确
   - 包含相关上下文信息
   - 使用专业术语

3. **系统维护**
   - 定期查看系统日志
   - 监控资源使用情况
   - 及时备份重要数据

4. **性能优化**
   - 合理设置检索参数
   - 监控响应时间
   - 优化数据库查询