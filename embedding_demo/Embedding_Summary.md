# 🎯 Eino Embedding 组件完全指南

本文档是对 Eino 框架中 `Embedding` 组件的核心功能和使用方式的完整总结，结合官方文档和实际项目示例。

## 🚀 快速开始

### 🛠️ 配置文件
项目使用 `config.yaml` 配置文件，也可以通过环境变量设置：
```yaml
ARK_API_KEY: "${ARK_API_KEY}"
EMBEDDER_MODEL: "doubao-embedding-text-240715"
```

---

## 📖 基本介绍

`Embedding` 组件是一个专门用于**文本向量化**的智能组件。它的主要作用是将文本内容映射到高维向量空间，使得**语义相似的文本在向量空间中距离较近**。这个组件在 AI 应用开发中扮演着**"语义理解引擎"**的角色。

### 🎯 核心价值

在传统的文本处理中，我们只能进行关键词匹配搜索。而 Embedding 组件让我们能够：

```
传统匹配：关键词精确匹配 + 字符串比较  ❌
Embedding：语义理解 + 向量相似度计算 + 智能匹配  ✅
```

### 🚀 主要应用场景

- **🔍 语义搜索**: 基于语义相似度的智能文档检索系统
- **📊 文本相似度计算**: 判断两段文本在意思上的接近程度
- **🎯 文本聚类分析**: 将相似主题的文本自动分组和分类
- **🤖 RAG 系统**: 检索增强生成系统的向量化核心
- **💡 推荐系统**: 基于内容相似度的智能内容推荐
- **🧩 组件协作**: 与其他 Eino 组件无缝集成构建完整工作流

---

## 🔧 核心接口

`Embedding` 组件提供了简洁而强大的接口设计：

### 基础接口

```go
type Embedder interface {
    EmbedStrings(ctx context.Context, texts []string, opts ...Option) ([][]float64, error)
}
```

### 接口详解

#### 📝 EmbedStrings 方法
- **功能**: 将文本列表转换为对应的向量表示
- **输入**:
    - `ctx`: 上下文对象，用于控制超时、取消等
    - `texts`: 文本字符串列表 (`[]string`)
    - `opts`: 可选配置参数
- **输出**:
    - `[][]float64`: 向量列表，每个向量对应一个输入文本
    - `error`: 向量化过程中的错误信息

---

## 🏗️ 创建和使用 Embedding

### 基础使用流程

```go
import (
    "github.com/cloudwego/eino-ext/components/embedding/ark"
)

// 1️⃣ 初始化 Embedder
embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
    APIKey: "your-api-key",
    Model:  "doubao-embedding-text-240715",
    Timeout: 30 * time.Second, // 可选：设置超时时间
})
if err != nil {
    log.Fatal("Embedder 初始化失败:", err)
}

// 2️⃣ 准备文本
texts := []string{
    "这是第一个示例文本",
    "这是第二个示例文本",
    "这是第三个示例文本",
}

// 3️⃣ 执行向量化
vectors, err := embedder.EmbedStrings(ctx, texts)
if err != nil {
    log.Fatal("文本向量化失败:", err)
}

// 4️⃣ 处理结果
fmt.Printf("成功生成 %d 个向量，每个向量维度: %d\n", len(vectors), len(vectors[0]))
```

### 🎯 实用配置示例

#### ARK Embedder 配置
```go
type EmbeddingConfig struct {
    APIKey  string         // ARK API密钥
    Model   string         // 模型名称，如 "doubao-embedding-text-240715"
    Timeout *time.Duration // 请求超时时间
}

// 创建配置
config := &ark.EmbeddingConfig{
    APIKey:  "your-ark-api-key",
    Model:   "doubao-embedding-text-240715",
    Timeout: func() *time.Duration { t := 30 * time.Second; return &t }(),
}
```

#### 复杂文本示例
```go
complexTexts := []string{
    "人工智能技术发展现状：深度学习在图像识别、自然语言处理等领域取得了突破性进展。",
    "云原生架构设计原则：微服务、容器化、持续集成和部署是现代软件开发的核心理念。",
    "量子计算研究前沿：量子比特的相干性和纠缠性为解决复杂计算问题提供了新的可能。",
}

vectors, err := embedder.EmbedStrings(ctx, complexTexts)
if err != nil {
    log.Fatal("复杂文本向量化失败:", err)
}
```

---

## 🚀 编排集成最佳实践

虽然可以直接使用 Embedding，但**官方强烈推荐**将其集成到编排工作流中，与其他组件协同工作。

### 🔗 Chain 编排模式

Chain 是最常用的编排方式，适合线性处理流程：

```go
import "github.com/cloudwego/eino/compose"

// 1️⃣ 创建 Chain - 声明输入输出类型  
chain := compose.NewChain[[]string, [][]float64]()

// 2️⃣ 添加组件 - 按处理顺序添加
chain.AppendEmbedding(embedder)

// 3️⃣ 编译执行
runnable, err := chain.Compile(ctx)
if err != nil {
    log.Fatalf("链编译失败: %v", err)
}

// 4️⃣ 运行工作流
vectors, err := runnable.Invoke(ctx, texts)
```

### 🔄 完整文本处理工作流

```go
func createTextProcessingWorkflow() (*compose.Runnable, error) {
    ctx := context.Background()
    
    // 🔧 初始化组件
    embedder, err := initEmbedder(ctx)
    if err != nil {
        return nil, err
    }
    
    // 🔗 构建处理链
    // 文本输入 → 向量化 → 返回向量列表
    chain := compose.NewChain[[]string, [][]float64]()
    chain.AppendEmbedding(embedder)
    
    // ⚙️ 编译成可运行实例
    return chain.Compile(ctx)
}

// 使用示例
func processTexts(texts []string) ([][]float64, error) {
    workflow, err := createTextProcessingWorkflow()
    if err != nil {
        return nil, fmt.Errorf("工作流创建失败: %w", err)
    }
    
    vectors, err := workflow.Invoke(context.Background(), texts)
    if err != nil {
        return nil, fmt.Errorf("文本处理失败: %w", err)
    }
    
    return vectors, nil
}
```

---

## ⚙️ 高级配置和选项

### Option 配置

Embedding 支持通过 Option 在运行时传入额外配置：

```go
// WithModel - 临时切换模型
vectors, err := embedder.EmbedStrings(ctx, texts,
    embedding.WithModel("text-embedding-3-small"),
)

// WithBatchSize - 设置批处理大小（适合大量文本处理）
vectors, err := embedder.EmbedStrings(ctx, texts,
    embedding.WithBatchSize(100),
)

// 组合多个选项
vectors, err := embedder.EmbedStrings(ctx, texts,
    embedding.WithModel("text-embedding-3-large"),
    embedding.WithBatchSize(50),
)
```

### Callback 机制

回调机制允许在关键生命周期节点注入自定义逻辑：

```go
// 创建回调处理器
callbackHandler := callbacks.NewHandlerBuilder().
    OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
        fmt.Printf("📝 开始向量化 %d 个文本\n", len(input.([]string)))
        return ctx
    }).
    OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) {
        vectors := output.([][]float64)
        fmt.Printf("✅ 成功向量化 %d 个文本\n", len(vectors))
    }).
    OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) {
        fmt.Printf("❌ 向量化失败: %v\n", err)
    }).
    Build()

// 在编排中使用回调
chain := compose.NewChain[[]string, [][]float64]()
chain.AppendEmbedding(embedder, compose.WithCallbacks(callbackHandler))
```

---

## 🎓 高级用法和技巧

### 1. 📊 相似度计算工具

不同的相似度计算方法适用于不同场景：

```go
// 余弦相似度（最常用）
func cosineSimilarity(a, b []float64) float64 {
    var dotProduct, normA, normB float64
    for i := range a {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// 欧几里得距离
func euclideanDistance(a, b []float64) float64 {
    var sum float64
    for i := range a {
        diff := a[i] - b[i]
        sum += diff * diff
    }
    return math.Sqrt(sum)
}

// 曼哈顿距离
func manhattanDistance(a, b []float64) float64 {
    var sum float64
    for i := range a {
        sum += math.Abs(a[i] - b[i])
    }
    return sum
}
```

### 2. 🔄 批量处理优化

```go
func batchEmbedTexts(embedder *ark.Embedder, texts []string, batchSize int) ([][]float64, error) {
    var allVectors [][]float64
    
    for i := 0; i < len(texts); i += batchSize {
        end := i + batchSize
        if end > len(texts) {
            end = len(texts)
        }
        
        batch := texts[i:end]
        vectors, err := embedder.EmbedStrings(context.Background(), batch)
        if err != nil {
            return nil, fmt.Errorf("批次 %d-%d 向量化失败: %w", i, end-1, err)
        }
        
        allVectors = append(allVectors, vectors...)
        
        // 添加适当延迟避免过载
        time.Sleep(100 * time.Millisecond)
    }
    
    return allVectors, nil
}
```

### 3. 📈 性能监控

```go
type EmbeddingMetrics struct {
    TotalTexts       int64
    SuccessfulEmbeds int64
    FailedEmbeds     int64
    AverageEmbedTime time.Duration
    LastEmbedTime    time.Time
}

func (m *EmbeddingMetrics) RecordEmbed(textCount int, duration time.Duration, success bool) {
    m.TotalTexts += int64(textCount)
    m.LastEmbedTime = time.Now()
    
    if success {
        m.SuccessfulEmbeds++
    } else {
        m.FailedEmbeds++
    }
    
    // 更新平均时间（简化版本）
    m.AverageEmbedTime = (m.AverageEmbedTime + duration) / 2
}

func embedWithMetrics(embedder *ark.Embedder, texts []string, metrics *EmbeddingMetrics) ([][]float64, error) {
    startTime := time.Now()
    
    vectors, err := embedder.EmbedStrings(context.Background(), texts)
    
    duration := time.Since(startTime)
    metrics.RecordEmbed(len(texts), duration, err == nil)
    
    return vectors, err
}
```

### 4. 🏗️ 语义搜索引擎

```go
type SemanticSearchEngine struct {
    embedder  *ark.Embedder
    documents []Document
    vectors   [][]float64
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

## ❓ 常见问题和解决方案

### Q1: API调用失败，提示认证错误

**问题**: 运行时提示API Key无效
```
error: invalid api key
```

**解决方案**: 
```go
// ✅ 确保API Key正确配置
config := &ark.EmbeddingConfig{
    APIKey: os.Getenv("ARK_API_KEY"), // 确认环境变量设置正确
    Model:  "doubao-embedding-text-240715",
}

// 验证API Key格式
if config.APIKey == "" {
    log.Fatal("API Key 未配置")
}
```

### Q2: 向量维度不一致错误

**问题**: 相似度计算时向量维度不匹配
```go
// ❌ 错误做法：混用不同模型的向量
vectors1, _ := embedder1.EmbedStrings(ctx, texts1) // 模型A：1536维
vectors2, _ := embedder2.EmbedStrings(ctx, texts2) // 模型B：1024维
similarity := cosineSimilarity(vectors1[0], vectors2[0]) // 维度不匹配错误
```

**解决方案**:
```go
// ✅ 确保使用同一个模型/embedder实例
embedder, _ := ark.NewEmbedder(ctx, config)
vectors1, _ := embedder.EmbedStrings(ctx, texts1)
vectors2, _ := embedder.EmbedStrings(ctx, texts2)
similarity := cosineSimilarity(vectors1[0], vectors2[0]) // 维度一致
```

### Q3: 大批量文本处理超时

**问题**: 处理大量文本时请求超时
```go
// ❌ 可能导致超时的大批量请求
largeTexts := make([]string, 1000)
vectors, err := embedder.EmbedStrings(ctx, largeTexts) // 可能超时
```

**解决方案**:
```go
// ✅ 分批处理策略
func batchProcess(embedder *ark.Embedder, texts []string, batchSize int) ([][]float64, error) {
    var allVectors [][]float64
    
    for i := 0; i < len(texts); i += batchSize {
        end := i + batchSize
        if end > len(texts) {
            end = len(texts)
        }
        
        batch := texts[i:end]
        vectors, err := embedder.EmbedStrings(context.Background(), batch)
        if err != nil {
            return nil, fmt.Errorf("批次处理失败: %w", err)
        }
        
        allVectors = append(allVectors, vectors...)
        time.Sleep(100 * time.Millisecond) // 避免频率限制
    }
    
    return allVectors, nil
}
```

### Q4: 内存使用过多

**问题**: 向量缓存占用大量内存
```go
// ❌ 无限制的向量缓存
type UnlimitedCache struct {
    cache map[string][]float64 // 可能无限增长
}
```

**解决方案**:
```go
// ✅ LRU缓存策略
type LRUVectorCache struct {
    cache    map[string][]float64
    usage    []string
    maxSize  int
    mu       sync.RWMutex
}

func (lru *LRUVectorCache) Get(text string) ([]float64, bool) {
    lru.mu.Lock()
    defer lru.mu.Unlock()
    
    if vec, exists := lru.cache[text]; exists {
        // 更新使用顺序
        lru.moveToFront(text)
        return vec, true
    }
    return nil, false
}

func (lru *LRUVectorCache) Set(text string, vector []float64) {
    lru.mu.Lock()
    defer lru.mu.Unlock()
    
    if len(lru.cache) >= lru.maxSize {
        // 移除最少使用的项
        oldest := lru.usage[len(lru.usage)-1]
        delete(lru.cache, oldest)
        lru.usage = lru.usage[:len(lru.usage)-1]
    }
    
    lru.cache[text] = vector
    lru.usage = append([]string{text}, lru.usage...)
}
```

---

## 🎉 总结

Embedding 是 Eino 框架中的**核心向量化组件**，掌握它的使用对于构建高质量的 AI 应用至关重要：

### 🏆 核心优势
- 🧠 **语义理解**: 深度理解文本语义，超越关键词匹配
- ⚡ **高性能**: 支持批量处理和并发操作，适应大规模数据
- 🔍 **精准匹配**: 提供强大的相似度计算和语义搜索能力
- 🧩 **组件化**: 与 Eino 生态系统深度集成，构建完整工作流
- 🛡️ **可靠性**: 完善的错误处理和恢复机制，保证服务稳定
- 🔧 **灵活性**: 支持多种向量化策略和模型选择

### 💡 最佳实践总结
1. **合理配置**: 根据业务需求选择合适的模型和参数
2. **批量优化**: 使用适当的批量大小提升处理效率和吞吐量
3. **错误处理**: 实施完善的错误检测、分类和恢复机制
4. **性能监控**: 定期监控向量化性能、成功率和系统健康状态
5. **缓存策略**: 合理使用向量缓存减少重复计算开销
6. **编排集成**: 优先使用 Chain/Graph 编排构建自动化工作流

### 🔗 相关资源
- 📚 [官方文档](https://www.cloudwego.io/zh/docs/eino/core_modules/components/embedding_guide/)
- 🌐 [GitHub 仓库](https://github.com/cloudwego/eino)
- 🎯 [ARK 平台](https://www.volcengine.com/product/ark)

通过掌握 Embedding 组件的各种功能和最佳实践，你将能够构建出更加智能、高效和可扩展的文本处理和语义搜索系统！🚀