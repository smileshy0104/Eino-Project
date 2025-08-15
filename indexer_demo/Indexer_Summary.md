# Eino Indexer 组件核心要点总结

本文档是对 Eino 框架中 `Indexer` 组件的核心功能和使用方式的总结。

---

## 1. 核心功能

`Indexer` 组件是一个用于**存储和索引文档**的组件，它充当了 Eino 框架与**后端向量数据库**（如 Milvus, VikingDB）之间的桥梁。

**主要应用场景:**
- **构建 RAG 的知识库**: `Indexer` 是将文档（文本、元数据、向量）存入向量数据库，构建可供检索的知识库的核心工具。
- **持久化文档数据**: 将 `schema.Document` 对象及其向量表示进行持久化存储。
- **管理数据索引**: 抽象了不同向量数据库的底层实现，提供统一的存储接口。

---

## 2. 核心接口

`Indexer` 组件的核心接口同样非常简洁，主要围绕 `Store` 方法：

```go
type Indexer interface {
    Store(ctx context.Context, docs []*schema.Document, opts ...Option) (ids []string, err error)
}
```

- **`Store` 方法**: 这是该组件最核心的方法。
    - **输入**: 一个待存储的文档列表 (`[]*schema.Document`)。
    - **输出**: 一个成功存储的文档 ID 列表 (`[]string`)。

---

## 3. 核心概念：向量化策略

`Indexer` 组件支持两种主要的文本向量化策略，这通常在具体实现（如 `volc_vikingdb` 或 `milvus`）的配置中决定：

1.  **服务端向量化 (Server-Side Embedding)**
    - **流程**: 客户端只提供原始文本文档，由后端的向量数据库（如 VikingDB）**内置的 Embedding 模型**负责将文本转换为向量。
    - **优点**: 客户端逻辑简单，无需管理 Embedding 模型。
    - **示例**: `volc_vikingdb.IndexerConfig` 中的 `UseBuiltin: true`。

2.  **客户端向量化 (Client-Side Embedding)**
    - **流程**: 在客户端（代码中）先使用一个独立的 `Embedding` 组件将文本文档转换为向量，然后将**包含向量的文档**交给 `Indexer` 进行存储。
    - **优点**: 灵活性高，可以选择任意 `Embedding` 模型，并且可以在存储前对向量进行处理。
    - **示例**: `milvus.IndexerConfig` 中需要传入一个 `Embedding` 组件实例。

---

## 4. 使用方式

### 4.1. 单独使用

这是最直接的使用方式，用于将一批文档存入向量数据库。

```go
import (
    "github.com/cloudwego/eino/schema"
    "github.com/cloudwego/eino-ext/components/indexer/milvus"
)

// 1. 准备文档
doc := &schema.Document{
    ID:      "doc-001",
    Content: "这是文档的内容。",
    MetaData: map[string]interface{}{"source": "manual"},
}

// 2. 配置并创建 Indexer
// (需要提供客户端、Collection 名称、Embedding 组件等)
cfg := &milvus.IndexerConfig{ /* ... */ }
indexer, err := milvus.NewIndexer(ctx, cfg)

// 3. 调用 Store 方法
ids, err := indexer.Store(ctx, []*schema.Document{doc})
```

### 4.2. 在编排中使用 (推荐)

与 Eino 的其他组件一样，官方推荐将 `Indexer` 放入 `compose.Chain` 或 `compose.Graph` 中进行编排，构建自动化的数据处理管道（Pipeline）。

**典型流程**: `Retriever` 读取文档 -> `Transformer` 清洗文档 -> `Indexer` 存储文档。

```go
import "github.com/cloudwego/eino/compose"

// 1. 创建一个接收 []*schema.Document，输出 []string 的 Chain
chain := compose.NewChain[[]*schema.Document, []string]()

// 2. 将 indexer 附加到链中
chain.AppendIndexer(indexer)

// 3. 编译并运行
runnable, _ := chain.Compile(ctx)
ids, _ := runnable.Invoke(ctx, docs)
```

---

## 5. Option 和 Callback

### 5.1. Option (选项)

`Indexer` 组件支持通过 `Option` 在调用 `Store` 方法时传入额外参数。

- **`WithSubIndexes`**: （公共 Option）指定要操作的子索引，适用于某些支持命名空间或分区的数据库。
- **`WithEmbedding`**: （公共 Option）在调用时**临时替换**在 `Indexer` 初始化时配置的 `Embedding` 组件。这提供了极大的灵活性。

```go
// 临时使用另一个 embedder
ids, err := indexer.Store(ctx, docs,
    indexer.WithEmbedding(anotherEmbedder),
)
```

### 5.2. Callback (回调)

回调机制允许开发者在 `Indexer` 的生命周期关键点（如开始、结束、出错时）注入自定义逻辑，常用于日志记录、监控或进度更新。

- **`OnStart`**: 在 `Store` 开始执行时触发。
- **`OnEnd`**: 在成功存储所有文档后触发。
- **`OnError`**: 在发生错误时触发。

回调通常与 `compose` 编排工具结合使用，通过 `compose.WithCallbacks(handler)` 选项传入。