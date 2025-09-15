# 🔄 Eino Transformer 组件完全指南

本文档是对 Eino 框架中 `Transformer` 组件的核心功能和使用方式的完整总结，结合官方文档和实际项目示例。

## 🚀 快速开始

### 🛠️ 配置文件
项目使用 `config.yaml` 配置文件，也可以通过环境变量设置：
```yaml
ARK_API_KEY: "${ARK_API_KEY}"
EMBEDDER_MODEL: "doubao-embedding-text-240715"
MILVUS_ADDRESS: "localhost:19530"
MILVUS_COLLECTION: "eino_demo_collection"
```

---

## 📖 基本介绍

`Transformer` 组件是一个专门用于**文档处理和转换**的智能组件。它的主要作用是对原始文档进行预处理，如分割、过滤、格式转换等操作，为后续的向量化和检索做准备。这个组件在 AI 应用开发中扮演着**"智能文档预处理器"**的角色。

### 🎯 核心价值

在传统的文档处理中，我们只能进行简单的字符串操作。而 Transformer 组件让我们能够：

```
传统处理：简单分割 + 固定规则 + 手动处理  ❌
Transformer：语义理解 + 智能分割 + 自动化处理 + 结构保持  ✅
```

### 🚀 主要应用场景

- **📄 文档分割**: 将长文档分割成语义完整的小块，提升检索精度
- **🎭 格式转换**: 不同文档格式之间的转换和标准化处理
- **🔍 内容过滤**: 基于规则或条件过滤文档中的特定内容
- **🏗️ 结构重组**: 重新组织文档结构，优化信息展示方式
- **📊 批量处理**: 支持大规模文档集合的批量转换操作
- **🧩 组件协作**: 与其他 Eino 组件无缝集成构建完整工作流

---

## 🔧 核心接口

`Transformer` 组件提供了简洁而强大的接口设计：

### 基础接口

```go
type Transformer interface {
    Transform(ctx context.Context, src []*schema.Document, opts ...TransformerOption) ([]*schema.Document, error)
}
```

### 接口详解

#### 🔄 Transform 方法
- **功能**: 对输入文档进行转换处理
- **输入**:
    - `ctx`: 上下文对象，用于控制超时、取消等
    - `src`: 待转换的原始文档列表 (`[]*schema.Document`)
    - `opts`: 可选的转换参数
- **输出**:
    - `[]*schema.Document`: 转换后的文档列表
    - `error`: 转换过程中的错误信息

---

## 📨 Document 结构体

`Document` 是转换的基本数据结构，支持丰富的文档类型：

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

- **🔑 ID**: 文档的唯一标识符，转换后可能会生成新的 ID（如分割后的块）
- **📄 Content**: 文档主要文本内容，是转换操作的核心目标
- **🏷️ MetaData**: 结构化元数据，在转换过程中可能会被保留、修改或扩展：
  - 原始文档的来源信息
  - 转换操作的相关信息（如分割依据、块序号等）
  - 文档的结构信息（如标题层级、章节信息等）
  - 其他自定义转换参数

---

## 🎯 转换策略

Transformer 支持多种文档转换策略：

### 1. 📑 Markdown 标题分割 (Header Splitting)

**工作原理**:
```
长 Markdown 文档 → 按标题层级分割 → 多个语义完整的文档块
```

**特点**:
- 保持文档的逻辑结构完整性
- 支持多级标题的灵活配置
- 自动生成层级化的元数据信息

**适用场景**:
- 技术文档、博客文章的智能分割
- 保持语义完整性的长文档处理
- 需要保留文档结构信息的应用

### 2. 📝 文本分割 (Text Splitting)

**工作原理**:
```
长文本内容 → 按字符数/句子/段落分割 → 固定大小的文本块
```

**特点**:
- 支持多种分割策略（字符数、句子、段落）
- 可配置重叠区域防止语义断裂
- 适合处理纯文本内容

**适用场景**:
- 大规模文本数据的向量化预处理
- 对分割粒度有特定要求的应用
- 处理无明显结构的长文本

### 3. 🔍 文档过滤 (Document Filtering)

**工作原理**:
```
文档集合 → 应用过滤规则 → 符合条件的文档子集
```

**特点**:
- 支持基于内容、元数据的复杂过滤
- 可组合多个过滤条件
- 保持原始文档结构不变

**适用场景**:
- 数据清洗和预处理
- 根据业务规则筛选相关文档
- 去除重复或无效内容

---

## 🏗️ 创建和使用 Transformer

### 基础使用流程

```go
import (
    "github.com/cloudwego/eino/schema"
    "github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
)

// 1️⃣ 准备原始文档
originalDoc := &schema.Document{
    ID: "eino-intro-doc",
    Content: `
# Eino 框架介绍
Eino 是一个先进的大模型应用开发框架。

## 核心组件
Eino 提供了多种核心组件，包括 Model, Retriever, Indexer, 和 Transformer。

## Transformer 详解
Transformer 组件负责文档的预处理。它可以将长文档分割成小块，过滤无关信息。

## 快速开始
要开始使用 Eino，请参考我们的官方文档和示例代码。`,
    MetaData: map[string]interface{}{
        "source": "official-docs",
        "author": "eino-team",
    },
}

// 2️⃣ 初始化 Transformer（以 Markdown Header Splitter 为例）
splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
    Headers: map[string]string{
        "##": "Header 2",  // 使用二级标题作为分割点
    },
})
if err != nil {
    log.Fatal("Transformer 初始化失败:", err)
}

// 3️⃣ 执行转换
chunks, err := splitter.Transform(ctx, []*schema.Document{originalDoc})
if err != nil {
    log.Fatal("文档转换失败:", err)
}

fmt.Printf("转换完成，原文档被分割成 %d 个块\n", len(chunks))

// 4️⃣ 处理转换结果
for i, chunk := range chunks {
    fmt.Printf("\n--- 文档块 %d ---\n", i+1)
    fmt.Printf("ID: %s\n", chunk.ID)
    fmt.Printf("内容: %s\n", chunk.Content)
    fmt.Printf("元数据: %v\n", chunk.MetaData)
}
```

### 🎯 高级配置示例

#### Markdown Header Splitter 详细配置
```go
// 支持多级标题的复杂分割策略
splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
    Headers: map[string]string{
        "#":   "Header 1",  // 一级标题
        "##":  "Header 2",  // 二级标题
        "###": "Header 3",  // 三级标题
    },
    // 可选: 添加其他配置参数
    PreserveStructure: true,  // 保持层级结构
    IncludeHeaders:   true,   // 在块中包含标题
})
```

#### 复杂文档转换示例
```go
complexDoc := &schema.Document{
    ID: "technical_manual_001",
    Content: `# AI 应用开发完全手册

本手册将指导您完成从零开始的 AI 应用开发全流程。

## 第一章：环境准备
在开始开发之前，需要准备必要的开发环境。

### 1.1 硬件要求
- CPU: Intel i7 或 AMD Ryzen 7 以上
- 内存: 16GB 以上 RAM
- 存储: 500GB 以上 SSD

### 1.2 软件依赖
- Python 3.9+
- Docker 20.10+
- Git 2.30+

## 第二章：框架选择
选择合适的 AI 开发框架是成功的关键。

### 2.1 Eino 框架优势
Eino 提供了完整的 RAG 开发工具链。

### 2.2 竞品对比
与其他框架相比，Eino 具有以下优势...`,
    MetaData: map[string]interface{}{
        "document_type": "technical_manual",
        "version":      "1.0.0",
        "language":     "zh-CN",
        "chapters":     2,
        "sections":     4,
        "author":       "AI 开发团队",
        "create_date":  "2024-09-14",
        "tags":         []string{"AI", "开发手册", "Eino", "RAG"},
    },
}

// 使用二级标题进行章节分割
chapterChunks, err := splitter.Transform(ctx, []*schema.Document{complexDoc})
if err != nil {
    log.Fatal("章节分割失败:", err)
}

fmt.Printf("技术手册被分割成 %d 个章节\n", len(chapterChunks))
```

---

## 🚀 编排集成最佳实践

虽然可以直接使用 Transformer，但**官方强烈推荐**将其集成到编排工作流中，与其他组件协同工作。

### 🔗 Chain 编排模式

Chain 是最常用的编排方式，适合线性处理流程：

```go
import "github.com/cloudwego/eino/compose"

// 1️⃣ 创建 Chain - 声明输入输出类型
chain := compose.NewChain[[]*schema.Document, []*schema.Document]()

// 2️⃣ 添加组件 - 按处理顺序添加
chain.AppendTransformer(transformer)

// 3️⃣ 编译执行
runnable, err := chain.Compile(ctx)
if err != nil {
    log.Fatalf("链编译失败: %v", err)
}

// 4️⃣ 运行工作流
processedDocs, err := runnable.Invoke(ctx, originalDocs)
```

### 🔄 完整文档处理工作流

```go
func createDocumentProcessingWorkflow() (*compose.Runnable, error) {
    ctx := context.Background()

    // 🔧 初始化组件
    transformer, err := initTransformer(ctx)
    if err != nil {
        return nil, err
    }

    embedder, err := initEmbedder(ctx)
    if err != nil {
        return nil, err
    }

    indexer, err := initIndexer(ctx, embedder)
    if err != nil {
        return nil, err
    }

    // 🔗 构建处理链
    // 原始文档 → 文档分割 → 向量化 → 存储索引 → 返回ID列表
    chain := compose.NewChain[[]*schema.Document, []string]()
    chain.AppendTransformer(transformer).
          AppendIndexer(indexer)

    // ⚙️ 编译成可运行实例
    return chain.Compile(ctx)
}

// 使用示例
func processDocumentsWorkflow(docs []*schema.Document) ([]string, error) {
    workflow, err := createDocumentProcessingWorkflow()
    if err != nil {
        return nil, fmt.Errorf("工作流创建失败: %w", err)
    }

    storedIDs, err := workflow.Invoke(context.Background(), docs)
    if err != nil {
        return nil, fmt.Errorf("文档处理失败: %w", err)
    }

    return storedIDs, nil
}
```

---

## ⚙️ 高级配置和选项

### Option 配置

Transformer 支持通过 Option 在运行时传入额外配置：

```go
// WithMaxChunkSize - 设置最大块大小
chunks, err := transformer.Transform(ctx, documents,
    transformer.WithMaxChunkSize(2000),
)

// WithOverlapSize - 设置重叠区域大小
chunks, err := transformer.Transform(ctx, documents,
    transformer.WithOverlapSize(200),
)

// WithPreserveMetadata - 保留原始元数据
chunks, err := transformer.Transform(ctx, documents,
    transformer.WithPreserveMetadata(true),
)
```

### Callback 机制

回调机制允许在关键生命周期节点注入自定义逻辑：

```go
// 创建回调处理器
callbackHandler := callbacks.NewHandlerBuilder().
    OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
        docs := input.([]*schema.Document)
        fmt.Printf("🔄 开始转换 %d 个文档\n", len(docs))
        return ctx
    }).
    OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) {
        chunks := output.([]*schema.Document)
        fmt.Printf("✅ 转换完成，生成 %d 个文档块\n", len(chunks))
    }).
    OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) {
        fmt.Printf("❌ 转换失败: %v\n", err)
    }).
    Build()

// 在编排中使用回调
chain := compose.NewChain[[]*schema.Document, []*schema.Document]()
chain.AppendTransformer(transformer, compose.WithCallbacks(callbackHandler))
```

---

## 🎓 高级用法和技巧

### 1. 📊 动态转换策略

根据文档类型动态选择转换策略：

```go
type TransformerManager struct {
    strategies map[string]Transformer
}

func (tm *TransformerManager) GetTransformer(docType string) (Transformer, error) {
    if transformer, exists := tm.strategies[docType]; exists {
        return transformer, nil
    }
    return nil, fmt.Errorf("未找到文档类型 %s 的转换策略", docType)
}

func (tm *TransformerManager) TransformByType(ctx context.Context, docs []*schema.Document) ([]*schema.Document, error) {
    var allChunks []*schema.Document

    // 按文档类型分组
    docGroups := make(map[string][]*schema.Document)
    for _, doc := range docs {
        docType := getDocumentType(doc)
        docGroups[docType] = append(docGroups[docType], doc)
    }

    // 对每个类型的文档应用相应的转换策略
    for docType, groupDocs := range docGroups {
        transformer, err := tm.GetTransformer(docType)
        if err != nil {
            return nil, err
        }

        chunks, err := transformer.Transform(ctx, groupDocs)
        if err != nil {
            return nil, fmt.Errorf("转换文档类型 %s 失败: %w", docType, err)
        }

        allChunks = append(allChunks, chunks...)
    }

    return allChunks, nil
}
```

### 2. 🔄 批量转换优化

```go
func batchTransformDocuments(transformer Transformer, docs []*schema.Document, batchSize int) ([]*schema.Document, error) {
    var allChunks []*schema.Document

    for i := 0; i < len(docs); i += batchSize {
        end := i + batchSize
        if end > len(docs) {
            end = len(docs)
        }

        batch := docs[i:end]
        chunks, err := transformer.Transform(context.Background(), batch)
        if err != nil {
            return nil, fmt.Errorf("批次 %d-%d 转换失败: %w", i, end-1, err)
        }

        allChunks = append(allChunks, chunks...)

        // 添加适当延迟避免过载
        time.Sleep(50 * time.Millisecond)
        fmt.Printf("完成批次 %d-%d，累计生成 %d 个文档块\n", i, end-1, len(allChunks))
    }

    return allChunks, nil
}
```

### 3. 📈 转换质量监控

```go
type TransformMetrics struct {
    TotalInputDocs     int64
    TotalOutputChunks  int64
    AverageChunksPerDoc float64
    AverageTransformTime time.Duration
    SuccessfulTransforms int64
    FailedTransforms     int64
}

func (m *TransformMetrics) RecordTransform(inputCount, outputCount int, duration time.Duration, success bool) {
    m.TotalInputDocs += int64(inputCount)
    m.TotalOutputChunks += int64(outputCount)

    if success {
        m.SuccessfulTransforms++
    } else {
        m.FailedTransforms++
    }

    // 更新平均值
    if m.TotalInputDocs > 0 {
        m.AverageChunksPerDoc = float64(m.TotalOutputChunks) / float64(m.TotalInputDocs)
    }
    m.AverageTransformTime = (m.AverageTransformTime + duration) / 2
}

func transformWithMetrics(transformer Transformer, docs []*schema.Document, metrics *TransformMetrics) ([]*schema.Document, error) {
    startTime := time.Now()

    chunks, err := transformer.Transform(context.Background(), docs)

    duration := time.Since(startTime)
    metrics.RecordTransform(len(docs), len(chunks), duration, err == nil)

    return chunks, err
}
```

### 4. 🎯 智能文档分析

```go
type DocumentAnalyzer struct {
    transformer Transformer
}

func (da *DocumentAnalyzer) AnalyzeAndTransform(ctx context.Context, doc *schema.Document) (*TransformResult, error) {
    // 分析文档特征
    analysis := da.analyzeDocument(doc)

    // 基于分析结果选择转换参数
    opts := da.selectTransformOptions(analysis)

    // 执行转换
    chunks, err := da.transformer.Transform(ctx, []*schema.Document{doc}, opts...)
    if err != nil {
        return nil, err
    }

    return &TransformResult{
        OriginalDoc: doc,
        Chunks:      chunks,
        Analysis:    analysis,
        ChunkCount:  len(chunks),
    }, nil
}

type DocumentAnalysis struct {
    WordCount       int
    SentenceCount   int
    ParagraphCount  int
    HeaderCount     int
    ComplexityScore float64
    Language        string
    DocumentType    string
}

type TransformResult struct {
    OriginalDoc *schema.Document
    Chunks      []*schema.Document
    Analysis    DocumentAnalysis
    ChunkCount  int
}
```

---

## ❓ 常见问题和解决方案

### Q1: 文档分割后 ID 冲突问题

**问题**: 分割后的文档块产生重复或无效的 ID
```go
// ❌ 可能产生 ID 冲突
originalDoc := &schema.Document{
    ID: "doc_001",
    Content: "很长的文档内容...",
}
chunks, _ := transformer.Transform(ctx, []*schema.Document{originalDoc})
// 分割后的块可能有相同或空的 ID
```

**解决方案**:
```go
// ✅ 确保分割后的 ID 唯一性
func ensureUniqueIDs(chunks []*schema.Document, originalID string) {
    for i, chunk := range chunks {
        if chunk.ID == "" || chunk.ID == originalID {
            chunk.ID = fmt.Sprintf("%s_chunk_%d", originalID, i)
        }
    }
}

// 使用示例
chunks, err := transformer.Transform(ctx, []*schema.Document{originalDoc})
if err != nil {
    return err
}
ensureUniqueIDs(chunks, originalDoc.ID)
```

### Q2: 大文档分割后块过多问题

**问题**: 长文档分割后产生过多的小块，影响后续处理效率
```go
// ❌ 可能产生过多的小块
splitter, _ := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
    Headers: map[string]string{
        "#":     "Header 1",
        "##":    "Header 2",
        "###":   "Header 3",
        "####":  "Header 4",  // 过细的分割
        "#####": "Header 5",  // 可能产生很多小块
    },
})
```

**解决方案**:
```go
// ✅ 合理配置分割策略
func createOptimalSplitter(ctx context.Context, maxChunks int) (Transformer, error) {
    // 根据目标块数选择合适的分割层级
    config := &markdown.HeaderConfig{
        Headers: map[string]string{
            "##": "Header 2",  // 主要使用二级标题
        },
    }

    // 如果需要更细粒度的分割，可以添加三级标题
    if maxChunks > 10 {
        config.Headers["###"] = "Header 3"
    }

    return markdown.NewHeaderSplitter(ctx, config)
}

// 后处理：合并过小的块
func mergeSmallChunks(chunks []*schema.Document, minSize int) []*schema.Document {
    var mergedChunks []*schema.Document
    var currentChunk *schema.Document

    for _, chunk := range chunks {
        if len(chunk.Content) < minSize && currentChunk != nil {
            // 合并到当前块
            currentChunk.Content += "\n" + chunk.Content
        } else {
            if currentChunk != nil {
                mergedChunks = append(mergedChunks, currentChunk)
            }
            currentChunk = &schema.Document{
                ID:       chunk.ID,
                Content:  chunk.Content,
                MetaData: chunk.MetaData,
            }
        }
    }

    if currentChunk != nil {
        mergedChunks = append(mergedChunks, currentChunk)
    }

    return mergedChunks
}
```

### Q3: 元数据信息丢失问题

**问题**: 转换过程中丢失重要的元数据信息
```go
// ❌ 转换后元数据可能丢失或不完整
originalDoc := &schema.Document{
    ID: "important_doc",
    Content: "...",
    MetaData: map[string]interface{}{
        "author":      "专家团队",
        "create_date": "2024-09-14",
        "importance":  "high",
        "category":    "技术文档",
    },
}
```

**解决方案**:
```go
// ✅ 保持和增强元数据
func enhanceChunkMetadata(originalDoc *schema.Document, chunks []*schema.Document) {
    for i, chunk := range chunks {
        // 复制原始元数据
        newMetaData := make(map[string]interface{})
        for k, v := range originalDoc.MetaData {
            newMetaData[k] = v
        }

        // 添加分割相关信息
        newMetaData["original_id"] = originalDoc.ID
        newMetaData["chunk_index"] = i
        newMetaData["total_chunks"] = len(chunks)
        newMetaData["transform_time"] = time.Now().Format(time.RFC3339)
        newMetaData["chunk_word_count"] = len(strings.Fields(chunk.Content))

        chunk.MetaData = newMetaData
    }
}
```

### Q4: 性能优化问题

**问题**: 大批量文档转换时性能不佳
```go
// ✅ 并行处理优化
func parallelTransform(transformer Transformer, docs []*schema.Document, workers int) ([]*schema.Document, error) {
    docsChan := make(chan *schema.Document, len(docs))
    resultsChan := make(chan TransformResult, len(docs))

    // 启动工作协程
    for i := 0; i < workers; i++ {
        go func() {
            for doc := range docsChan {
                chunks, err := transformer.Transform(context.Background(), []*schema.Document{doc})
                resultsChan <- TransformResult{
                    Chunks: chunks,
                    Error:  err,
                    DocID:  doc.ID,
                }
            }
        }()
    }

    // 发送任务
    for _, doc := range docs {
        docsChan <- doc
    }
    close(docsChan)

    // 收集结果
    var allChunks []*schema.Document
    for i := 0; i < len(docs); i++ {
        result := <-resultsChan
        if result.Error != nil {
            return nil, fmt.Errorf("转换文档 %s 失败: %w", result.DocID, result.Error)
        }
        allChunks = append(allChunks, result.Chunks...)
    }

    return allChunks, nil
}

type TransformResult struct {
    Chunks []*schema.Document
    Error  error
    DocID  string
}
```

---

## 🎉 总结

Transformer 是 Eino 框架中的**核心预处理组件**，掌握它的使用对于构建高质量的 AI 应用至关重要：

### 🏆 核心优势
- 🔄 **智能转换**: 支持多种文档转换策略，保持语义完整性
- ⚡ **高性能**: 支持批量处理和并发操作，适应大规模数据
- 🎯 **灵活配置**: 丰富的配置选项和回调机制，满足各种需求
- 🧩 **组件化**: 与 Eino 生态系统深度集成，构建完整工作流
- 🛡️ **可靠性**: 完善的错误处理和恢复机制，保证处理质量
- 🔧 **扩展性**: 支持自定义转换策略和处理逻辑

### 💡 最佳实践总结
1. **策略选择**: 根据文档类型和业务需求选择合适的转换策略
2. **参数调优**: 合理配置分割大小、重叠区域等关键参数
3. **元数据管理**: 妥善处理和保持文档的元数据信息
4. **性能优化**: 使用批量处理和并行处理提升转换效率
5. **质量监控**: 定期监控转换质量和系统性能指标
6. **编排集成**: 优先使用 Chain/Graph 编排构建自动化工作流

### 🔗 相关资源
- 📚 [官方文档](https://www.cloudwego.io/zh/docs/eino/core_modules/components/document_transformer_guide/)
- 🌐 [GitHub 仓库](https://github.com/cloudwego/eino)
- 🔄 [Transformer 实现示例](https://github.com/cloudwego/eino-ext/tree/main/components/document/transformer)

通过掌握 Transformer 组件的各种功能和最佳实践，你将能够构建出更加智能、高效和可扩展的文档处理和转换系统！🚀