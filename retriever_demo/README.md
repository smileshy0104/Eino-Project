# 🎯 Eino Retriever 组件完全指南

本文档是对 Eino 框架中 `Retriever` 组件的核心功能和使用方式的完整总结，结合官方文档和实际项目示例。

## 🚀 快速开始

### 🛠️ 配置文件
项目使用 `config.yaml` 配置文件，也可以通过环境变量设置：
```yaml
ARK_API_KEY: "${ARK_API_KEY}"
ARK_MODEL: "deepseek-v3-1-250821"
EMBEDDER_MODEL: "doubao-embedding-text-240715"
MILVUS_ADDRESS: "localhost:19530"
MILVUS_COLLECTION: "eino_test"
```

---

## 📖 基本介绍

`Retriever` 组件是一个专门用于**检索和获取文档**的智能组件。它的主要作用是根据用户查询从向量数据库中找到**语义相关的文档**，为后续的处理（如问答生成）提供知识支持。这个组件在 AI 应用开发中扮演着**"智能检索引擎"**的角色。

### 🎯 核心价值

在传统的文档检索中，我们只能进行关键词匹配搜索。而 Retriever 组件让我们能够：

```
传统检索：关键词匹配 + 精确搜索  ❌
Retriever：语义理解 + 向量相似度搜索 + 智能检索  ✅
```

### 🚀 主要应用场景

- **🔍 语义搜索**: 基于语义相似度的智能文档检索系统
- **🤖 RAG系统**: 检索增强生成系统的知识检索中心
- **📚 知识库问答**: 从企业知识库中检索相关答案
- **💡 推荐系统**: 基于内容相似度的智能推荐
- **🔄 文档关联**: 发现相关和相似的文档内容
- **🧩 组件协作**: 与其他 Eino 组件无缝集成构建完整工作流

---

## 🔧 核心接口

`Retriever` 组件提供了简洁而强大的接口设计：

### 基础接口

```go
type Retriever interface {
    Retrieve(ctx context.Context, query string, opts ...Option) ([]*schema.Document, error)
}
```

### 接口详解

#### 📝 Retrieve 方法
- **功能**: 根据查询检索相关文档
- **输入**:
    - `ctx`: 上下文对象，用于控制超时、取消等
    - `query`: 自然语言查询文本（如 "Eino框架的主要特性"）
    - `opts`: 可选配置参数
- **输出**:
    - `[]*schema.Document`: 检索到的相关文档列表，按相似度排序
    - `error`: 检索过程中的错误信息

---

## 📨 Document 结构体

`Document` 是检索的基本数据结构，承载丰富的文档信息：

```go
type Document struct {
    // ID 是文档的唯一标识符
    ID string
    // Content 是文档的主要文本内容
    Content string
    // MetaData 存储文档的元数据信息
    MetaData map[string]interface{}
}
```

### 🎭 文档字段说明

- **🔑 ID**: 文档的唯一标识符，用于在系统中唯一标识一个文档
- **📄 Content**: 文档主要文本内容，用于语义理解和检索
- **🏷️ MetaData**: 结构化元数据，支持复杂查询和过滤。文档的元数据，可以存储如下信息：
  - 文档的来源信息
  - 文档的相似度分数（用于排序）
  - 文档的分类标签（用于过滤）
  - 文档的创建时间（用于时间过滤）
  - 其他自定义元数据

---

## 🏗️ 创建和使用 Retriever

### 基础使用流程

```go
import (
    "github.com/cloudwego/eino/schema"
    "github.com/cloudwego/eino-ext/components/retriever/milvus"
    "github.com/cloudwego/eino-ext/components/embedding/ark"
    cli "github.com/milvus-io/milvus-sdk-go/v2/client"
)

// 1️⃣ 初始化 Embedder（向量化组件）
embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
    APIKey: "your-api-key",
    Model:  "doubao-embedding-text-240715",
})
if err != nil {
    log.Fatal("Embedder 初始化失败:", err)
}

// 2️⃣ 配置 Milvus 客户端
client, err := cli.NewClient(ctx, cli.Config{
    Address: "localhost:19530",
})
if err != nil {
    log.Fatal("Milvus 客户端创建失败:", err)
}

// 3️⃣ 创建 Retriever
cfg := &milvus.RetrieverConfig{
    Client:       client,
    Collection:   "eino_test",
    VectorField:  "vector",
    Embedding:    embedder,
    OutputFields: []string{"id", "content", "metadata"},
    TopK:         5,
}
retriever, err := milvus.NewRetriever(ctx, cfg)
if err != nil {
    log.Fatal("Retriever 创建失败:", err)
}

// 4️⃣ 执行检索
query := "Eino框架的主要特性"
docs, err := retriever.Retrieve(ctx, query)
if err != nil {
    log.Fatal("检索失败:", err)
}

// 5️⃣ 处理检索结果
for i, doc := range docs {
    fmt.Printf("文档%d: %s\n", i+1, doc.Content)
    fmt.Printf("元数据: %v\n", doc.MetaData)
}
```

### 🎯 实用配置示例

#### 不同TopK配置的效果
```go
// 精准检索：只要最相关的
cfg.TopK = 1

// 平衡检索：常用配置
cfg.TopK = 5

// 探索检索：获取更多候选
cfg.TopK = 10
```

#### 复杂检索查询示例
```go
// 技术问题查询
query := "如何优化向量数据库的检索性能？"
docs, err := retriever.Retrieve(ctx, query)

// 概念解释查询
query = "RAG系统的工作原理"
docs, err := retriever.Retrieve(ctx, query)

// 实践指导查询
query = "Eino框架最佳实践"
docs, err := retriever.Retrieve(ctx, query)
```

---

## 🚀 编排集成最佳实践

虽然可以直接使用 Retriever，但**官方强烈推荐**将其集成到编排工作流中，与其他组件协同工作。

### 🔗 Chain 编排模式

Chain 是最常用的编排方式，适合线性处理流程：

```go
import "github.com/cloudwego/eino/compose"

// 1️⃣ 创建 Chain - 声明输入输出类型  
chain := compose.NewChain[string, *schema.Message]()

// 2️⃣ 添加组件 - 按处理顺序添加
var queryForPrompt string

// 保存查询并转换格式
chain.AppendLambda(
    compose.InvokableLambda(func(ctx context.Context, query string) (map[string]any, error) {
        queryForPrompt = query
        return map[string]any{"query": query}, nil
    }),
)

// 添加检索步骤
chain.AppendRetriever(retriever, compose.WithInputKey("query"))

// 构建Prompt
chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, docs []*schema.Document) ([]*schema.Message, error) {
    prompt := "请根据以下背景知识回答问题:\n\n"
    for i, doc := range docs {
        prompt += fmt.Sprintf("[%d] %s\n", i+1, doc.Content)
    }
    prompt += fmt.Sprintf("\n问题: %s", queryForPrompt)
    
    return []*schema.Message{
        schema.UserMessage(prompt),
    }, nil
}))

// 添加ChatModel生成答案
chain.AppendChatModel(chatModel)

// 3️⃣ 编译执行
runnable, err := chain.Compile(ctx)
if err != nil {
    log.Fatalf("链编译失败: %v", err)
}

// 4️⃣ 运行RAG工作流
result, err := runnable.Invoke(ctx, "Eino框架是什么？")
```

### 🔄 完整RAG处理工作流

```go
func createRAGWorkflow() (*compose.Runnable, error) {
    ctx := context.Background()
    
    // 🔧 初始化组件
    embedder, err := initEmbedder(ctx)
    if err != nil {
        return nil, err
    }
    
    retriever, err := initRetriever(ctx, embedder)
    if err != nil {
        return nil, err
    }
    
    chatModel, err := initChatModel(ctx)
    if err != nil {
        return nil, err
    }
    
    // 🔗 构建RAG处理链
    // 查询输入 → 文档检索 → Prompt构建 → 答案生成
    chain := compose.NewChain[string, *schema.Message]()
    
    var currentQuery string
    
    // 步骤1: 查询预处理
    chain.AppendLambda(
        compose.InvokableLambda(func(ctx context.Context, query string) (map[string]any, error) {
            currentQuery = query
            return map[string]any{"query": query}, nil
        }),
    )
    
    // 步骤2: 文档检索
    chain.AppendRetriever(retriever, compose.WithInputKey("query"))
    
    // 步骤3: Prompt构建
    chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, docs []*schema.Document) ([]*schema.Message, error) {
        if len(docs) == 0 {
            return []*schema.Message{
                schema.UserMessage(fmt.Sprintf("请回答问题：%s", currentQuery)),
            }, nil
        }
        
        prompt := "请根据以下背景知识严格回答问题。如果背景知识不足，请说明。\n\n背景知识:\n"
        for i, doc := range docs {
            prompt += fmt.Sprintf("%d. %s\n\n", i+1, doc.Content)
        }
        prompt += fmt.Sprintf("问题: %s", currentQuery)
        
        return []*schema.Message{
            schema.SystemMessage("你是一个严谨的AI助手，请基于提供的背景知识回答问题。"),
            schema.UserMessage(prompt),
        }, nil
    }))
    
    // 步骤4: 答案生成
    chain.AppendChatModel(chatModel)
    
    // ⚙️ 编译成可运行实例
    return chain.Compile(ctx)
}

// 使用示例
func processQuery(query string) (*schema.Message, error) {
    workflow, err := createRAGWorkflow()
    if err != nil {
        return nil, fmt.Errorf("RAG工作流创建失败: %w", err)
    }
    
    result, err := workflow.Invoke(context.Background(), query)
    if err != nil {
        return nil, fmt.Errorf("查询处理失败: %w", err)
    }
    
    return result, nil
}
```

---

## ⚙️ 高级配置和选项

### Option 配置

Retriever 支持通过 Option 在运行时传入额外配置：

```go
import retrieverCm "github.com/cloudwego/eino/components/retriever"

// WithTopK - 设置返回文档数量
docs, err := retriever.Retrieve(ctx, query,
    retrieverCm.WithTopK(10),
)

// WithScoreThreshold - 设置相似度阈值
docs, err := retriever.Retrieve(ctx, query,
    retrieverCm.WithTopK(5),
    retrieverCm.WithScoreThreshold(0.7),
)

// WithOutputFields - 指定返回字段
docs, err := retriever.Retrieve(ctx, query,
    retrieverCm.WithOutputFields([]string{"id", "content", "score"}),
)
```

### Callback 机制

回调机制允许在关键生命周期节点注入自定义逻辑：

```go
// 创建回调处理器
callbackHandler := callbacks.NewHandlerBuilder().
    OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
        fmt.Printf("🚀 开始检索查询: %s\n", input.(string))
        return ctx
    }).
    OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) {
        docs := output.([]*schema.Document)
        fmt.Printf("✅ 检索完成，找到 %d 个文档\n", len(docs))
    }).
    OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) {
        fmt.Printf("❌ 检索失败: %v\n", err)
    }).
    Build()

// 在编排中使用回调
chain := compose.NewChain[string, []*schema.Document]()
chain.AppendRetriever(retriever, compose.WithCallbacks(callbackHandler))
```

---

## 🎓 高级用法和技巧

### 1. 📊 动态配置管理

根据不同场景动态选择检索参数：

```go
type RetrieverManager struct {
    configs   map[string]*milvus.RetrieverConfig
    retrievers map[string]*milvus.Retriever
}

func (rm *RetrieverManager) GetRetriever(scenario string) (*milvus.Retriever, error) {
    if retriever, exists := rm.retrievers[scenario]; exists {
        return retriever, nil
    }
    
    config := rm.configs[scenario]
    if config == nil {
        return nil, fmt.Errorf("未找到场景 %s 的配置", scenario)
    }
    
    retriever, err := milvus.NewRetriever(context.Background(), config)
    if err != nil {
        return nil, err
    }
    
    rm.retrievers[scenario] = retriever
    return retriever, nil
}
```

### 2. 🔄 检索结果后处理

```go
func postProcessResults(docs []*schema.Document, query string) []*schema.Document {
    // 去重处理
    seen := make(map[string]bool)
    uniqueDocs := []*schema.Document{}
    
    for _, doc := range docs {
        if !seen[doc.ID] {
            uniqueDocs = append(uniqueDocs, doc)
            seen[doc.ID] = true
        }
    }
    
    // 根据查询相关性进一步排序
    // ... 自定义排序逻辑
    
    return uniqueDocs
}
```

### 3. 📈 性能监控

```go
type RetrieverMetrics struct {
    TotalQueries     int64
    SuccessfulQueries int64
    FailedQueries    int64
    AverageLatency   time.Duration
    LastQueryTime    time.Time
}

func (m *RetrieverMetrics) RecordQuery(duration time.Duration, success bool, docCount int) {
    m.TotalQueries++
    m.LastQueryTime = time.Now()
    
    if success {
        m.SuccessfulQueries++
    } else {
        m.FailedQueries++
    }
    
    // 更新平均延迟（简化版本）
    m.AverageLatency = (m.AverageLatency + duration) / 2
}

func retrieveWithMetrics(retriever *milvus.Retriever, query string, metrics *RetrieverMetrics) ([]*schema.Document, error) {
    startTime := time.Now()
    
    docs, err := retriever.Retrieve(context.Background(), query)
    
    duration := time.Since(startTime)
    metrics.RecordQuery(duration, err == nil, len(docs))
    
    return docs, err
}
```

---

## ❓ 常见问题和解决方案

### Q1: 检索返回空结果

**问题**: 查询总是返回0个文档
```
❌ 检索到0个相关文档，请确保已运行indexer_demo填充数据
```

**解决方案**: 
```bash
# 首先运行indexer_demo填充数据
cd ../indexer_demo
go run main.go

# 然后运行retriever_demo
cd ../retriever_demo
go run main.go
```

### Q2: Milvus连接失败

**问题**: 无法连接到Milvus服务
```
❌ 初始化Milvus客户端失败: connection refused
```

**解决方案**:
```bash
# 检查Milvus服务状态
docker-compose ps

# 启动Milvus服务
docker-compose up -d

# 验证连接
telnet localhost 19530
```

### Q3: 向量维度不匹配

**问题**: Embedding模型输出维度与存储不匹配
```go
// ✅ 确保配置一致
embedderConfig := &ark.EmbeddingConfig{
    Model: "doubao-embedding-text-240715",  // 确认模型输出维度
}

// 对应的Milvus集合字段定义
vectorField := &entity.Field{
    Name:        "vector",
    DataType:    entity.FieldTypeBinaryVector,
    TypeParams:  map[string]string{"dim": "81920"},  // 匹配模型输出
}
```

### Q4: 检索性能慢

**解决方案**:
```go
// ✅ 性能优化配置
cfg := &milvus.RetrieverConfig{
    Client:       client,
    Collection:   "eino_test",
    VectorField:  "vector",
    Embedding:    embedder,
    OutputFields: []string{"id", "content"},  // 减少返回字段
    TopK:         3,  // 适当减少TopK
}
```

---

## 💡 使用最佳实践

### 性能优化

1. **合理设置TopK**: 根据业务需求平衡相关性和性能
2. **选择合适的向量维度**: 平衡精度和计算开销
3. **使用连接池**: 减少频繁连接建立的开销
4. **结果缓存**: 对热点查询进行智能缓存
5. **批量检索**: 合并相似查询提高效率

### 错误处理

1. **配置验证**: 启动前检查所有必要配置
2. **连接管理**: 实现健康检查和自动重连
3. **重试策略**: 对临时错误实施指数退避重试
4. **降级方案**: 准备服务不可用时的备选策略
5. **日志记录**: 详细记录错误信息便于排查

### 生产部署

1. **监控指标**: 监控检索延迟、成功率、吞吐量
2. **资源管理**: 合理设置超时时间和并发限制
3. **索引优化**: 定期优化向量索引提高查询效率
4. **数据治理**: 保持文档数据的新鲜度和质量
5. **安全控制**: 实施访问控制和审计日志

---

## 📊 性能基准

### 测试环境
- **硬件**: MacBook Pro (M1 Pro)
- **Milvus版本**: v2.4.2
- **向量维度**: 81920维 (doubao-embedding-text-240715)
- **网络**: 千兆宽带

### 基准数据
- **单次检索**: ~200-440ms
- **批量检索**: 3.75-6.02 QPS
- **TopK=5**: 最佳性价比配置
- **成功率**: 100% (稳定运行)

---

## 🚀 进阶用法

### 自定义检索策略

项目提供多种检索配置模式：

**精准检索**（高质量）:
- 配置TopK=1, 适当提高相似度阈值
- 只返回最相关的文档，确保高质量结果
- 适用于需要精确答案的场景

**探索检索**（信息发现）:
- 配置TopK=10, 降低相似度阈值
- 返回更多候选文档，适合信息发现
- 适用于研究分析、知识探索场景

**高性能检索**（速度优先）:
- 配置TopK=3, 精简OutputFields
- 减少数据传输，提高响应速度
- 适用于实时系统、高并发场景

### RAG系统集成

Retriever可以无缝集成到RAG系统中：

- **知识检索**：从向量数据库获取相关背景知识
- **上下文构建**：为大语言模型提供相关信息
- **答案增强**：提高生成答案的准确性和可信度
- **工作流编排**：与其他组件协同构建智能应用

---

## 🎉 总结

Retriever 是 Eino 框架中的**核心检索组件**，掌握它的使用对于构建高质量的 AI 应用至关重要：

### 🏆 核心优势
- **🔍 语义检索**: 强大的向量相似度搜索和智能检索能力
- **⚡ 高性能**: 支持批量处理和并发操作，适应大规模应用
- **🤖 RAG集成**: 完美集成到检索增强生成系统中
- **🧩 组件化**: 与 Eino 生态系统深度集成，构建完整工作流
- **🛡️ 可靠性**: 完善的错误处理和恢复机制，保证服务稳定
- **🔧 灵活性**: 支持多种配置选项和自定义扩展

### 💡 最佳实践总结
1. **合理配置**: 根据应用场景选择合适的TopK和输出字段
2. **性能优化**: 使用适当的批量大小和连接池管理
3. **错误处理**: 实施完善的错误检测、分类和恢复机制
4. **资源管理**: 正确管理数据库连接、内存使用和并发控制
5. **监控观测**: 定期监控检索性能、成功率和系统健康状态
6. **编排集成**: 优先使用 Chain 编排构建自动化工作流

### 🔗 相关资源
- 📚 [Eino官方文档](https://www.cloudwego.io/zh/docs/eino/)
- 🌐 [GitHub 仓库](https://github.com/cloudwego/eino)
- 📝 [Retriever API参考](https://www.cloudwego.io/zh/docs/eino/core_modules/components/retriever_guide/)
- 🗄️ [Milvus 官方文档](https://milvus.io/docs)

通过掌握 Retriever 组件的各种功能和最佳实践，你将能够构建出更加智能、高效和可扩展的文档检索和RAG系统！🚀