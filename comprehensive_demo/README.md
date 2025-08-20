# Eino 综合演示系统

这是一个完整的 Eino 框架演示程序，集成了 Transformer、Indexer、Retriever 和 Tool 等所有核心组件，展示了如何构建一个功能完整的智能 RAG + Tool 系统。

## 🚀 系统特性

### 核心组件集成
- **📝 Transformer**: 智能文档分割，支持 Markdown 格式
- **📚 Indexer**: 文档向量化与存储到 Milvus
- **🔍 Retriever**: 基于语义相似度的知识检索
- **🔧 Tools**: 多种实用工具集成
- **🤖 RAG**: 检索增强生成，提供准确回答
- **⚡ Chain**: 端到端工作流编排

### 内置工具
1. **知识搜索工具** - 从向量数据库检索相关知识
2. **文档处理工具** - 分割和索引新文档到知识库
3. **计算器工具** - 执行基本数学计算
4. **天气查询工具** - 模拟天气信息查询

## 📋 运行前准备

### 1. 环境要求
- Go 1.19+
- Milvus 2.3+ (可使用 Docker 快速部署)
- 火山方舟 API Key

### 2. 启动 Milvus (Docker)
```bash
# 下载并启动 Milvus
wget https://github.com/milvus-io/milvus/releases/download/v2.3.0/milvus-standalone-docker-compose.yml -O docker-compose.yml
docker-compose up -d
```

### 3. 配置设置
复制配置文件模板并填入实际配置:
```bash
cp config.yaml.example config.yaml
# 编辑 config.yaml，填入您的 API Key 和其他配置
```

## 🎯 配置说明

### config.yaml 配置项
```yaml
# Milvus 向量数据库配置
MILVUS_ADDRESS: "localhost:19530"          # Milvus 服务地址
MILVUS_COLLECTION: "eino_comprehensive"    # 集合名称

# 火山方舟 API 配置
ARK_API_KEY: "your-ark-api-key-here"       # 火山方舟 API Key
EMBEDDER_MODEL: "your-embedder-model"      # 嵌入模型名称
ARK_MODEL: "your-chat-model"               # 聊天模型名称
```

### 环境变量配置 (可选)
也可通过环境变量设置:
```bash
export MILVUS_ADDRESS="localhost:19530"
export MILVUS_COLLECTION="eino_comprehensive"
export ARK_API_KEY="your-ark-api-key"
export EMBEDDER_MODEL="doubao-embedding"
export ARK_MODEL="doubao-pro-4k"
```

## 🚀 运行演示

### 基本运行
```bash
# 进入演示目录
cd comprehensive_demo

# 安装依赖
go mod tidy

# 运行演示
go run main.go
```

### 输出示例
程序运行后将展示以下过程:

1. **系统初始化** - 各组件启动和连接
2. **知识库加载** - 分割和索引示例文档
3. **查询处理演示** - 处理多个示例查询
4. **工具调用展示** - 演示各种工具的使用
5. **RAG 回答生成** - 基于检索知识的智能回答

## 📖 系统架构

### 数据流程
```
用户查询 → 知识检索 → 工具调用 → 提示构建 → LLM 生成 → 最终回答
    ↑         ↑         ↑         ↑         ↑         ↑
Transformer → Indexer → Tools → RAG Chain → ChatModel → Response
```

### 核心类说明

#### `ComprehensiveRAGSystem`
- 系统主类，协调所有组件
- 管理配置、初始化和资源清理
- 提供统一的查询处理接口

#### `KnowledgeSearchTool`
- 知识搜索工具实现
- 基于 Retriever 组件进行语义检索
- 支持自定义 TopK 参数

#### `DocumentProcessorTool`
- 文档处理工具实现
- 集成 Transformer 和 Indexer
- 支持实时文档添加到知识库

## 🔧 自定义扩展

### 添加新工具
1. 实现 `tool.BaseTool` 接口
2. 在 `initTools` 方法中注册工具
3. 工具会自动集成到系统中

```go
type CustomTool struct{}

func (c *CustomTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    // 返回工具信息
}

func (c *CustomTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
    // 实现工具逻辑
}
```

### 扩展文档类型
- 修改 `initTransformer` 方法
- 添加新的文档分割器
- 支持更多文档格式

### 自定义 RAG 流程
- 修改 `buildChain` 方法
- 使用 `compose` 包构建复杂工作流
- 添加更多处理节点

## 📚 演示内容

### 内置知识库
系统预加载了以下示例文档:
- Eino 框架介绍
- RAG 技术详解
- 工具使用指南

### 示例查询
演示程序会处理以下查询:
1. "什么是 Eino 框架？"
2. "RAG 技术有什么优势？"
3. "如何使用工具系统？"

## 🛠️ 故障排除

### 常见问题

1. **Milvus 连接失败**
   - 确保 Milvus 服务正在运行
   - 检查 `MILVUS_ADDRESS` 配置

2. **API Key 错误**
   - 验证 `ARK_API_KEY` 配置
   - 确保模型名称正确

3. **向量维度不匹配**
   - 确保 `milvusSchema` 中的维度与模型匹配
   - 检查 embedding 模型输出

### 调试模式
程序包含详细的日志输出，便于调试:
- ✓ 成功操作标记
- ❌ 错误操作标记
- 详细的执行步骤记录

## 📈 性能优化建议

1. **Milvus 配置优化**
   - 调整索引参数
   - 配置适当的分片数量

2. **批量处理**
   - 批量索引文档
   - 并发处理查询

3. **缓存机制**
   - 缓存常用查询结果
   - 复用 embedding 结果

## 🤝 贡献指南

欢迎贡献代码和改进建议:
1. Fork 项目
2. 创建功能分支
3. 提交 Pull Request
4. 详细描述修改内容

## 📄 许可证

本项目遵循 MIT 许可证。

## 🙋‍♂️ 支持

如有问题或建议，请：
- 提交 Issue
- 查看 Eino 官方文档
- 联系开发团队

---

*这个综合演示展示了 Eino 框架的强大功能，为构建智能应用提供了完整的参考实现。*