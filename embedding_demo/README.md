# 🎯 Eino Embedding 组件完全指南

本文档是对 Eino 框架中 `Embedding` 组件的核心功能和使用方式的完整总结，结合官方文档和实际项目示例。

---

## 📖 基本介绍

`Embedding` 组件是一个用于将**文本转换为向量表示**的核心组件。它的主要作用是将文本内容映射到高维向量空间，使得**语义相似的文本在向量空间中的距离较近**。

### 🎯 核心价值

在传统的文本处理中，我们只能进行精确的字符串匹配。而 Embedding 组件让我们能够理解文本的**语义含义**：

```
传统匹配："苹果" ≠ "水果"  ❌
Embedding："苹果" ≈ "水果"  ✅ (语义相似)
```

### 🚀 主要应用场景

- **🔍 语义搜索**: 查找与用户查询意图最相符的文档，而不仅仅是关键字匹配
- **📊 文本相似度计算**: 判断两段文本在意思上的接近程度
- **🎯 文本聚类分析**: 将相似的文本自动分组或分类
- **🤖 RAG (检索增强生成)**: 在 RAG 流程中，`Embedding` 是实现"检索"步骤的关键技术
- **💡 推荐系统**: 基于内容相似度的智能推荐

---

## 🔧 核心接口

`Embedding` 组件的核心接口非常简洁且强大：

```go
type Embedder interface {
    EmbedStrings(ctx context.Context, texts []string, opts ...Option) ([][]float64, error)
}
```

### 接口详解

- **`EmbedStrings` 方法**: 这是该组件最核心的方法
  - **输入**: 一个待转换的文本字符串列表 (`[]string`)
  - **输出**: 一个向量列表 (`[][]float64`)，每个向量对应一个输入文本
  - **向量维度**: 由具体模型决定（如 OpenAI 的 `text-embedding-3-small` 为 1536 维）
  - **选项参数**: 支持通过 `opts` 传入额外配置

---

## 🛠️ 使用方式

### 3.1. 单独使用

这是最直接的使用方式，适合快速获取文本的向量表示：

```go
import "github.com/cloudwego/eino-ext/components/embedding/ark"

// 1. 初始化 Embedder
// 需要提供 API Key 和指定的模型名称
embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
    APIKey: "YOUR_API_KEY",
    Model:  "bge-large-zh",
    Timeout: 30 * time.Second, // 可选：设置超时时间
})
if err != nil {
    log.Fatal("初始化 Embedder 失败:", err)
}

// 2. 调用 EmbedStrings
texts := []string{"你好", "你好吗？", "今天天气不错"}
vectors, err := embedder.EmbedStrings(ctx, texts)
if err != nil {
    log.Fatal("向量化失败:", err)
}

// vectors[0] 就是 "你好" 的向量表示
// vectors[1] 就是 "你好吗？" 的向量表示
// vectors[2] 就是 "今天天气不错" 的向量表示
fmt.Printf("生成了 %d 个向量，每个向量维度为 %d\n", len(vectors), len(vectors[0]))
```

### 3.2. 在编排中使用 (推荐)

与 `ChatTemplate` 类似，官方推荐将 `Embedding` 组件放入 `compose.Chain` 或 `compose.Graph` 中进行编排，以构建更复杂的应用（如完整的 RAG 流程）：

```go
import "github.com/cloudwego/eino/compose"

// 1. 创建一个接收 []string，输出 [][]float64 的 Chain
chain := compose.NewChain[[]string, [][]float64]()

// 2. 将 embedder 附加到链中
chain.AppendEmbedding(embedder)

// 3. 编译并运行
runnable, err := chain.Compile(ctx)
if err != nil {
    log.Fatal("编译链失败:", err)
}

vectors, err := runnable.Invoke(ctx, []string{"hello", "how are you"})
if err != nil {
    log.Fatal("执行链失败:", err)
}
```

### 3.3. 流式处理支持

Eino 的 Embedding 组件还支持流式处理，适合处理大量文本：

```go
// 创建流式处理链
streamChain := compose.NewChain[[]string, [][]float64]()
streamChain.AppendEmbedding(embedder)

runnable, _ := streamChain.Compile(ctx)

// 流式调用
stream, err := runnable.Stream(ctx, []string{"大量文本1", "大量文本2", "大量文本3"})
if err != nil {
    log.Fatal("流式处理失败:", err)
}

// 处理流式结果
for chunk := range stream {
    if chunk.Err != nil {
        log.Printf("处理出错: %v", chunk.Err)
        continue
    }
    fmt.Printf("接收到向量块: %d 个向量\n", len(chunk.Data))
}
```

---

## ⚙️ 配置选项 (Options)

### 4.1. 通用选项

`Embedding` 组件支持通过 `Option` 在调用时传入额外参数：

```go
import "github.com/cloudwego/eino/components/embedding"

// 在调用时临时切换模型
vectors, err := embedder.EmbedStrings(ctx, texts,
    embedding.WithModel("text-embedding-3-small"),
)

// 设置批处理大小（适合大量文本处理）
vectors, err := embedder.EmbedStrings(ctx, texts,
    embedding.WithBatchSize(100),
)

// 组合多个选项
vectors, err := embedder.EmbedStrings(ctx, texts,
    embedding.WithModel("text-embedding-3-large"),
    embedding.WithBatchSize(50),
)
```

### 4.2. 提供商特定选项

不同的实现（如 `ark`、`openai`）可能还支持更多特有的 `Option`：

```go
// ARK 特有选项
vectors, err := arkEmbedder.EmbedStrings(ctx, texts,
    ark.WithEndpoint("https://custom-endpoint.com"),
    ark.WithRetryCount(3),
)

// OpenAI 特有选项
vectors, err := openaiEmbedder.EmbedStrings(ctx, texts,
    openai.WithUser("user-123"),
    openai.WithDimensions(512), // 降维处理
)
```

---

## 📊 回调机制 (Callbacks)

回调机制允许开发者在 `Embedding` 的生命周期关键点注入自定义逻辑，常用于**日志记录**、**性能监控**或**调试分析**。

### 5.1. 回调事件

- **`OnStart`**: 在 `EmbedStrings` 开始执行时触发
- **`OnEnd`**: 在成功生成所有向量后触发
- **`OnError`**: 在发生错误时触发

### 5.2. 使用示例

```go
import "github.com/cloudwego/eino/callbacks"

// 1. 创建 Callback Handler
handler := &callbacks.EmbeddingCallbackHandler{
    OnStart: func(ctx context.Context, info *callbacks.EmbeddingStartInfo) {
        fmt.Printf("开始向量化 %d 个文本，使用模型: %s\n", 
            len(info.Texts), info.Model)
    },
    OnEnd: func(ctx context.Context, info *callbacks.EmbeddingEndInfo) {
        fmt.Printf("向量化完成，耗时: %v，生成 %d 个向量\n", 
            info.Duration, len(info.Vectors))
    },
    OnError: func(ctx context.Context, info *callbacks.EmbeddingErrorInfo) {
        fmt.Printf("向量化失败: %v\n", info.Error)
    },
}

callbackHandler := callbacks.NewHandlerHelper().Embedding(handler).Handler()

// 2. 在编排中使用回调
chain := compose.NewChain[[]string, [][]float64]()
chain.AppendEmbedding(embedder)

runnable, _ := chain.Compile(ctx)
vectors, _ := runnable.Invoke(ctx, texts,
    compose.WithCallbacks(callbackHandler),
)
```

### 5.3. 高级回调应用

```go
// 性能监控回调
performanceHandler := &callbacks.EmbeddingCallbackHandler{
    OnStart: func(ctx context.Context, info *callbacks.EmbeddingStartInfo) {
        // 记录开始时间
        ctx = context.WithValue(ctx, "start_time", time.Now())
    },
    OnEnd: func(ctx context.Context, info *callbacks.EmbeddingEndInfo) {
        // 计算并记录性能指标
        if startTime, ok := ctx.Value("start_time").(time.Time); ok {
            duration := time.Since(startTime)
            tokensPerSecond := float64(len(info.Vectors)) / duration.Seconds()
            fmt.Printf("性能指标 - 向量/秒: %.2f\n", tokensPerSecond)
        }
    },
}
```

---

## 🎯 实际应用示例

### 6.1. 文本相似度计算

```go
// 计算两个文本的余弦相似度
func calculateSimilarity(embedder embedding.Embedder, text1, text2 string) (float64, error) {
    vectors, err := embedder.EmbedStrings(context.Background(), []string{text1, text2})
    if err != nil {
        return 0, err
    }
    
    return cosineSimilarity(vectors[0], vectors[1]), nil
}

// 余弦相似度计算
func cosineSimilarity(a, b []float64) float64 {
    var dotProduct, normA, normB float64
    for i := range a {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

### 6.2. 语义搜索引擎

```go
type SemanticSearchEngine struct {
    embedder embedding.Embedder
    documents []Document
    vectors [][]float64
}

type Document struct {
    ID      string
    Content string
    Vector  []float64
}

// 添加文档到搜索引擎
func (s *SemanticSearchEngine) AddDocuments(docs []string) error {
    vectors, err := s.embedder.EmbedStrings(context.Background(), docs)
    if err != nil {
        return err
    }
    
    for i, doc := range docs {
        s.documents = append(s.documents, Document{
            ID:      fmt.Sprintf("doc_%d", len(s.documents)),
            Content: doc,
            Vector:  vectors[i],
        })
    }
    return nil
}

// 语义搜索
func (s *SemanticSearchEngine) Search(query string, topK int) ([]Document, error) {
    queryVectors, err := s.embedder.EmbedStrings(context.Background(), []string{query})
    if err != nil {
        return nil, err
    }
    
    queryVector := queryVectors[0]
    
    // 计算相似度并排序
    type ScoredDoc struct {
        Document
        Score float64
    }
    
    var scoredDocs []ScoredDoc
    for _, doc := range s.documents {
        score := cosineSimilarity(queryVector, doc.Vector)
        scoredDocs = append(scoredDocs, ScoredDoc{doc, score})
    }
    
    // 按相似度排序
    sort.Slice(scoredDocs, func(i, j int) bool {
        return scoredDocs[i].Score > scoredDocs[j].Score
    })
    
    // 返回 topK 结果
    var results []Document
    for i := 0; i < topK && i < len(scoredDocs); i++ {
        results = append(results, scoredDocs[i].Document)
    }
    
    return results, nil
}
```

---

## 🔧 最佳实践

### 7.1. 性能优化

1. **批量处理**: 尽量批量处理多个文本，而不是逐个调用
2. **合理设置批处理大小**: 根据模型和硬件能力调整 `BatchSize`
3. **缓存向量**: 对于重复的文本，考虑缓存其向量表示
4. **异步处理**: 对于大量文本，使用流式处理或异步处理

### 7.2. 错误处理

```go
// 带重试的向量化
func embedWithRetry(embedder embedding.Embedder, texts []string, maxRetries int) ([][]float64, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        vectors, err := embedder.EmbedStrings(context.Background(), texts)
        if err == nil {
            return vectors, nil
        }
        
        lastErr = err
        time.Sleep(time.Duration(i+1) * time.Second) // 指数退避
    }
    
    return nil, fmt.Errorf("重试 %d 次后仍然失败: %w", maxRetries, lastErr)
}
```

### 7.3. 监控和日志

```go
// 使用回调进行详细监控
monitoringHandler := &callbacks.EmbeddingCallbackHandler{
    OnStart: func(ctx context.Context, info *callbacks.EmbeddingStartInfo) {
        log.Printf("[EMBEDDING] 开始处理 %d 个文本", len(info.Texts))
    },
    OnEnd: func(ctx context.Context, info *callbacks.EmbeddingEndInfo) {
        log.Printf("[EMBEDDING] 处理完成，耗时: %v", info.Duration)
    },
    OnError: func(ctx context.Context, info *callbacks.EmbeddingErrorInfo) {
        log.Printf("[EMBEDDING] 处理失败: %v", info.Error)
    },
}
```

---

## 📚 相关资源

- **官方文档**: [Eino Embedding 组件指南](https://www.cloudwego.io/zh/docs/eino/core_modules/components/embedding_guide/)
- **GitHub 仓库**: [cloudwego/eino](https://github.com/cloudwego/eino)
- **示例代码**: 查看本项目的 `main.go` 文件获取完整示例
- **社区支持**: [CloudWeGo 社区](https://github.com/cloudwego/community)

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进这个示例项目！
