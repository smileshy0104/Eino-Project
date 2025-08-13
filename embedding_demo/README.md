# Eino Embedding 组件核心要点总结

本文档是对 Eino 框架中 `Embedding` 组件的核心功能和使用方式的总结。

---

## 1. 核心功能

`Embedding` 组件是一个用于将**文本转换为向量表示**的组件。它的主要作用是将文本内容映射到一个高维数学空间中，使得**语义相似的文本在空间中的距离更近**。

**主要应用场景:**
- **语义搜索**: 查找与用户查询意图最相符的文档，而不仅仅是关键字匹配。
- **文本相似度计算**: 判断两段文本在意思上的接近程度。
- **文本聚类与分类**: 将相似的文本分组或打上标签。
- **RAG (检索增强生成)**: 在 RAG 流程中，`Embedding` 是实现“检索”步骤的关键技术。

---

## 2. 核心接口

`Embedding` 组件的核心接口非常简洁：

```go
type Embedder interface {
    EmbedStrings(ctx context.Context, texts []string, opts ...Option) ([][]float64, error)
}
```

- **`EmbedStrings` 方法**: 这是该组件最核心的方法。
    - **输入**: 一个待转换的文本字符串列表 (`[]string`)。
    - **输出**: 一个向量列表 (`[][]float64`)，每个向量对应一个输入文本。向量的维度由具体模型决定。

---

## 3. 使用方式

### 3.1. 单独使用

这是最直接的使用方式，用于快速获取文本的向量表示。

```go
import "github.com/cloudwego/eino-ext/components/embedding/ark"

// 1. 初始化 Embedder
// 需要提供 API Key 和指定的模型名称
embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
    APIKey: "YOUR_API_KEY",
    Model:  "bge-large-zh",
})

// 2. 调用 EmbedStrings
texts := []string{"你好", "你好吗？"}
vectors, err := embedder.EmbedStrings(ctx, texts)

// vectors[0] 就是 "你好" 的向量表示
// vectors[1] 就是 "你好吗？" 的向量表示
```

### 3.2. 在编排中使用 (推荐)

与 `ChatTemplate` 类似，官方推荐将 `Embedding` 组件放入 `compose.Chain` 或 `compose.Graph` 中进行编排，以构建更复杂的应用（如完整的 RAG 流程）。

```go
import "github.com/cloudwego/eino/compose"

// 1. 创建一个接收 []string，输出 [][]float64 的 Chain
chain := compose.NewChain[[]string, [][]float64]()

// 2. 将 embedder 附加到链中
chain.AppendEmbedding(embedder)

// 3. 编译并运行
runnable, _ := chain.Compile(ctx)
vectors, _ := runnable.Invoke(ctx, []string{"hello", "how are you"})
```

---

## 4. Option 和 Callback

### 4.1. Option (选项)

`Embedding` 组件支持通过 `Option` 在调用时传入额外参数，最常见的公共 `Option` 是 `WithModel`。

```go
import "github.com/cloudwego/eino/components/embedding"

// 在调用时临时切换模型
vectors, err := embedder.EmbedStrings(ctx, texts,
    embedding.WithModel("text-embedding-3-small"),
)
```
*注意：具体的实现（如 `ark` 或 `openai`）可能还支持更多特有的 `Option`。*

### 4.2. Callback (回调)

回调机制允许开发者在 `Embedding` 的生命周期关键点（如开始、结束、出错时）注入自定义逻辑，常用于日志记录、监控或调试。

- **`OnStart`**: 在 `EmbedStrings` 开始执行时触发。
- **`OnEnd`**: 在成功生成所有向量后触发。
- **`OnError`**: 在发生错误时触发。

回调通常与 `compose` 编排工具结合使用，通过 `compose.WithCallbacks(handler)` 选项传入。

```go
// 1. 创建 Callback Handler
handler := &callbacksHelper.EmbeddingCallbackHandler{
    OnStart: func(...) { /* ... */ },
    OnEnd:   func(...) { /* ... */ },
}
callbackHandler := callbacksHelper.NewHandlerHelper().Embedding(handler).Handler()

// 2. 在 Invoke 时传入
vectors, _ = runnable.Invoke(ctx, texts,
    compose.WithCallbacks(callbackHandler),
)
```

这提供了一种非侵入式的方式来观察和控制组件的执行过程。